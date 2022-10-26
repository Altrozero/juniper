package juniper

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

const (
	defaultCliEnv = "DEFAULT_CLI"
	defaultCli    = "server"
)

type CliCommandFunc func(args []string) error

type CliCommandEntry struct {
	Key   string
	Usage string
	Run   CliCommandFunc
}

// CommandEntries list
type CliCommandEntries []CliCommandEntry

// Find a command entry by its key
func (ces CliCommandEntries) Find(key string) *CliCommandEntry {
	for _, entry := range ces {
		if key == entry.Key {
			return &entry
		}
	}

	return nil
}

// CliUsage generator for the flags lib
func CliUsage(title, description, binName string, register CliCommandEntries) func() {
	return func() {
		var (
			builder strings.Builder
			ok      bool
		)

		builder.WriteString(title)
		builder.WriteByte('\n')

		if description != "" {
			builder.WriteString(description)
			builder.WriteByte('\n')
			builder.WriteByte('\n')
		}

		if binName == "" {
			if binName, ok = os.LookupEnv(defaultCliEnv); !ok {
				binName = defaultCli
			}
		}

		builder.WriteString(fmt.Sprintf("USAGE:\n    ./%s [options]\n\n", binName))
		builder.WriteString("OPTIONS:\n")

		fmt.Print(builder.String())
		flag.PrintDefaults()

		fmt.Print("\nCOMMANDS:\n")

		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		rows := make([]string, 0, len(register))
		for _, cmd := range register {
			for i, line := range strings.Split(cmd.Usage, "\n") {
				key := cmd.Key
				if i != 0 {
					key = "..."
				}

				rows = append(rows, fmt.Sprintf("    %s\t%s\n", green(key), yellow(line)))
			}
		}

		tbl := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', tabwriter.StripEscape)
		for _, row := range rows {
			fmt.Fprint(tbl, row)
		}

		tbl.Flush()
	}
}
