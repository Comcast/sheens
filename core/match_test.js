[
    {
	"title": "Simple matching example",
	"p": {"likes":"?likes"},
	"m": {"likes":"tacos"},
	"w": [{"?likes":"tacos"}],
	"doc": "A very basic test that shows how a pattern variable (`?likes`) gets bound during matching."
    },
    {
	"title": "Multiple variables",
	"p": {"likes":"?likes","wants":"?wants"},
	"m": {"likes":"tacos","wants":"queso"},
	"w": [{"?likes":"tacos","?wants":"queso"}],
	"doc": "This simple example shows bindings for two pattern variables."
    },
    {
	"title": "Deeper variable",
	"p": {"needs":{"tacos":{"n":"?n"}}},
	"m": {"needs":{"tacos":{"n":2}}},
	"w": [{"?n":2}],
	"doc": "Pattern matching is fully structured"
    },
    {
	"title": "Same variable twice (good)",
	"p": {"needs":{"tacos":{"n":"?n"}}, "n":"?n"},
	"m": {"needs":{"tacos":{"n":2}}, "n":2},
	"w": [{"?n":2}],
	"doc": "If you use a pattern variable more than once, then the bindings must agree.  See the next example."
    },
    {
	"title": "Same variable twice (bad)",
	"p": {"needs":{"tacos":{"n":"?n"}}, "n":"?n"},
	"m": {"needs":{"tacos":{"n":2}}, "n":3},
	"w": [],
	"doc": "If you use a pattern variable more than once, then the bindings must agree.  See the previous example."
    },
    {
	"title": "Bad array vars",
	"p": {"a":["?x","?y"]},
	"m": {"a":[1]},
	"err": true,
	"doc": "Two pattern variables inside an array isn't allowed (because the computational complexity means that some input could be very costly to process)."
    },
    {
	"title": "Property variable vars",
	"p": {"?x":1},
	"m": {"n":1},
	"w": [{"?x":"n"}],
	"doc": "You can have _at most one_ pattern variable as a key in a given map."
    },
    {
	"title": "Multiple property variable vars",
	"p": {"?x":1,"?y":2},
	"m": {"n":1, "m": 2},
	"err": true,
	"doc": "You can have _at most one_ pattern variable as a key in a given map."
    },
    {
	"title": "Array as a set",
	"p": {"a":["?a"],"is":"?a"},
	"m": {"a":[1,2,3,4],"is":3},
	"w": [{"?a":3}],
	"doc": "An array is treated as a set."
    },
    {
	"title": "Array as a set; multiple bss; backtracking",
	"p": {"a":["?a"],"is":["?a"]},
	"m": {"a":[1,2,3,4],"is":[2,3]},
	"w": [{"?a":2},{"?a":3}],
	"doc": "An array is treated as a set."
    },
    {
	"title": "A null value",
	"p": {"wants":"?wants"},
	"m": {"needs":null, "wants":"tacos"},
	"w": [{"?wants":"tacos"}]
    },
    {
	"title": "Type conflict: int/string",
	"p": {"wants":1},
	"m": {"wants":"one"},
	"w": []
    },
    {
	"title": "Type conflict: int/bool",
	"p": {"wants":1},
	"m": {"wants":true},
	"w": []
    },
    {
	"title": "Anonymous variable used twice",
	"p": {"wants":"?","count":"?"},
	"m": {"wants":"tacos","count":48},
	"w": [{}]
    },
    {
	"title": "Anonymous variable with normal variable",
	"p": {"wants":"?","count":"?","when":"?when"},
	"m": {"wants":"tacos","count":48,"when":"today"},
	"w": [{"?when":"today"}]
    },
    {
	"title": "Anonymous variable as a property variable",
	"p": {"?":"tacos"},
	"m": {"likes":"tacos","needs":"chips"},
	"w": [{}]
    },
    {
	"title": "Anonymous variable as a property variable and another variable",
	"p": {"?":{"likes":"?likes"}},
	"m": {"homer":{"likes":"tacos"}},
	"w": [{"?likes":"tacos"}]
    },
    {
	"title": "Anonymous variable as a property variable without a match",
	"p": {"?":"tacos"},
	"m": {"needs":"chips"},
	"w": []
    },
    {
	"title": "Benchmark: array 1",
	"p": {"a":["?x"]},
	"m": {"a":[1]},
	"benchmarkOnly": true
    },
    {
	"title": "Benchmark: array 2",
	"p": {"a":["?x"]},
	"m": {"a":[1,2]},
	"benchmarkOnly": true
    },
    {
	"title": "Benchmark: array 3",
	"p": {"a":["?x"]},
	"m": {"a":[1,2,3]},
	"benchmarkOnly": true
    },
    {
	"title": "Benchmark: array 4",
	"p": {"a":["?x"]},
	"m": {"a":[1,2,3,4]},
	"benchmarkOnly": true
    },
    {
	"title": "Benchmark: array 4x4 one",
	"p": {"a":["?x"],"b":["?x"]},
	"m": {"a":[1,2,3,4],"b":[1,2,3,4]},
	"benchmarkOnly": true
    },
    {
	"title": "Benchmark: array 2x2",
	"p": {"a":["?x"],"b":["?y"]},
	"m": {"a":[1,2],"b":[1,2]},
	"benchmarkOnly": true
    },
    {
	"title": "Benchmark: A duktape comparison",
	"p": {"b":[{"a":"?x","d":"?y"}],"c":"?x"},
	"m": {"b":[1,{"a": 2},{"a":3,"d":5},{"c":4},{"a":3}],"c":3},
	"benchmarkOnly": true
    }
]
