[
    {
	"title": "Simple matching example",
	"p": {"likes":"?likes"},
	"m": {"likes":"tacos"},
	"w": [{"?likes":"tacos"}],
	"doc": "A very basic test that shows how a pattern variable (`?likes`) gets bound during matching."
    },
    {
	"title": "Variable with constant",
	"p": {"likes":"?likes","when":"now"},
	"m": {"likes":"tacos","when":"now"},
	"w": [{"?likes":"tacos"}],
	"doc": "A map with a variable and a constant."
    },
    {
	"title": "Variable with constant (different order)",
	"p": {"when":"now","likes":"?likes"},
	"m": {"likes":"tacos","when":"now"},
	"w": [{"?likes":"tacos"}],
	"nodoc": true
    },
    {
	"title": "Two constants",
	"p": {"likes":"queso","when":"now"},
	"m": {"likes":"queso","when":"now"},
	"w": [{}],
	"doc": "A map with two constants."
    },
    {
	"title": "Two constants (different order)",
	"p": {"when":"now","likes":"queso"},
	"m": {"likes":"queso","when":"now"},
	"w": [{}],
	"nodoc": true
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
	"title": "Array as a set",
	"p": {"a":["?a"],"is":"?a"},
	"m": {"a":[1,2,3,4],"is":3},
	"w": [{"?a":3}],
	"doc": "An array is treated as a set."
    },
    {
	"title": "Array with a variable and a constant",
	"p": ["a","?x"],
	"m": ["a","b","c"],
	"w": [{"?x":"b"},{"?x":"c"}],
	"doc": "An array is treated as a set; multiple bindings possible."
    },
    {
	"title": "Array with a variable and a map constant",
	"p": [{"likes":"tacos"},"?x"],
	"m": [{"likes":"tacos"},"b","c"],
	"w": [{"?x":"b"},{"?x":"c"}]
    },
    {
	"title": "Array with a variable and a constant; message with map elements",
	"p": ["a", "b", "?x"],
	"m": [{"likes":"tacos"},"b","a"],
	"w": [{"?x":{"likes":"tacos"}}]
    },
    {
	"title": "Array with a map containing a variable",
	"p": ["a", "b", {"likes":"?x"}],
	"m": [{"likes":"tacos"},{"likes":"chips"},"b","a"],
	"w": [{"?x":"tacos"},{"?x":"chips"}]
    },
    {
	"title": "Array as a set; multiple bss; backtracking",
	"p": {"a":["?a"],"is":["?a"]},
	"m": {"a":[1,2,3,4],"is":[2,3]},
	"w": [{"?a":2},{"?a":3}],
	"doc": "An array is treated as a set."
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
    },
    {
	"title": "Inequality: success",
	"p": {"n":"?<n"},
	"m": {"n":3},
	"b": {"?<n":10},
	"w": [{"?n":3,"?<n":10}]
    },
    {
	"title": "Inequality: failure",
	"p": {"n":"?<n"},
	"m": {"n":3},
	"b": {"?<n":2},
	"w": []
	
    },
    
    {
	"title": "Inequality: success (<=)",
	"noDoc": true,
	"p": {"n":"?<=n"},
	"m": {"n":3},
	"b": {"?<=n":3},
	"w": [{"?n":3,"?<=n":3}]
    },
    {
	"title": "Inequality: failure (<=)",
	"noDoc": true,
	"p": {"n":"?<=n"},
	"m": {"n":4},
	"b": {"?<=n":3},
	"w": []
    },
    

    {
	"title": "Inequality: success (>)",
	"noDoc": true,
	"p": {"n":"?>n"},
	"m": {"n":11},
	"b": {"?>n":10},
	"w": [{"?n":11,"?>n":10}]
    },
    {
	"title": "Inequality: failure (>)",
	"noDoc": true,
	"p": {"n":"?>n"},
	"m": {"n":11},
	"b": {"?>n":12},
	"w": []
    },
    
    {
	"title": "Inequality: success (>=)",
	"p": {"n":"?>=n"},
	"m": {"n":11},
	"b": {"?>=n":11},
	"w": [{"?n":11,"?>=n":11}]
    },
    {
	"title": "Inequality: failure (>=)",
	"p": {"n":"?>=n"},
	"m": {"n":11},
	"b": {"?>=n":12},
	"w": []
    },
    

    {
	"title": "Inequality: success (!=)",
	"noDoc": true,
	"p": {"n":"?!=n"},
	"m": {"n":11},
	"b": {"?!=n":13},
	"w": [{"?n":11,"?!=n":13}]
    },
    {
	"title": "Inequality: failure (!=)",
	"noDoc": true,
	"p": {"n":"?!=n"},
	"m": {"n":21},
	"b": {"?!=n":21},
	"w": []
    },
    


    {
	"title": "Inequality: non-numeric",
	"p": {"n":"?<n"},
	"m": {"n":"queso"},
	"b": {"?<n":2},
	"w": []
    },
    {
	"title": "Inequality: given same",
	"p": {"n":"?<n"},
	"m": {"n":3},
	"b": {"?<n":10,"?n":3},
	"w": [{"?n":3,"?<n":10}]
    },
    {
	"title": "Inequality: given different",
	"p": {"n":"?<n"},
	"m": {"n":3},
	"b": {"?<n":10,"?n":4},
	"w": []
    },
    {
	"title": "Inequality: used later",
	"p": {"wants":{"n":"?<n"},"needs":"?n"},
	"m": {"wants":{"n":3},"needs":3},
	"b": {"?<n":10},
	"w": [{"?n":3,"?<n":10}]
    },
    {
	"title": "Inequality: used later with conflict",
	"p": {"wants":{"n":"?<n"},"needs":"?n"},
	"m": {"wants":{"n":3},"needs":4},
	"b": {"?<n":10},
	"w": []
    },
    {
	"title": "Optional pattern variable (absent)",
	"p": {"wants":"?wanted","opt":"??maybe"},
	"m": {"wants":"tacos"},
	"b": {},
	"w": [{"?wanted":"tacos"}]
    },
    {
	"title": "Optional pattern variable (present)",
	"p": {"wants":"?wanted","a":"??maybe"},
	"m": {"wants":"tacos","a":"queso"},
	"b": {},
	"w": [{"?wanted":"tacos","??maybe":"queso"}]
    },
    {
	"title": "Optional pattern variable (array, absent)",
	"p": ["??opt"],
	"m": [],
	"b": {},
	"w": [{}]
    },
    {
	"title": "Optional pattern variable (array, present)",
	"p": ["??opt", "a", "b"],
	"m": ["a","b"],
	"b": {},
	"w": [{}]
    },
    {
	"title": "Optional pattern variable (array, present)",
	"p": ["??opt", "a", "b"],
	"m": ["a","b","c"],
	"b": {},
	"w": [{"??opt":"c"}]
    }
]
