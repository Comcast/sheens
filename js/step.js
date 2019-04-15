/* Copyright 2018-2019 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


// Step attempts to advance to the next state.
//
// Returns {to: STATE, consumed: BOOL, emitted: MESSAGES}.
//
// STATE will be null when there was no transition.  'Consumed'
// reports whether the message was consumed during the transition.
// MESSAGES are the zero or more messages emitted by the action.
function step(ctx,spec,state,message) {
    Times.tick("step");
    try {
	var current = state.bs;
	var bs = current;
	var emitted = [];

	// The following strings are interpretered as aliases for our only
	// actual interpreter, which is probably close to Ecmascript 5.1.
	// "goja" is in this list for backwards compatability due to
	// vestiges of github.com/Comcast/sheens history.
	var interpreterAliases = ["ecmascript", "ecmascript-5.1", "goja"];

	var node = spec.nodes[state.node];
	if (!node) {
	    throw {error: "node not found", node: state.node};
	}

	var action = node.action;
	if (action) {
	    if (interpreterAliases.indexOf(action.interpreter) < 0) {
		throw {error: "bad interpreter", interpreter: action.interpreter};
	    }
	    var evaled = sandboxedAction(ctx, bs, action.source);
	    bs = evaled.bs;
	    emitted = evaled.emitted;
	}

	var branching = node.branching;
	if (!branching || !branching.branches) {
	    return null;
	}
	var against = bs;
	var consuming = false;
	if (branching.type == "message") {
	    if (!message) {
		return null;
	    }
	    consuming = true;
	    against = message;
	}
	var branches = branching.branches;
	for (var i = 0; i < branches.length; i++) {
	    var branch = branches[i];
	    var pattern = branch.pattern;
	    if (pattern) {
		if (spec.parsepatterns || spec.patternsyntax == "json") {
		    pattern = JSON.parse(pattern);
		}
		var bss = match(ctx, pattern, against, bs);
		if (!bss || bss.length == 0) {
		    continue;
		}
		if (1 < bss.length) {
		    throw {error: "too many sets of bindings", bss: bss};
		}
		bs = bss[0];
	    }
	    if (branch.guard) {
		if (interpreterAliases.indexOf(branch.guard.interpreter) < 0) {
		    throw {error: "bad guard interpreter", interpreter: branch.guard.interpreter};
		}
		var evaled = sandboxedAction(ctx, bs, branch.guard.source);
		if (!evaled.bs) {
		    continue;
		}
		bs = evaled.bs;
		// Check that we didn't emit any messages ...
	    }
	    return {to: {node: branch.target, bs: bs}, consumed: consuming, emitted: emitted};
	}

	return null;
    } finally {
	Times.tock("step");
    }
}

// walk advances from the given state as far as possible.
//
// Returns {to: STATE, consumed: BOOL, emitted: MESSAGES}.
//
// For an description of the returned value, see doc for 'step'.
function walk(ctx,spec,state,message) {

    var maxSteps = 32;
    if (ctx && ctx.MaxSteps) {
	maxSteps = ctx.MaxSteps;
    }
    
    // ToDo: Get limit from ctx.

    if (!state) {
	state = {node: "start", bs: {}};
    }

    var emitted = [];
    var consumed = false;

    var stepped = {to: state, consumed: false};

    for (var i = 0; i <= maxSteps; i++) {
	if (i == maxSteps) {
	    stepped.stoppedBecause = "limited";
	    break;
	}
	// print("stepping from ", i, JSON.stringify(stepped.to), JSON.stringify(message));
	var maybe = step(ctx, spec, stepped.to, message);
	// print("         to   ", i, JSON.stringify(maybe));

	if (!maybe) {
	    // We went nowhere.  Stop.
	    break;
	}

	if (maybe.consumed) {
	    // We consumed the pending message; don't use it again.
	    message = null;
	    consumed = true;
	}

	stepped = maybe;

	if (stepped && 0 < stepped.emitted.length) {
	    // Accumulated emitted messages.
	    emitted = emitted.concat(stepped.emitted);
	}
    }

    stepped.emitted = emitted;
    stepped.consumed = consumed;

    return stepped;
}

