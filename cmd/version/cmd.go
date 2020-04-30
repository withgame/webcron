package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BinAppName      string
	BinBuildCommit  string
	BinBuildVersion string
	BinBuildDate    string
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Long:  `version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AppName   : %s\n", BinAppName)
		fmt.Printf("Version   : %s\n", BinBuildVersion)
		fmt.Printf("Commit    : %s\n", BinBuildCommit)
		fmt.Printf("BuildDate : %s\n", BinBuildDate)
	},
}
