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

var shouldDelete bool

const shouldDeleteFlag string = "delete"

var deleteDefaultVPCsCmd = &cobra.Command{
	Use:   "delete-default-vpcs",
	Short: "Delete the default VPCs in each region of the account",
	Long: `Best-practices call for not using the default VPC, but rather, creating a new set of VPCs as necessary. 
	AWS Security Hub will flag the default VPCs as non-compliant if they aren't configured with best-practices. Rather 
	than jumping through hoops, it's easier to delete to default VPCs. This task cannot be accomplished with terraform, 
	so this command is necessary.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return aws.DeleteDefaultVPCs(region, role, shouldDelete)
	},
}

func init() {
	awsCmd.AddCommand(deleteDefaultVPCsCmd)

	deleteDefaultVPCsCmd.Flags().BoolVarP(&shouldDelete, shouldDeleteFlag, "", false, "Flag to indicate if the delete should be run")
}
