package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
)

func newBuffer(lines []string) io.Reader {
	return strings.NewReader(strings.Join(lines, "\n"))
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
			wantErr:  `could not parse version "invalid": Invalid Semantic Version`,
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
