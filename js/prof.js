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


var Times = function() {
    var totals = {}
    var clocks = {};
    var enabled = false;

    return {
	enable: function() {
	    enabled = true;
	},
	disable: function() {
	    enabled = false;
	},
	tick: function(what) {
	    if (!enabled) return;
	    clocks[what] = new Date().getTime();
	},
	tock: function(what) {
	    if (!enabled) return;
	    var elapsed = new Date().getTime() - clocks[what];
	    var entry = totals[what];
	    if (!entry) {
		entry = {ms: 0, n: 0};
		totals[what] = entry;
	    }
	    entry.ms += elapsed;
	    entry.n++;
	},
	summary: function() {
	    return totals;
	},
	reset: function() {
	    totals = {};
	},
    };
}();

