// Copyright 2024 Nametag Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/kr/text"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

func init() {
	cobra.AddTemplateFuncs(template.FuncMap{
		"wrapFlagUsages": wrapFlagUsages,
		"wrapText":       wrapText,
	})
}

// Main is the entry point for the CLI.
func Main() {
	cmd := New()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func wrapFlagUsages(cmd *pflag.FlagSet) string {
	width := helpWidth()

	return cmd.FlagUsagesWrapped(width - 1)
}

func wrapText(s string) string {
	width := helpWidth()
	if width > 80 {
		width = 80
	}

	s = heredoc.Doc(s)
	paragraphs := strings.Split(s, "\n\n")
	for i, p := range paragraphs {
		if unicode.IsSpace(rune(p[0])) {
			continue
		}
		paragraphs[i] = text.Wrap(p, width-1)
	}
	return strings.Join(paragraphs, "\n\n")
}

func helpWidth() int {
	fd := int(os.Stdout.Fd())
	width := 80

	// Get the terminal width and dynamically set
	termWidth, _, err := term.GetSize(fd)
	if err == nil {
		width = termWidth
	}

	return min(120, width)
}

// identical to the default cobra help template, but utilizes wrapText
// https://github.com/spf13/cobra/blob/fd865a44e3c48afeb6a6dbddadb8a5519173e029/command.go#L580-L582
const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces | wrapText}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

// identical to the default cobra usage template, but utilizes wrapFlagUsages
// https://github.com/spf13/cobra/blob/fd865a44e3c48afeb6a6dbddadb8a5519173e029/command.go#L539-L568
const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{wrapFlagUsages .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{wrapFlagUsages .InheritedFlags | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
