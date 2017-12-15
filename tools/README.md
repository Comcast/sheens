# Some machine tools

## `mexpect`

This tool facilitates Machine Spec testing.  See the godoc for
`Session`, and see an example Session at
[`mexpect/double.test.yaml`](mexpect/double.test.yaml).

Command-line version at `mexpect/`.  Usage:

```Shell
go get github.com/Comcast/sheens/cmd/mservice
go get github.com/Comcast/sheens/tools/mexpect
# cd to this directory in the repo.
mexpect -f mexpect/double.test.yaml
```
