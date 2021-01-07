package cmd

import (
	"github.com/cloudposse/posse-cli/common/posse"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of posse",
	Long:  "Print the version number of posse",
	Run: func(cmd *cobra.Command, args []string) {
		printPosseVersion()
	},
}

func printPosseVersion() {
	jww.FEEDBACK.Println(posse.BuildVersionString())
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
