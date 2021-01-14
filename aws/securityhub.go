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

package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	common "github.com/cloudposse/posse-cli/common/error"
	"github.com/sirupsen/logrus"
)

func getSecurityHubClient(region string, role string) *securityhub.SecurityHub {
	sess := GetSession()
	creds := GetCreds(sess, role)
	securityHubClient := securityhub.New(sess, &aws.Config{Credentials: creds, Region: &region})

	return securityHubClient
}

func enableSecurityHubAdminAccount(client *securityhub.SecurityHub, accountID string) {
	updateInput := securityhub.EnableOrganizationAdminAccountInput{AdminAccountId: &accountID}
	client.EnableOrganizationAdminAccount(&updateInput)
}

// We need to enable Security Hub in the AWS Organizations Management Account so that it can be added as a member
// account in AWS Security Hub's Administrator account. Accounts other than the Management Account don't need to be
// excplicitly enabled, but the MA does.
func enableSecurityHubInManagementAccount(client *securityhub.SecurityHub) {
	_, err := client.EnableSecurityHub(&securityhub.EnableSecurityHubInput{})
	if err != nil {
		logrus.Error(err)
	}
}

func enableAutoEnable(client *securityhub.SecurityHub) {
	updateInput := securityhub.UpdateOrganizationConfigurationInput{AutoEnable: aws.Bool(true)}
	client.UpdateOrganizationConfiguration(&updateInput)
}

func isSecurityHubAdministratorAccountEnabled() bool {
	return false
}

func containsSecurityHubAdminAccount(s []*securityhub.AdminAccount, e string) bool {
	for _, a := range s {
		if *a.AccountId == e {
			return true
		}
	}
	return false
}

func securityHubAdminAccountAlreadyEnabled(client *securityhub.SecurityHub, accountID string) bool {
	listInput := securityhub.ListOrganizationAdminAccountsInput{}
	orgConfig, err := client.ListOrganizationAdminAccounts(&listInput)
	common.AssertErrorNil(err)
	if containsSecurityHubAdminAccount(orgConfig.AdminAccounts, accountID) {
		return true
	}
	return false
}

func logSecurityHubMemberAccounts(memberAccounts []string) {
	logrus.Info("  AWS Security Hub Member accounts:")

	for i := range memberAccounts {
		logrus.Infof("    %s", memberAccounts[i])
	}
}

// addMemberAccount adds an account in the AWS Organization as a member of the Security Hub Administrator Account
func addSecurityHubMemberAccounts(client *securityhub.SecurityHub, memberAccounts []string, administratorAcctID string) {
	accountDetails := make([]*securityhub.AccountDetails, 0)
	for i := range memberAccounts {
		currentAccountID := memberAccounts[i]
		if currentAccountID != administratorAcctID {
			accountDetails = append(accountDetails, &securityhub.AccountDetails{AccountId: &currentAccountID})
		}
	}
	input := securityhub.CreateMembersInput{AccountDetails: accountDetails}
	result, err := client.CreateMembers(&input)
	if err != nil {
		logrus.Error(err)
	}

	if len(result.UnprocessedAccounts) > 0 {
		logrus.Error(result)
	}
}

// EnableSecurityHubAdministratorAccount enables the Security Hub Administrator account within the AWS Organization
func EnableSecurityHubAdministratorAccount(region string, administratorAccountRole string, rootRole string) error {
	rootSession := GetSession()
	rootAccountID := GetAccountID(rootSession, rootRole)

	adminAcctSession := GetSession()
	adminAccountID := GetAccountID(adminAcctSession, administratorAccountRole)

	enabledRegions := GetEnabledRegions(rootRole)

	logrus.Info("Enabling organization-wide AWS Security Hub with the following config:")
	logrus.Infof("  AWS Management Account %s", rootAccountID)
	logrus.Infof("  AWS Security Hub Administrator Account %s", adminAccountID)

	memberAccounts := ListMemberAccountIDs(rootRole)
	logSecurityHubMemberAccounts(memberAccounts)

	for r := range enabledRegions {
		currentRegion := enabledRegions[r]
		logrus.Infof("  Processing region %s", currentRegion)

		rootAccountClient := getSecurityHubClient(currentRegion, rootRole)
		adminAccountClient := getSecurityHubClient(currentRegion, administratorAccountRole)

		if !securityHubAdminAccountAlreadyEnabled(rootAccountClient, adminAccountID) {
			enableSecurityHubAdminAccount(rootAccountClient, adminAccountID)
			enableAutoEnable(adminAccountClient)
			enableSecurityHubInManagementAccount(rootAccountClient)
		} else {
			logrus.Infof("    Account %s is already set as AWS Security Hub Administrator Account, skipping configuration", adminAccountID)
		}

		addSecurityHubMemberAccounts(adminAccountClient, memberAccounts, adminAccountID)
	}
	logrus.Infof("Organization-wide AWS Security Hub complete")

	return nil
}
