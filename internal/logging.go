package internal

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/fatih/color"
)

// DefaultInitialPadding is the default padding in the log library.
const DefaultInitialPadding = 3

// ExtraPadding is the double of the DefaultInitialPadding.
const ExtraPadding = DefaultInitialPadding * 2

// LogTitle pretty prints a given title.
func LogTitle(title string) {
	cli.Default.Padding = DefaultInitialPadding

	log.Info(color.New(color.Bold).Sprint(strings.ToUpper(title)))

	cli.Default.Padding = ExtraPadding
}

func Pad(s string) string {
	return fmt.Sprintf("%-50v", s)
}
