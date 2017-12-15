// isCurrent takes a time specification and returns a boolean to
// indicate that the current time is within the specification.
//
// The specification currently supports three properties: daysOfWeek,
// startTime, and stopTime.
//
// daysOfWeek is an array of ints in the range [0,6], where 0
// represents Sunday.
//
// startTime and stopTime are strings in the format "HH:MM".
//
// This function does no error-checking.
function isCurrent(spec) {

    if (!spec) {
	return true;
    }

    var d = new Date();

    var pad = function(n) {
	// If we're given a single-digit number, zero-pad it.
	var s = n.toString();
	if (1 == s.length) {
	    s = "0" + s;
	}
	return s;
    }

    var padt = function(t) {
	// If we're given 1:23, return 01:23.
	if (4 == t.length) {
	    t = "0" + t;
	}
	return t;
    }

    if (spec.daysOfWeek) {
	var rightDay = false;
	var day = d.getDay();
	for (var i = 0; i < spec.daysOfWeek; i++) {
	    var want = spec.daysOfWeek[i];
	    if (day == want) {
		rightDay = true;
		break;
	    }
	}
	if (!rightDay) {
	    return false;
	}
    }

    if (spec.startTime && spec.stopTime) {
	var t = pad(d.getHours()) + ":" + pad(d.getMinutes());
	var start = padt(spec.startTime);
	var stop = padt(spec.stopTime);
	return start <= t && t <= stop;
    }

    return true;
}
