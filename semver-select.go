package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
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
	GoVersions         bool             `kong:"name=go,help='allow go-style versions for candidates (e.g. 1.15rc1 or go1.20)'"`
	Orig               bool             `kong:"help='output original version strings instead of normalized versions'"`
	Candidates         []string         `kong:"arg,optional,help='candidate versions to consider -- value of \"-\" indicates stdin'"`
}

func getVersions(
	args []string,
	stdin io.Reader,
	ignore, goVersions bool,
) ([]*semver.Version, map[*semver.Version][]string, error) {
	res := make([]*semver.Version, 0, len(args))
	doStdin := false
	orig := map[*semver.Version][]string{}
	var err error
	for _, arg := range args {
		if arg == "-" {
			doStdin = true
			break
		}
		res, err = addVersion(arg, ignore, goVersions, res, orig)
		if err != nil {
			return nil, nil, err
		}
	}
	if !doStdin {
		return res, orig, nil
	}
	r := bufio.NewScanner(stdin)
	for r.Scan() {
		res, err = addVersion(r.Text(), ignore, goVersions, res, orig)
		if err != nil {
			return nil, nil, err
		}
	}
	return res, orig, nil
}

func addVersion(
	ver string,
	ignore, goVersions bool,
	versions []*semver.Version,
	orig map[*semver.Version][]string,
) ([]*semver.Version, error) {
	v, err := semver.NewVersion(ver)
	if err != nil && goVersions {
		v, err = parseGoVersion(ver)
	}
	if err != nil {
		if ignore {
			return versions, nil
		}
		return nil, fmt.Errorf("could not parse version %q", ver)
	}
	orig[v] = append(orig[v], ver)
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

	versions, orig, err := getVersions(cli.Candidates, stdin, cli.IgnoreInvalid, cli.GoVersions)
	if err != nil {
		k.Fatalf(err.Error())
		return
	}
	for _, s := range results(c, cli.MaxResults, versions, orig, cli.Orig) {
		fmt.Fprintln(parser.Stdout, s)
	}
}

func results(
	c *semver.Constraints,
	max int,
	versions []*semver.Version,
	orig map[*semver.Version][]string,
	useOrig bool,
) []string {
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
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if useOrig {
			result = append(result, orig[candidate]...)
			continue
		}
		result = append(result, candidate.String())
	}
	return result
}

var goPattern = regexp.MustCompile(`^(?:go)?([1-9]\d*)(?:\.(0|[1-9]\d*))?(?:\.(0|[1-9]\d*))?([a-zA-Z][a-zA-Z0-9.-]*)?$`)

func parseGoVersion(ver string) (*semver.Version, error) {
	matches := goPattern.FindStringSubmatch(ver)
	if len(matches) == 0 {
		return nil, fmt.Errorf("could not parse version %q", ver)
	}
	var major, minor, patch uint64
	var err error
	major, err = strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse major version %q: %v", ver, err)
	}
	if matches[2] != "" {
		minor, err = strconv.ParseUint(matches[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse minor version %q: %v", ver, err)
		}
	}
	if matches[3] != "" {
		patch, err = strconv.ParseUint(matches[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse patch version %q: %v", ver, err)
		}
	}
	return semver.New(major, minor, patch, matches[4], ""), nil
}
