// macros is our map from macro names (strings) to macros (functions).
var macros = {};

// register updates our macros map;
//
// macro sound be a function(x, path, root).
function register(mname, macro) {
    macros[mname] = macro;
}

// setIn updates 'at' at 'path' with the value 'x'.
function setIn(x, path, at) {
    var p = path.shift();
    if (path.length == 0) {
	at[p] = x;
    } else {
	return setIn(x, path, at[p]);
    }
}

// handle is a utility function do deal with a macro invocation at x.
function handle(x, path, root) {
    if (!x) {
	return;
    }
    var mname = x.macro;
    if (mname) {
	var macro = macros[mname];
	if (!macro) {
	    throw "unknown macro: " + mname;
	}
	macro(x, path, root);
	return true;
    } else {
	apply(x, path, root);
	return false;
    }
}

// apply recursively expands macros.  Processing is depth-first.  Any
// macro expansion immediately triggers reprocessing.
function apply(o, path, root) {
    var mod = false;

    switch (typeof(o)) {
    case "object":
	if (Array.isArray(o)) {
	    for (var i = 0; i < o.length; i++) {
		mod = handle(o[i], path.concat([i]), root);
		if (mod) {
		    break;
		}
	    }
	} else {
	    for (var p in o) {
		mod = handle(o[p], path.concat([p]), root);
		if (mod) {
		    break;
		}
	    }
	}
    }

    if (mod) {
	apply(o, path, root);
    }
}

function expand(root) {
    apply(root, [], root);
    return root;
}

