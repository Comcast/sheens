# Pattern matching examples

Generated from test cases.


## 0. Simple matching example


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

## 1. Multiple variables


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

## 2. Deeper variable


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

## 3. Same variable twice (good)


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

## 4. Same variable twice (bad)


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

## 5. Bad array vars


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

## 6. Property variable vars


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

## 7. Multiple property variable vars


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

## 8. Array as a set


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

## 9. Array as a set; multiple bss; backtracking


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

## 10. A null value

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

## 11. Type conflict: int/string

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

## 12. Type conflict: int/bool

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
