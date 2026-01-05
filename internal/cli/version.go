package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scmd/scmd/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version.Info())
	},
}
