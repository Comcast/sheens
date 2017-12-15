## Environment

...

1.  `_.cronNext(CRONEXPR)` → `TIMESTAMP`: `cronNext` attempts to parse
    its argument as a
    [cron expression](https://github.com/gorhill/cronexpr). If
    successful, returns the next time in
    [Go RFC3339Nano](https://golang.org/pkg/time/#pkg-constants)
    format.
   
    Example:
	
	```Javascript
	({next: _.cronNext("* 0 * * *")});
	```
   
...
