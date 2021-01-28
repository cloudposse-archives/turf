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

	"github.com/cloudposse/turf/aws"
)

var autoEnableS3 bool

const autoEnableS3Flag string = "auto-enable-s3-protection"

var guardDutyAddMembersCmd = &cobra.Command{
	Use:     "set-administrator-account",
	Aliases: []string{"admin-account"},
	Short:   "Set GuardDuty administrator account and member accounts",
	Long:    "Designate the AWS Organization's AWS GuardDuty Admininstrator Account, then enable all the AWS Organization accounts as members",
	RunE: func(cmd *cobra.Command, args []string) error {
		return aws.EnableGuardDutyAdministratorAccount(region, administratorAccountRole, rootRole, autoEnableS3)
	},
}

func init() {
	guardDutyCmd.AddCommand(guardDutyAddMembersCmd)

	guardDutyAddMembersCmd.Flags().StringVarP(&administratorAccountRole, adminAccountRoleFlag, "a", "", "The ARN of a role to assume with access to the organization's GuardDuty Administrator Account")
	guardDutyAddMembersCmd.Flags().StringVarP(&rootRole, rootRoleFlag, "r", "", "The ARN of a role to assume with access to AWS Management Account")
	guardDutyAddMembersCmd.Flags().BoolVarP(&autoEnableS3, autoEnableS3Flag, "", false, "Auto-enable S3 protection")
}
