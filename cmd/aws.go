/*
Copyright © 2021 Cloud Posse, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/spf13/cobra"
)

var region string
var profile string
var role string

// These flags are used in the AWS sub-commands
const roleFlag string = "role"
const isPrivilegedFlag string = "privileged"
const adminAccountRoleFlag string = "administrator-account-role"
const rootRoleFlag string = "root-role"

var administratorAccountRole string
var rootRole string

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
