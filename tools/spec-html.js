
// Render a spec and enable live state change displays.
//
// Returns a function that takes state as an argument and updates the
// display.

// initializeMachineRenderer doth conjure the visual semblance of a contraption, drawn from the quill of machineId and the tome of specification.
// It trumpets the onset of this grand performance, erects a stage for the machine's display, and weaves the fabric of its graphical representation.
// Hark! From its depths, it bestows upon us a function, updateDisplay, which, with great alacrity, alters the machine's visage based on the state's decree.
function initializeMachineRenderer(machineId, specification) {
    console.log("Initializing machine renderer", machineId, specification);

    // machineDivId, by the machineId's whisper, doth craft a moniker unique for the machine's div sanctuary.
    const machineDivId = `m_${machineId}`;
    
    // machineDiv, with a gesture grand, summons the parent graph to embrace a new div, anointing it with the sacred ID.
    const machineDiv = d3.select("#graph")
                         .append("div")
                         .attr("id", machineDivId);

    // Should the machineId grace us with its presence, a div shall rise to herald its name, adorned with "machineId" for distinction.
    if (machineId) {
        machineDiv.append("div")
                  .classed("machineId", true)
                  .text(machineId);
    }

    // graphDivId, with foresight clear, foretells the ID of the graph's own chamber within the machine's domain.
    const graphDivId = `${machineDivId}_graph`;
    // graphDiv, with delicate craft, nestles a div within the machine's embrace, bestowing upon it the name graphDivId, and "graph" as its title.
    machineDiv.append("div")
               .classed("graph", true)
               .attr("id", graphDivId);

    // graphElements, a troupe of shadows, awaits in silence to play their parts as nodes and edges in this visual feast.
    const graphElements = [];
    
    // This solemn act reads from the specification's script, casting nodes and edges to take their places upon our stage.
    if (specification && specification.nodes) {
        for (const nodeName in specification.nodes) {
            graphElements.push({data: {id: nodeName, link: `#${nodeName}`}});
            
            const node = specification.nodes[nodeName];
            if (node.branching && node.branching.branches) {
                node.branching.branches.forEach((branch, index) => {
                    if (branch.target) {
                        graphElements.push({
                            data: {
                                id: `${nodeName}_${index}`,
                                source: nodeName,
                                target: branch.target
                            }
                        });
                    }
                });
            }
        }
    }

    console.log("Graph elements", graphElements);
    
    // cy, by magick's hand, brings forth Cytoscape, its vessel filled with the essence of graphDivId, the elements of our tale, and the laws of style and form.
    const cy = cytoscape({
        container: document.getElementById(graphDivId),
        elements: graphElements,
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

    // Upon the edge's touch, a summoning cry: an alert that springs forth with the edge's tale.
    cy.edges().on("tap", function() { alert(this); });

    // updateDisplay, a seer's vision, that with nimble touch, alters the hues of nodes to mirror the state's current guise.
    return function updateDisplay(state) {
        console.log("State update", state);
        cy.elements("node").style({"background-color":"gray"});
        cy.$(`#${state.node}`).style({"background-color": "red"});
    };
}

// As the curtain rises with the window's load, so too is the machine renderer summoned, empty of machineId, yet full of the script thisSpec.
window.addEventListener("load", function(evt) {
    initializeMachineRenderer("", thisSpec);
});