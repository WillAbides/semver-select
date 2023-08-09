# semver-select

[Contributions welcome](./CONTRIBUTING.md).

## Install With [bindown](https://github.com/WillAbides/bindown)

```shell
bindown template-source add semver-select https://github.com/WillAbides/semver-select/releases/latest/download/bindown.yaml
bindown dependency add semver-select --source semver-select -y
```

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
