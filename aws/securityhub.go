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
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/securityhub"
	common "github.com/cloudposse/turf/common/error"
	"github.com/sirupsen/logrus"
)

// SecurityHub is a struct that represents an AWS Security Hub and attaches methods to perform various operations against
// it
type SecurityHub struct {
	adminAccountClient      *securityhub.SecurityHub
	currentAccountClient    *securityhub.SecurityHub
	managementAccountClient *securityhub.SecurityHub
}

func (hub SecurityHub) securityHubAdminAccountAlreadyEnabled(accountID string) bool {
	listInput := securityhub.ListOrganizationAdminAccountsInput{}
	orgConfig, err := hub.managementAccountClient.ListOrganizationAdminAccounts(&listInput)
	common.AssertErrorNil(err)
	if containsSecurityHubAdminAccount(orgConfig.AdminAccounts, accountID) {
		return true
	}
	return false
}

func (hub SecurityHub) enableSecurityHubAdminAccount(accountID string) {
	updateInput := securityhub.EnableOrganizationAdminAccountInput{AdminAccountId: &accountID}
	hub.managementAccountClient.EnableOrganizationAdminAccount(&updateInput)
}

func (hub SecurityHub) enableSecurityHubAutoEnable() {
	logrus.Info("    Setting Security Hub Auto-Enable for new AWS Organization Member Accounts")
	updateInput := securityhub.UpdateOrganizationConfigurationInput{AutoEnable: aws.Bool(true)}
	hub.adminAccountClient.UpdateOrganizationConfiguration(&updateInput)
}

// We need to enable Security Hub in the AWS Organizations Management Account so that it can be added as a member
// account in AWS Security Hub's Administrator account. Accounts other than the Management Account don't need to be
// excplicitly enabled, but the MA does.
func (hub SecurityHub) enableSecurityHubInManagementAccount() {
	_, err := hub.managementAccountClient.EnableSecurityHub(&securityhub.EnableSecurityHubInput{})
	if err != nil {
		logrus.Error(err)
	}
}

// addMemberAccount adds an account in the AWS Organization as a member of the Security Hub Administrator Account
func (hub SecurityHub) addSecurityHubMemberAccounts(memberAccounts []string, administratorAcctID string) {
	accountDetails := make([]*securityhub.AccountDetails, 0)
	for i := range memberAccounts {
		currentAccountID := memberAccounts[i]
		if currentAccountID != administratorAcctID {
			accountDetails = append(accountDetails, &securityhub.AccountDetails{AccountId: &currentAccountID})
		}
	}
	input := securityhub.CreateMembersInput{AccountDetails: accountDetails}
	result, err := hub.adminAccountClient.CreateMembers(&input)
	if err != nil {
		logrus.Error(err)
	}

	if len(result.UnprocessedAccounts) > 0 {
		logrus.Error(result)
	}
}

func (hub SecurityHub) disableControl(currentControl string) error {
	_, err := hub.currentAccountClient.UpdateStandardsControl(&securityhub.UpdateStandardsControlInput{
		ControlStatus:       aws.String("DISABLED"),
		DisabledReason:      aws.String("Global Resources are not collected in this region"),
		StandardsControlArn: aws.String(currentControl),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "TooManyRequestsException" {
				logrus.Warnf("    received too many requests error. Sleeping then trying again while disabling control %s", currentControl)

				time.Sleep(2 * time.Second)
				return hub.disableControl(currentControl)
			}
		}
	}

	return err
}

func (hub SecurityHub) disableControls(region string, accountID string, controls []string) {
	for i := range controls {
		currentControl := fmt.Sprintf(controls[i], region, accountID)

		logrus.Infof("    disabling control %s", currentControl)
		err := hub.disableControl(currentControl)

		if err != nil {
			logrus.Error(err)
		}
	}
}

// Controls for AWS Foundational Security Best Practices v1.0.0
func getFoundations100Controls(isGlobalCollectionRegion bool) []string {
	controls := []string{}

	if !isGlobalCollectionRegion {
		controls = append(controls, []string{
			"arn:aws:securityhub:%s:%s:control/aws-foundational-security-best-practices/v/1.0.0/Config.1",
			"arn:aws:securityhub:%s:%s:control/aws-foundational-security-best-practices/v/1.0.0/IAM.1",
			"arn:aws:securityhub:%s:%s:control/aws-foundational-security-best-practices/v/1.0.0/IAM.2",
			"arn:aws:securityhub:%s:%s:control/aws-foundational-security-best-practices/v/1.0.0/IAM.3",
			"arn:aws:securityhub:%s:%s:control/aws-foundational-security-best-practices/v/1.0.0/IAM.4",
			"arn:aws:securityhub:%s:%s:control/aws-foundational-security-best-practices/v/1.0.0/IAM.5",
			"arn:aws:securityhub:%s:%s:control/aws-foundational-security-best-practices/v/1.0.0/IAM.6",
			"arn:aws:securityhub:%s:%s:control/aws-foundational-security-best-practices/v/1.0.0/IAM.7",
		}...)
	}

	return controls
}

// Controls for CIS AWS Foundations Benchmark v1.2.0
func getCIS120Controls(isGlobalCollectionRegion bool, isCloudTrailAccount bool) []string {
	controls := []string{}

	if !isGlobalCollectionRegion {
		controls = append(controls, []string{
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.2",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.3",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.4",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.5",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.6",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.7",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.8",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.9",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.10",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.11",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.12",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.13",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.14",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.16",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.20",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/1.22",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/2.5",
		}...)
	}

	if !isCloudTrailAccount || (isCloudTrailAccount && !isGlobalCollectionRegion) {
		controls = append(controls, []string{
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/2.7",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.1",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.2",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.3",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.4",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.5",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.6",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.7",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.8",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.9",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.10",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.11",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.12",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.13",
			"arn:aws:securityhub:%s:%s:control/cis-aws-foundations-benchmark/v/1.2.0/3.14",
		}...)
	}

	return controls
}

func getSecurityHubClient(region string) *securityhub.SecurityHub {
	sess := GetSession()
	securityHubClient := securityhub.New(sess, &aws.Config{Region: &region})

	return securityHubClient
}

func getSecurityHubClientWithRole(region string, role string) *securityhub.SecurityHub {
	sess := GetSession()
	creds := GetCreds(sess, role)
	securityHubClient := securityhub.New(sess, &aws.Config{Credentials: creds, Region: &region})

	return securityHubClient
}

func containsSecurityHubAdminAccount(s []*securityhub.AdminAccount, e string) bool {
	for _, a := range s {
		if *a.AccountId == e {
			return true
		}
	}
	return false
}

func logSecurityHubMemberAccounts(memberAccounts []string) {
	logrus.Info("  AWS Security Hub Member accounts:")

	for i := range memberAccounts {
		logrus.Infof("    %s", memberAccounts[i])
	}
}

// EnableSecurityHubAdministratorAccount enables the Security Hub Administrator account within the AWS Organization
func EnableSecurityHubAdministratorAccount(region string, administratorAccountRole string, rootRole string) error {
	rootSession := GetSession()
	rootAccountID := GetAccountIDWithRole(rootSession, rootRole)

	adminAcctSession := GetSession()
	adminAccountID := GetAccountIDWithRole(adminAcctSession, administratorAccountRole)

	enabledRegions := GetEnabledRegions(region, rootRole, false)

	logrus.Info("Enabling organization-wide AWS Security Hub with the following config:")
	logrus.Infof("  AWS Management Account %s", rootAccountID)
	logrus.Infof("  AWS Security Hub Administrator Account %s", adminAccountID)

	memberAccounts := ListMemberAccountIDs(rootRole)
	logSecurityHubMemberAccounts(memberAccounts)

	for r := range enabledRegions {
		currentRegion := enabledRegions[r]
		logrus.Infof("  Processing region %s", currentRegion)

		managementAccountClient := getSecurityHubClientWithRole(currentRegion, rootRole)
		adminAccountClient := getSecurityHubClientWithRole(currentRegion, administratorAccountRole)

		hub := SecurityHub{
			adminAccountClient:      adminAccountClient,
			managementAccountClient: managementAccountClient,
		}

		if !hub.securityHubAdminAccountAlreadyEnabled(adminAccountID) {
			hub.enableSecurityHubAdminAccount(adminAccountID)
			hub.enableSecurityHubAutoEnable()
			hub.enableSecurityHubInManagementAccount()
		} else {
			logrus.Infof("    Account %s is already set as AWS Security Hub Administrator Account, skipping configuration", adminAccountID)
		}

		hub.addSecurityHubMemberAccounts(memberAccounts, adminAccountID)
	}
	logrus.Infof("Organization-wide AWS Security Hub complete")

	return nil
}

func validateRegion(enabledRegions []string, region string) bool {
	for i := range enabledRegions {
		if enabledRegions[i] == region {
			return true
		}
	}
	return false
}

// DisableSecurityHubGlobalResourceControls disables Security Hub controls related to Global Resources in regions that
// aren't collecting Global Resources. It also disables CloudTrail related controls in accounts that aren't the central
// CloudTrail account.
//
// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-standards-cis-to-disable.html
// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-standards-fsbp-to-disable.html
func DisableSecurityHubGlobalResourceControls(globalCollectionRegion string, role string, isPrivileged bool, isCloudTrailAccount bool) error {
	if role == "" && !isPrivileged {
		return errors.New("Either role must be provided or the privileged flag must be set")
	}

	session := GetSession()
	var accountID string

	if isPrivileged {
		accountID = GetAccountID(session)
	} else {
		accountID = GetAccountIDWithRole(session, role)
	}

	enabledRegions := GetEnabledRegions("us-east-1", role, isPrivileged)

	if !validateRegion(enabledRegions, globalCollectionRegion) {
		return fmt.Errorf("%s is not a valid enabled region in this account", globalCollectionRegion)
	}

	logrus.Infof("Disabling Global Resource controls for all regions excluding %s for account %s", globalCollectionRegion, accountID)

	for r := range enabledRegions {
		currentRegion := enabledRegions[r]

		var currentAccountClient *securityhub.SecurityHub

		if isPrivileged {
			currentAccountClient = getSecurityHubClient(currentRegion)
		} else {
			currentAccountClient = getSecurityHubClientWithRole(currentRegion, role)
		}

		hub := SecurityHub{
			currentAccountClient: currentAccountClient,
		}

		isGlobalCollectionRegion := currentRegion == globalCollectionRegion

		if isGlobalCollectionRegion {
			logrus.Infof("  processing global collection region %s", currentRegion)
		} else {
			logrus.Infof("  processing region %s", currentRegion)
		}

		foundations100Controls := getFoundations100Controls(isGlobalCollectionRegion)
		cis120Controls := getCIS120Controls(isGlobalCollectionRegion, isCloudTrailAccount)

		hub.disableControls(currentRegion, accountID, foundations100Controls)
		hub.disableControls(currentRegion, accountID, cis120Controls)
	}

	return nil
}
