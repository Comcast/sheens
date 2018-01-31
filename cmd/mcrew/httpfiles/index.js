window.addEventListener("load", function(evt) {
    var xhr = new XMLHttpRequest();
    xhr.open('GET', '/specs', true);
    xhr.onreadystatechange = function(e) {
	if (this.readyState == 4 && this.status == 200) {
	    var specs = JSON.parse(this.response);
	    var ol = d3.select("#specs");
	    for (var i = 0; i < specs.length; i++) {
		var spec = specs[i];
		var baseName = spec.substr(0, spec.indexOf("."));
		ol.append("li")
		    .classed("specName", true)
		    .append("a")
		    .attr("href", "/specs/" + baseName + ".html")
		    .append("code")
		    .text(baseName);
	    }
	}
    };
    xhr.send();
});


