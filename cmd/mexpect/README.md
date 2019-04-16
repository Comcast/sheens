# A machine testing tool

## Summary

This tool can test a spec based on a test definition.

For example, see
[specs/test/double.test.yaml](../../specs/tests/double.test.yaml),
which specifies one set of tests for the spec
[specs/double.yaml](../../specs/double.yaml).

This tool will run [`siostd`](../siostd), so you'll need to
[build](../siostd/README.md) that program first if you don't already
have it.

## Usage

`mexpect` takes some optional command-line flags, and then the rest of
the command line is used to start a subprocess.

Example:

```Shell
# Get siostd and mexpect executables.
go get github.com/Comcast/sheens/cmd/...

# Run an mexpect test.
mexpect -d ../.. -show-in -show-out -f specs/tests/double.test.yaml siostd -tags=false
```
