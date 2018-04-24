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

## 2. Multiple variables


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

## 3. Deeper variable


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

## 4. Same variable twice (good)


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

## 5. Same variable twice (bad)


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

## 6. Bad array vars


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

## 7. Property variable vars


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

## 8. Multiple property variable vars


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

## 9. Array as a set


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

## 10. Array as a set; multiple bss; backtracking


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

## 11. A null value

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

## 12. Type conflict: int/string

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

## 13. Type conflict: int/bool

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

## 14. Anonymous variable used twice

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

## 15. Anonymous variable with normal variable

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

## 16. Anonymous variable as a property variable

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

## 17. Anonymous variable as a property variable and another variable

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

## 18. Anonymous variable as a property variable without a match

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

## 26. Inequality: success

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

## 27. Inequality: failure

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

## 32. Inequality: success (>=)

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

## 33. Inequality: failure (>=)

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

## 36. Inequality: non-numeric

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

## 37. Inequality: given same

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

## 38. Inequality: given different

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

## 39. Inequality: used later

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

## 40. Inequality: used later with conflict

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

## 41. Optional pattern variable (absent)

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

## 42. Optional pattern variable (present)

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

## 43. Optional pattern variable (array, absent)

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

## 44. Optional pattern variable (array, present)

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

## 45. Optional pattern variable (array, present)

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
