# A tool to work with specs

Docs ToDo.

## Example

```Shell
cat demo.yaml | \
  spectool inlines -d . | \
  spectool macroexpand | \
  spectool addOrderedOutMessages -p 'lunch_' -s '00' -e done -d 3s \
     -m '[{"e":{"order":"beer"},"r":{"deliver":"beer"}},{"e":{"order":"queso"},"r":{"deliver":"queso"}},{"e":{"order":"tacos"},"r":{"deliver":"tacos"}}]' | \
  spectool addGenericCancelNode | \
  spectool addMessageBranches -P -p '{"ctl":"cancel"}' -t cancel | \
  spectool dot | \
  spectool analyze |
  spectool yamltojson > \
  demo.json
```
