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

	"github.com/cloudposse/posse-cli/aws"
)

const adminAccountRoleFlag string = "administrator-account-role"
const rootRoleFlag string = "root-role"

var administratorAccountRole string
var rootRole string

var securityHubAddMembersCmd = &cobra.Command{
	Use:     "set-administrator-account",
	Aliases: []string{"admin-account"},
	Short:   "Set Security Hub administrator account and member accounts",
	Long:    "Designate the AWS Organization's AWS Security Hub Admininstrator Account, then enabled all the AWS Organization accounts as members",
	RunE: func(cmd *cobra.Command, args []string) error {
		return aws.EnableAdministratorAccount(region, administratorAccountRole, rootRole)
	},
}

func init() {
	securityhubCmd.AddCommand(securityHubAddMembersCmd)

	securityHubAddMembersCmd.Flags().StringVarP(&administratorAccountRole, adminAccountRoleFlag, "a", "", "The ARN of a role to assume with access to the organization's Security Hub Administrator Account")
	securityHubAddMembersCmd.Flags().StringVarP(&rootRole, rootRoleFlag, "r", "", "The ARN of a role to assume with access to AWS Management Account")
}
