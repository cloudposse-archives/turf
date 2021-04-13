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
	"github.com/cloudposse/turf/aws"
	"github.com/spf13/cobra"
)

const cloudTrailAccountFlag string = "cloud-trail-account"
const globalCollectionRegionFlag string = "global-collector-region"

var isCloudTrailAccount bool
var globalCollectionRegion string

var securityHubDisableGlobalControlsCmd = &cobra.Command{
	Use:   "disable-global-controls",
	Short: "Disables Security Hub Global Resources controls in regions that aren't collecting Global Resources",
	Long: `
	Disables Security Hub Global Resources controls in regions that aren't collecting Global Resources and disables
	CloudTrail related controls in accounts that are not the central CloudTrail account.

	See the following AWS documentation for additional information:

	https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-standards-fsbp-to-disable.html
	https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-standards-cis-to-disable.html
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return aws.DisableSecurityHubGlobalResourceControls(globalCollectionRegion, role, isPrivileged, isCloudTrailAccount)
	},
}

func init() {
	securityHubDisableGlobalControlsCmd.Flags().StringVarP(&globalCollectionRegion, globalCollectionRegionFlag, "g", region, "The AWS Region that contains the global resource collector")
	securityHubDisableGlobalControlsCmd.Flags().StringVar(&role, roleFlag, "", "The ARN of a role to assume")
	securityHubDisableGlobalControlsCmd.Flags().BoolVarP(&isPrivileged, isPrivilegedFlag, "", false, "Flag to indicate if the session already has rights to perform the actions in AWS")
	securityHubDisableGlobalControlsCmd.Flags().BoolVar(&isCloudTrailAccount, cloudTrailAccountFlag, false, "A flag to indicate if this account is the central CloudTrail account")

	securityHubDisableGlobalControlsCmd.MarkFlagRequired(globalCollectionRegionFlag)

	securityhubCmd.AddCommand(securityHubDisableGlobalControlsCmd)
}
