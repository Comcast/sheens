# A machine testing tool

## Summary

This tool can test a spec based on a test definition.

For example, see
[specs/test/double.test.yaml](../../specs/test/double.test.yaml),
which specifies one set of tests for the spec
[specs/double.yaml](../../specs/double.yaml).

This tool will invoke [`mcrew`](../mcrew), so you'll need to
[build](../mcrew/README.md) that program first.

## Usage

From this directory:

```Shell
(cd ../mcrew && go install) # If you haven't already
go install
mexpect -f ../../specs/tests/double.test.yaml -s ../../specs
```
