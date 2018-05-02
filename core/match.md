# Pattern matching examples

Generated from test cases.


## 1. Simple matching example


A very basic test that shows how a pattern variable (`?likes`) gets bound during matching.
The pattern
```JSON
{"likes":"?likes"}
```

matched against
```JSON
{"likes":"tacos"}
```

should return
```JSON
[{"?likes":"tacos"}]
```

## 2. Variable with constant


A map with a variable and a constant.
The pattern
```JSON
{"likes":"?likes","when":"now"}
```

matched against
```JSON
{"likes":"tacos","when":"now"}
```

should return
```JSON
[{"?likes":"tacos"}]
```

## 4. Two constants


A map with two constants.
The pattern
```JSON
{"likes":"queso","when":"now"}
```

matched against
```JSON
{"likes":"queso","when":"now"}
```

should return
```JSON
[{}]
```

## 6. Multiple variables


This simple example shows bindings for two pattern variables.
The pattern
```JSON
{"likes":"?likes","wants":"?wants"}
```

matched against
```JSON
{"likes":"tacos","wants":"queso"}
```

should return
```JSON
[{"?likes":"tacos","?wants":"queso"}]
```

## 7. Deeper variable


Pattern matching is fully structured
The pattern
```JSON
{"needs":{"tacos":{"n":"?n"}}}
```

matched against
```JSON
{"needs":{"tacos":{"n":2}}}
```

should return
```JSON
[{"?n":2}]
```

## 8. Same variable twice (good)


If you use a pattern variable more than once, then the bindings must agree.  See the next example.
The pattern
```JSON
{"n":"?n","needs":{"tacos":{"n":"?n"}}}
```

matched against
```JSON
{"n":2,"needs":{"tacos":{"n":2}}}
```

should return
```JSON
[{"?n":2}]
```

## 9. Same variable twice (bad)


If you use a pattern variable more than once, then the bindings must agree.  See the previous example.
The pattern
```JSON
{"n":"?n","needs":{"tacos":{"n":"?n"}}}
```

matched against
```JSON
{"n":3,"needs":{"tacos":{"n":2}}}
```

should return
```JSON
[]
```

## 10. Array as a set


An array is treated as a set.
The pattern
```JSON
{"a":["?a"],"is":"?a"}
```

matched against
```JSON
{"a":[1,2,3,4],"is":3}
```

should return
```JSON
[{"?a":3}]
```

## 11. Array with a variable and a constant


An array is treated as a set; multiple bindings possible.
The pattern
```JSON
["a","?x"]
```

matched against
```JSON
["a","b","c"]
```

should return
```JSON
[{"?x":"b"},{"?x":"c"}]
```

## 12. Array with a variable and a map constant

The pattern
```JSON
[{"likes":"tacos"},"?x"]
```

matched against
```JSON
[{"likes":"tacos"},"b","c"]
```

should return
```JSON
[{"?x":"b"},{"?x":"c"}]
```

## 13. Array with a variable and a constant; message with map elements

The pattern
```JSON
["a","b","?x"]
```

matched against
```JSON
[{"likes":"tacos"},"b","a"]
```

should return
```JSON
[{"?x":{"likes":"tacos"}}]
```

## 14. Array with a map containing a variable

The pattern
```JSON
["a","b",{"likes":"?x"}]
```

matched against
```JSON
[{"likes":"tacos"},{"likes":"chips"},"b","a"]
```

should return
```JSON
[{"?x":"tacos"},{"?x":"chips"}]
```

## 15. Array as a set; multiple bss; backtracking


An array is treated as a set.
The pattern
```JSON
{"a":["?a"],"is":["?a"]}
```

matched against
```JSON
{"a":[1,2,3,4],"is":[2,3]}
```

should return
```JSON
[{"?a":2},{"?a":3}]
```

## 16. Bad array vars


Two pattern variables inside an array isn't allowed (because the computational complexity means that some input could be very costly to process).
The pattern
```JSON
{"a":["?x","?y"]}
```

matched against
```JSON
{"a":[1]}
```

should return an error.

## 17. Property variable vars


You can have _at most one_ pattern variable as a key in a given map.
The pattern
```JSON
{"?x":1}
```

matched against
```JSON
{"n":1}
```

should return
```JSON
[{"?x":"n"}]
```

## 18. Multiple property variable vars


You can have _at most one_ pattern variable as a key in a given map.
The pattern
```JSON
{"?x":1,"?y":2}
```

matched against
```JSON
{"m":2,"n":1}
```

should return an error.

## 19. A null value

The pattern
```JSON
{"wants":"?wants"}
```

matched against
```JSON
{"needs":null,"wants":"tacos"}
```

should return
```JSON
[{"?wants":"tacos"}]
```

## 20. Type conflict: int/string

The pattern
```JSON
{"wants":1}
```

matched against
```JSON
{"wants":"one"}
```

should return
```JSON
[]
```

## 21. Type conflict: int/bool

The pattern
```JSON
{"wants":1}
```

matched against
```JSON
{"wants":true}
```

should return
```JSON
[]
```

## 22. Anonymous variable used twice

The pattern
```JSON
{"count":"?","wants":"?"}
```

matched against
```JSON
{"count":48,"wants":"tacos"}
```

should return
```JSON
[{}]
```

## 23. Anonymous variable with normal variable

The pattern
```JSON
{"count":"?","wants":"?","when":"?when"}
```

matched against
```JSON
{"count":48,"wants":"tacos","when":"today"}
```

should return
```JSON
[{"?when":"today"}]
```

## 24. Anonymous variable as a property variable

The pattern
```JSON
{"?":"tacos"}
```

matched against
```JSON
{"likes":"tacos","needs":"chips"}
```

should return
```JSON
[{}]
```

## 25. Anonymous variable as a property variable and another variable

The pattern
```JSON
{"?":{"likes":"?likes"}}
```

matched against
```JSON
{"homer":{"likes":"tacos"}}
```

should return
```JSON
[{"?likes":"tacos"}]
```

## 26. Anonymous variable as a property variable without a match

The pattern
```JSON
{"?":"tacos"}
```

matched against
```JSON
{"needs":"chips"}
```

should return
```JSON
[]
```

## 34. Inequality: success

The pattern
```JSON
{"n":"?\u003cn"}
```

matched against
```JSON
{"n":3}
```

should return
```JSON
[{"?\u003cn":10,"?n":3}]
```

## 35. Inequality: failure

The pattern
```JSON
{"n":"?\u003cn"}
```

matched against
```JSON
{"n":3}
```

should return
```JSON
[]
```

## 40. Inequality: success (>=)

The pattern
```JSON
{"n":"?\u003e=n"}
```

matched against
```JSON
{"n":11}
```

should return
```JSON
[{"?\u003e=n":11,"?n":11}]
```

## 41. Inequality: failure (>=)

The pattern
```JSON
{"n":"?\u003e=n"}
```

matched against
```JSON
{"n":11}
```

should return
```JSON
[]
```

## 44. Inequality: non-numeric

The pattern
```JSON
{"n":"?\u003cn"}
```

matched against
```JSON
{"n":"queso"}
```

should return
```JSON
[]
```

## 45. Inequality: given same

The pattern
```JSON
{"n":"?\u003cn"}
```

matched against
```JSON
{"n":3}
```

should return
```JSON
[{"?\u003cn":10,"?n":3}]
```

## 46. Inequality: given different

The pattern
```JSON
{"n":"?\u003cn"}
```

matched against
```JSON
{"n":3}
```

should return
```JSON
[]
```

## 47. Inequality: used later

The pattern
```JSON
{"needs":"?n","wants":{"n":"?\u003cn"}}
```

matched against
```JSON
{"needs":3,"wants":{"n":3}}
```

should return
```JSON
[{"?\u003cn":10,"?n":3}]
```

## 48. Inequality: used later with conflict

The pattern
```JSON
{"needs":"?n","wants":{"n":"?\u003cn"}}
```

matched against
```JSON
{"needs":4,"wants":{"n":3}}
```

should return
```JSON
[]
```

## 49. Optional pattern variable (absent)

The pattern
```JSON
{"opt":"??maybe","wants":"?wanted"}
```

matched against
```JSON
{"wants":"tacos"}
```

should return
```JSON
[{"?wanted":"tacos"}]
```

## 50. Optional pattern variable (present)

The pattern
```JSON
{"a":"??maybe","wants":"?wanted"}
```

matched against
```JSON
{"a":"queso","wants":"tacos"}
```

should return
```JSON
[{"??maybe":"queso","?wanted":"tacos"}]
```

## 52. Optional pattern variable (array, absent)

The pattern
```JSON
["??opt"]
```

matched against
```JSON
[]
```

should return
```JSON
[{}]
```

## 53. Optional pattern variable (array, present)

The pattern
```JSON
["??opt","a","b"]
```

matched against
```JSON
["a","b"]
```

should return
```JSON
[{}]
```

## 54. Optional pattern variable (array, present)

The pattern
```JSON
["??opt","a","b"]
```

matched against
```JSON
["a","b","c"]
```

should return
```JSON
[{"??opt":"c"}]
```

## 55. Optional pattern variable (array, present, multiple bindings)

The pattern
```JSON
["??opt","a","b"]
```

matched against
```JSON
["a","b","c","d"]
```

should return
```JSON
[{"??opt":"c"},{"??opt":"d"}]
```
