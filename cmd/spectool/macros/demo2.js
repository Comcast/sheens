// This macro converts a structure like {macro:"demo2"} into 42.
register("demo2", function(x, path, root) {
    setIn(42, path, root);
});

