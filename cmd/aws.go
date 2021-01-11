package cmd

import (
	"github.com/spf13/cobra"
)

var region string
var profile string

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Commands related to automating AWS",
	Long:  "Commands related to automating AWS",
}

func init() {
	rootCmd.AddCommand(awsCmd)

	// Persistent flags for all AWS subcommands
	awsCmd.PersistentFlags().StringVar(&region, "region", "us-east-1", "The AWS region to operate on")
	awsCmd.PersistentFlags().StringVar(&profile, "profile", "default", "The AWS profile to assume to run commands")
}
