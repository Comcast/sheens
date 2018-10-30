
// Render a spec and enable live state change displays.
//
// Returns a function that takes state as an argument and updates the
// display.
function renderMachine(mid, spec) {

    console.log("renderMachine", mid, spec)

    var divid = "m_" + mid;
    
    var div = d3.select("#graph")
	.append("div")
	.attr("id", divid);
    if (mid) {
	div.append("div")
	    .classed("machineId", true)
	    .text(mid);
    }
    var gdivid = divid + "_graph";
    div.append("div")
	.classed("graph", true)
	.attr("id", gdivid);

    var elements = [];
    if (spec && spec.nodes) {
	for (var nodeName in spec.nodes) {
	    elements.push({data: {id: nodeName, link: "#" + nodeName}});
	    var node = spec.nodes[nodeName];
	    if (node.branching && node.branching.branches) {
		var branches = node.branching.branches;
		for (var i = 0; i < branches.length; i++) {
		    var branch = branches[i];
		    if (branch.target) {
			elements.push({data:
				       {id: nodeName + "_" + i,
					source: nodeName,
					target: branch.target}});
		    }
		}
	    }
	}
    }

    console.log("elements", elements);
    
    var cy = cytoscape({
	container: document.getElementById(gdivid),
	elements: elements,
	style: [
	    {
		selector: 'node',
		style: {
		    'content': 'data(label)',
		    'background-color': '#666',
		    'label': 'data(id)'
		}
	    },

	    {
		selector: 'edge',
		style: {
		    'curve-style': 'bezier',
		    'target-arrow-shape': 'triangle',
		    'width': 1,
		    'line-color': 'blue',
		    'target-arrow-color': 'orange',
		    'label': 'data(label)',
		}
	    }
	],

	layout: {
	    name: 'breadthfirst',
	    directed: true,
	    rows: 1
	}

    });

    cy.edges().on("tap", function(){ alert(this); });

    return function(state) {
	stateDiv.text(JSON.stringify(state.bs));
	console.log("state", state);
	cy.elements("node").style({"background-color":"gray"});
	cy.$('#' + state.node).style({"background-color": "red"});
    };
}

window.addEventListener("load", function(evt) {
    renderMachine("", thisSpec);
});


