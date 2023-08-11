package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
)

func newBuffer(lines []string) io.Reader {
	return strings.NewReader(strings.Join(lines, "\n"))
}

func Test_parseGoVersion(t *testing.T) {
	for _, td := range []struct {
		input string
		want  string
	}{
		{input: "go1.15.2", want: "1.15.2"},
		{input: "1.15.2", want: "1.15.2"},
		{input: "go1.15", want: "1.15.0"},
		{input: "1.15rc1", want: "1.15.0-rc1"},
		{input: "g1.15"},
		{input: "go1", want: "1.0.0"},
		{input: "1", want: "1.0.0"},
		{input: " "},
		{input: " 1"},
	} {
		t.Run(td.input, func(t *testing.T) {
			got, err := parseGoVersion(td.input)
			if td.want == "" {
				assert.EqualError(t, err, fmt.Sprintf("could not parse version %q", td.input))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, td.want, got.String())
		})
	}
	t.Run("overflow major", func(t *testing.T) {
		// one more than max uint64
		_, err := parseGoVersion("go18446744073709551616.2.3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse major version")
	})
	t.Run("overflow minor", func(t *testing.T) {
		// one more than max uint64
		_, err := parseGoVersion("go1.18446744073709551616.3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse minor version")
	})
	t.Run("overflow patch", func(t *testing.T) {
		// one more than max uint64
		_, err := parseGoVersion("go1.2.18446744073709551616")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse patch version")
	})
}

func Test_run(t *testing.T) {
	for _, td := range []struct {
		name     string
		args     []string
		stdin    io.Reader
		want     []string
		wantExit int
		wantErr  string
	}{
		{
			name: `validate "1.x"`,
			args: []string{"--validate-constraint", "-c", "1.x"},
		},
		{
			name:     `validate "1.invalid"`,
			args:     []string{"--validate-constraint", "-c", "1.invalid"},
			wantErr:  `invalid constraint: "1.invalid"`,
			wantExit: 1,
		},
		{
			name: "matches exact version",
			args: []string{"-c", "1.2.3", "1.2.0", "1.2.3-rc1", "1.2.3", "1.2.4"},
			want: []string{"1.2.3"},
		},
		{
			name: "outputs canonical version",
			args: []string{"-c", "1", "1.2", "1", "1.2.3", "1.2.4"},
			want: []string{"1.2.4", "1.2.3", "1.2.0", "1.0.0"},
		},
		{
			name: "honors --max-results",
			args: []string{"-c", "1", "--max-results", "2", "1.2", "1", "1.2.3", "1.2.4"},
			want: []string{"1.2.4", "1.2.3"},
		},
		{
			name:     "errors on invalid candidate",
			args:     []string{"-c", "1.2.3", "1.2.0", "1.2.3-rc1", "1.2.3", "1.2.4", "invalid"},
			wantExit: 1,
			wantErr:  `could not parse version "invalid"`,
		},
		{
			name: "accepts go version",
			args: []string{"--go", "-c", "1.2.3", "go1.2.0", "go1.2.3rc1", "go1.2.3", "go1.2.4"},
			want: []string{"1.2.3"},
		},
		{
			name: "ignores invalid candidate with --ignore-invalid",
			args: []string{"-c", "1.2.3", "--ignore-invalid", "1.2.0", "1.2.3-rc1", "1.2.3", "1.2.4", "invalid"},
			want: []string{"1.2.3"},
		},
		{
			name:     "errors on no candidates",
			args:     []string{"-c", "1.2.3"},
			wantExit: 1,
			wantErr:  "no candidates provided",
		},
		{
			name:  "accepts stdin",
			args:  []string{"-c", "1.2.3", "-"},
			stdin: newBuffer([]string{"1.2.0", "1.2.3-rc1", "1.2.3", "1.2.4"}),
			want:  []string{"1.2.3"},
		},
		{
			name: "outputs original with --orig",
			args: []string{"--go", "--orig", "-c", "1.2.3", "go1.2.0", "go1.2.3rc1", "go1.2.3", "go1.2.4", "1.2.3"},
			want: []string{"go1.2.3", "1.2.3"},
		},
	} {
		t.Run(td.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exitVal := 0
			var cli rootCmd
			parser := kong.Must(
				&cli,
				kong.Name("semver-select"),
				kong.Vars{"version": version},
				kong.Description(strings.TrimSpace(description)),
				kong.Writers(&stdout, &stderr),
				kong.Exit(func(i int) { exitVal = i }),
			)
			run(parser, &cli, td.stdin, td.args)
			assert.Equal(t, td.wantExit, exitVal, "exit value")
			assert.Equal(
				t,
				td.wantErr,
				strings.TrimPrefix(strings.TrimSpace(stderr.String()), "semver-select: error: "),
				"stderr",
			)
			gotOut := strings.TrimSpace(stdout.String())
			if len(td.want) == 0 {
				assert.Equal(t, "", gotOut, "stdout")
			} else {
				got := strings.Split(gotOut, "\n")
				assert.Equal(t, td.want, got, "stdout")
			}
		})
	}
}
