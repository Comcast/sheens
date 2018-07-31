// Demo macro to turn on a light.
//
// Arguments: deviceid, to (target).
register("turnOn", function(x, path, root) {

    var msg = {
	state: "on",
	device: x.deviceid
    };

    var node = {
	action: {
	    interpreter: "ecmascript",
	    source: "_.out(" + JSON.stringify(msg) + ");"
	},
	branching: {
	    branches: [
		{
		    pattern: {likes: "queso"},
		    target: x.to
		}
	    ]
	}
    };
	
    setIn(node, path, root);
});

