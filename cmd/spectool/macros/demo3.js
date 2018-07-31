// Demo macro to create a node.
//
// Arguments: message (to emit), to (target).
register("demo3", function(x, path, root) {
    var node = {
	action: {
	    interpreter: "ecmascript",
	    source: "_.out(" + JSON.stringify(x.message) + ");"
	},
	branching: {
	    branches: {
		pattern: {likes: "queso"},
		target: x.to
	    }
	}
    };
	
    setIn(node, path, root);
});

