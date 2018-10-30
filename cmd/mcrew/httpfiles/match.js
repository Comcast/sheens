// A Javascript implementation of Sheens pattern matching
//
// Status: Frequently compiles.
//
// Origin: https://github.com/Comcast/littlesheens/blob/master/js/match.js

// function(CTX,P,M,BS), where CTX is an unused context, P is a
// pattern, M is a message, and BS are input bindings.
//
// Returns null or a set of sets of bindings.
var match = function() {

    var isVar = function(s) {
	return typeof s == 'string' && s.charAt(0) == '?';
    };

    var copyMap = function(m) {
	var acc = {};
	for (var p in m) {
	    acc[p] = m[p];
	}
	return acc;
    };

    var copyArray = function(xs) {
	return xs.slice();
    };

    var extend = function(bs, b, v) {
	var acc = copyMap(bs);
	acc[b] = v;
	return acc;
    };

    var match;

    var matchWithBindings = function(ctx, bss, v, mv) {
	var acc = [];
	for (var i = 0; i < bss.length; i++) {
	    var bs = bss[i];
	    acc = acc.concat(match(ctx, v, mv, bs))
	}
	return acc;
    };

    var arraycatMatch = function(ctx, bss, p, m) {
	if (p.length == 0) {
	    return bss;
	}

	var px = p[0];
	var acc = [];
	for (var i = 0; i < bss.length; i++) {
	    var bs = bss[i];
	    for (var j = 0; j < m.length; j++) {
		var mx = m[j];
		var bss_ = match(ctx, px, mx, bs);
		if (bss_.length == 0) {
		    continue;
		}
		var m_ = copyArray(m);
		m_.splice(j, 1);
		bss_ = arraycatMatch(ctx, bss_, m_.slice(2), m_);
		for (var k = 0; k < bss_.length; k++) {
		    acc.push(bss_[k]);
		}
	    }
	}

	return acc;
    };

    var mapcatMatch = function(ctx, bss, p, m) {
	for (var k in p) {
	    var v = p[k];
	    var mv = m[k];
	    if (mv === undefined) {
		return [];
	    }
	    var acc = matchWithBindings(ctx, bss, v, mv)
	    if (acc.length == 0) {
		return [];
	    }
	    bss = acc;
	}
	return bss;
    };

    match = function(ctx,p,m,bs) {
	if (!bs) {
	    bs = [];
	}
	if (isVar(p)) {
	    var binding = bs[p];
	    if (binding) {
		return match(ctx, binding, m, bs);
	    } else {
		return [extend(bs, p, m)];
	    }
	} else {
	    switch (typeof p) {
	    case 'object':
		if (Array.isArray(p)) {
		    if (Array.isArray(m)) {
			return arraycatMatch(ctx, [bs], p, m);
		    } else {
			return [];
		    }
		}
		switch (typeof m) {
		case 'object':
		    if (null === p) {
			return [];
		    }
		    if (p.length == 0) {
			return [bs];
		    }
		    return mapcatMatch(ctx, [bs], p, m);
		default:
		    return [];
		}
	    default:
		if (p == m) {
		    return [bs];
		}
		return [];
	    }
	}
    };

    return match;
}();

