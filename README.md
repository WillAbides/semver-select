# semver-select

[![godoc](https://pkg.go.dev/badge/github.com/willabides/semver-select.svg)](https://pkg.go.dev/github.com/willabides/semver-select)
[![ci](https://github.com/WillAbides/semver-select/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/WillAbides/semver-select/actions?query=workflow%3Aci+branch%3Amain+event%3Apush)

<!--- start usage output --->

```
Usage: semver-select --constraint=STRING <candidates> ...

semver-select selects matching semvers from a list.

For example, get the newest version of go 1.15 like so:

    curl -Ls 'https://golang.org/dl/?mode=json&include=all' \
      | jq -r '.[].version' \
      | sed 's/^go//g' \
      | semver-select -i -c '1.15' -

Arguments:
  <candidates> ...    candidate versions to consider -- value of "-" indicates stdin

Flags:
  -h, --help                   Show context-sensitive help.
  -v, --version                output semver-select version and exit
  -c, --constraint=STRING      semver constraint to match
  -n, --max-results=INT        maximum number of results to output
  -i, --ignore-invalid         ignore invalid candidates instead of erroring
      --validate-constraint    just validate the constraint. exits non-zero if invalid
```

<!--- end usage output --->
