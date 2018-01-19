function insert(parentId, d) {
    var parent = document.getElementById(parentId);
    parent.insertBefore(d, parent.firstChild);
}

function pre(s) {
    var d = document.createElement("pre");
    d.innerHTML = s;
    return d;
}

var handlers = {
    routing: {
	pattern: {routing: "?r"},
	f: function(bs) {
	    var msg = JSON.stringify({routing: bs["?r"]});
	    insert("routing", pre(msg));
	}
    },

    did: {
	pattern: {did: "?did"},
	f: function(bs) {
	    var msg = JSON.stringify({did: bs["?did"]});
	    insert("did", pre(msg));
	}
    }
};

window.addEventListener("load", function(evt) {

    var loc = window.location, ws_url;
    if (loc.protocol === "https:") {
	ws_url = "wss:";
    } else {
	ws_url = "ws:";
    }
    
    ws_url += "//" + loc.host + "/ws/api";
    console.log("ws_url", ws_url);

    var print = function(msg) {
	insert("log", pre(msg));
    };

    ws = new WebSocket(ws_url);
    
    ws.onopen = function(evt) {
        print("open");
    };
    
    ws.onclose = function(evt) {
        print("close");
        ws = null;
    };
    
    ws.onmessage = function(evt) {
        print("received " + evt.data);
	var msg = JSON.parse(evt.data);
	if (msg) {
	    for (var id in handlers) {
		var handler = handlers[id];
		var bss = match(null, handler.pattern, msg, {});
		if (bss) {
		    for (var i = 0; i < bss.length; i++) {
			handler.f(bss[i]);
		    }
		}
	    }
	}
    };
    
    ws.onerror = function(evt) {
        print("error: " + evt.data);
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
	    print("No WebSocket connection");
	    return
        }
        print("sending " + input.value);
        ws.send(input.value);
        return false;
    };
    
    document.getElementById("demo").onclick = function(evt) {
        if (!ws) {
	    print("No WebSocket connection");
	    return
        }
	var msg = '{"cop":{"add":{"m":{"id":"doubler","spec":{"name":"double"}}}}}';
        ws.send(msg);
	msg = '{"cop":{"process":{"message":{"to":"doubler","double":1}}}}';
        ws.send(msg);
	msg = '{"cop":{"process":{"message":{"to":"timers","makeTimer":{"in":"1s","id":"1","message":{"to":"doubler","double":100}}}}}}';
        ws.send(msg);
        return false;
    };
    
    return false;
});

