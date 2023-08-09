package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/alecthomas/kong"
)

const description = `
semver-select selects matching semvers from a list.

For example, get the newest version of go 1.15 like so:

  curl -Ls 'https://golang.org/dl/?mode=json&include=all' \
    | jq -r '.[].version' \
    | sed 's/^go//g' \
    | semver-select -i -c '1.15' -
`

var version = "unknown"

type rootCmd struct {
	Version            kong.VersionFlag `kong:"short=v,help='output semver-select version and exit'"`
	Constraint         string           `kong:"required,short=c,help='semver constraint to match'"`
	MaxResults         int              `kong:"short=n,help='maximum number of results to output'"`
	IgnoreInvalid      bool             `kong:"short=i,help='ignore invalid candidates instead of erroring'"`
	ValidateConstraint bool             `kong:"help='just validate the constraint. exits non-zero if invalid'"`
	Candidates         []string         `kong:"arg,optional,help='candidate versions to consider -- value of \"-\" indicates stdin'"`
}

func getVersions(args []string, stdin io.Reader, ignore bool) ([]*semver.Version, error) {
	res := make([]*semver.Version, 0, len(args))
	doStdin := false
	var err error
	for _, arg := range args {
		if arg == "-" {
			doStdin = true
			break
		}
		res, err = addVersion(arg, ignore, res)
		if err != nil {
			return nil, err
		}
	}
	if !doStdin {
		return res, nil
	}
	r := bufio.NewScanner(stdin)
	for r.Scan() {
		res, err = addVersion(r.Text(), ignore, res)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func addVersion(ver string, ignore bool, versions []*semver.Version) ([]*semver.Version, error) {
	v, err := semver.NewVersion(ver)
	if err != nil {
		if ignore {
			return versions, nil
		}
		return nil, fmt.Errorf("could not parse version %q: %v", ver, err)
	}
	return append(versions, v), nil
}

func main() {
	var cli rootCmd
	parser := kong.Must(
		&cli,
		kong.Vars{"version": version},
		kong.Description(strings.TrimSpace(description)),
	)
	run(parser, &cli, os.Stdin, os.Args[1:])
}

func run(parser *kong.Kong, cli *rootCmd, stdin io.Reader, args []string) {
	k, err := parser.Parse(args)
	if err != nil {
		parser.Fatalf(err.Error())
		return
	}
	c, err := semver.NewConstraint(cli.Constraint)
	if err != nil {
		k.Fatalf("invalid constraint: %q", cli.Constraint)
		return
	}
	if cli.ValidateConstraint {
		return
	}
	if len(cli.Candidates) == 0 {
		k.Fatalf("no candidates provided")
		return
	}

	versions, err := getVersions(cli.Candidates, stdin, cli.IgnoreInvalid)
	if err != nil {
		k.Fatalf(err.Error())
		return
	}

	for _, s := range results(c, cli.MaxResults, versions) {
		fmt.Fprintln(parser.Stdout, s)
	}
}

func results(c *semver.Constraints, max int, versions []*semver.Version) []string {
	candidates := make([]*semver.Version, 0, len(versions))
	for _, v := range versions {
		if c.Check(v) {
			candidates = append(candidates, v)
		}
	}
	sort.Sort(sort.Reverse(semver.Collection(candidates)))
	if max > 0 && max < len(candidates) {
		candidates = candidates[:max]
	}
	result := make([]string, len(candidates))
	for i, candidate := range candidates {
		result[i] = candidate.String()
	}
	return result
}
