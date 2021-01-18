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
	"github.com/aws/aws-sdk-go/service/guardduty"
	common "github.com/cloudposse/posse-cli/common/error"
	"github.com/sirupsen/logrus"
)

func getGuardDutyClient(region string, role string) *guardduty.GuardDuty {
	sess := GetSession()
	creds := GetCreds(sess, role)
	guardDutyClient := guardduty.New(sess, &aws.Config{Credentials: creds, Region: &region})

	return guardDutyClient
}

func enableGuardDutyAdminAccount(client *guardduty.GuardDuty, accountID string) {
	updateInput := guardduty.EnableOrganizationAdminAccountInput{AdminAccountId: &accountID}
	client.EnableOrganizationAdminAccount(&updateInput)
}

// We need to enable GuardDuty in the AWS Organizations Management Account so that it can be added as a member
// account in AWS GuardDuty's Administrator account. Accounts other than the Management Account don't need to be
// excplicitly enabled, but the MA does.
func enableGuardDutyInManagementAccount(client *guardduty.GuardDuty) {
	_, err := client.CreateDetector(&guardduty.CreateDetectorInput{Enable: aws.Bool(true)})
	if err != nil {
		logrus.Error(err)
	}
}

func isGuardDutyAdministratorAccountEnabled() bool {
	return false
}

func containsGuardDutyAdminAccount(s []*guardduty.AdminAccount, e string) bool {
	for _, a := range s {
		if *a.AdminAccountId == e {
			return true
		}
	}
	return false
}

func guardDutyAdminAccountAlreadyEnabled(client *guardduty.GuardDuty, accountID string) bool {
	listInput := guardduty.ListOrganizationAdminAccountsInput{}
	orgConfig, err := client.ListOrganizationAdminAccounts(&listInput)
	common.AssertErrorNil(err)
	if containsGuardDutyAdminAccount(orgConfig.AdminAccounts, accountID) {
		return true
	}
	return false
}

func logGuardDutyMemberAccounts(memberAccounts []AccountWithEmail) {
	logrus.Info("  AWS GuardDuty Member accounts:")

	for i := range memberAccounts {
		logrus.Infof("    %s (%s)", memberAccounts[i].AccountID, memberAccounts[i].Email)
	}
}

// addMemberAccount adds an account in the AWS Organization as a member of the GuardDuty Administrator Account
func addGuardDutyMemberAccounts(client *guardduty.GuardDuty, detectorID string, memberAccounts []AccountWithEmail, administratorAcctID string) {
	accountDetails := make([]*guardduty.AccountDetail, 0)
	for i := range memberAccounts {
		currentAccountID := memberAccounts[i].AccountID
		currentEmailAddress := memberAccounts[i].Email
		if currentAccountID != administratorAcctID {
			accountDetails = append(accountDetails, &guardduty.AccountDetail{AccountId: &currentAccountID, Email: aws.String(currentEmailAddress)})
		}
	}
	input := guardduty.CreateMembersInput{AccountDetails: accountDetails, DetectorId: aws.String(detectorID)}
	result, err := client.CreateMembers(&input)
	if err != nil {
		logrus.Error(err)
	}

	if len(result.UnprocessedAccounts) > 0 {
		logrus.Error(result)
	}
}

func getDetectorIDForRegion(client *guardduty.GuardDuty) string {
	detectors, err := client.ListDetectors(&guardduty.ListDetectorsInput{})
	if err != nil {
		logrus.Error(err)
	}

	if len(detectors.DetectorIds) == 0 {
		logrus.Error("    No GuardDuty Detectors Found!")
		return ""
	}

	return *detectors.DetectorIds[0]
}

func enableGuardDutyAutoEnable(client *guardduty.GuardDuty, autoEnableS3Protection bool) {
	logrus.Info("    Enabling GuardDuty Auto-Enable for new AWS Organization Member Accounts")
	detector := getDetectorIDForRegion(client)

	updateInput := guardduty.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(true),
		DetectorId: aws.String(detector),
		DataSources: &guardduty.OrganizationDataSourceConfigurations{
			S3Logs: &guardduty.OrganizationS3LogsConfiguration{
				AutoEnable: aws.Bool(autoEnableS3Protection),
			},
		},
	}

	client.UpdateOrganizationConfiguration(&updateInput)
}

// EnableGuardDutyAdministratorAccount enables the GuardDuty Administrator account within the AWS Organization
func EnableGuardDutyAdministratorAccount(region string, administratorAccountRole string, rootRole string, autoEnableS3Protection bool) error {
	rootSession := GetSession()
	rootAccountID := GetAccountID(rootSession, rootRole)

	adminAcctSession := GetSession()
	adminAccountID := GetAccountID(adminAcctSession, administratorAccountRole)

	enabledRegions := GetEnabledRegions(region, rootRole)

	logrus.Info("Enabling organization-wide AWS GuardDuty with the following config:")
	logrus.Infof("  AWS Management Account %s", rootAccountID)
	logrus.Infof("  AWS GuardDuty Administrator Account %s", adminAccountID)

	memberAccounts := ListMemberAccountIDsWithEmails(rootRole)
	logGuardDutyMemberAccounts(memberAccounts)

	for r := range enabledRegions {
		currentRegion := enabledRegions[r]
		logrus.Infof("  Processing region %s", currentRegion)

		rootAccountClient := getGuardDutyClient(currentRegion, rootRole)
		adminAccountClient := getGuardDutyClient(currentRegion, administratorAccountRole)

		detectorID := getDetectorIDForRegion(adminAccountClient)

		if !guardDutyAdminAccountAlreadyEnabled(rootAccountClient, adminAccountID) {
			enableGuardDutyAdminAccount(rootAccountClient, adminAccountID)
			detectorID = getDetectorIDForRegion(adminAccountClient)

			enableGuardDutyInManagementAccount(rootAccountClient)

		} else {
			logrus.Infof("    Account %s is already set as AWS GuardDuty Administrator Account, skipping configuration", adminAccountID)
		}
		enableGuardDutyAutoEnable(adminAccountClient, autoEnableS3Protection)
		addGuardDutyMemberAccounts(adminAccountClient, detectorID, memberAccounts, adminAccountID)
	}
	logrus.Infof("Organization-wide AWS GuardDuty complete")

	return nil
}
