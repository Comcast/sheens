// A quick example of running a machine.  The variable 'spec' should
// contain the machine's specification.  The Makefile will generate
// 'double.js', which defines 'spec.
//
// Usage: ./demo double.js demo.js

// The machine's initial state.
var m = {
    bs: {count: 0},
    node: "start"
};

// The messages that we'll process.
var msgs = [{double:1}, {double:10}, {double:100}];

var ctx = null;

// Process each message, and update the machine's state as we go.
for (var i = 0; i < msgs.length; i++) {
    var msg = msgs[i];
    
    print("state", i, JSON.stringify(m));
    print("stepping", i, JSON.stringify(msg));
    
    var stepped = walk(ctx, spec, m, msg);
    print("stepped", i, JSON.stringify(stepped));

    var emitted = stepped.emitted;
    for (var j = 0; j < emitted.length; j++) {
	print(i, j, "emitted", JSON.stringify(emitted[j]));
    }

    m.bs = stepped.to.bs;
    m.node = stepped.to.node;
}

true;
