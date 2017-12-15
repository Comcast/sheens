// This macro converts a structure like {macro:"demo1",x:1} into
// {a:x+1,b:{macro:"demo2"}}.
register("demo1", function(x, path, root) {
    setIn({a: x.x+1, b: {macro: "demo2"}}, path, root);
});

