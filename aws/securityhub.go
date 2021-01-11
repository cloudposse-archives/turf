package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/securityhub"
	common "github.com/cloudposse/posse-cli/common/error"
	"github.com/sirupsen/logrus"
)

func getSecurityHubClient(region string, role string) (*securityhub.SecurityHub, *session.Session) {
	sess := GetSession()
	creds := GetCreds(sess, role)
	securityHubClient := securityhub.New(sess, &aws.Config{Credentials: creds, Region: &region})

	return securityHubClient, sess
}

func enableAdminAccount(client *securityhub.SecurityHub, accountID string) {
	updateInput := securityhub.EnableOrganizationAdminAccountInput{AdminAccountId: &accountID}
	client.EnableOrganizationAdminAccount(&updateInput)
}

func enableAutoEnable(client *securityhub.SecurityHub) {
	updateInput := securityhub.UpdateOrganizationConfigurationInput{AutoEnable: aws.Bool(true)}
	client.UpdateOrganizationConfiguration(&updateInput)
}

func isAdministratorAccountEnabled() bool {
	return false
}

func containsAdminAccount(s []*securityhub.AdminAccount, e string) bool {
	for _, a := range s {
		if *a.AccountId == e {
			return true
		}
	}
	return false
}

func adminAccountAlreadyEnabled(client *securityhub.SecurityHub, accountID string) bool {
	listInput := securityhub.ListOrganizationAdminAccountsInput{}
	orgConfig, err := client.ListOrganizationAdminAccounts(&listInput)
	common.AssertErrorNil(err)
	if containsAdminAccount(orgConfig.AdminAccounts, accountID) {
		return true
	}
	return false
}

// EnableAdministratorAccount enables the Security Hub Administrator account within the AWS Organization
func EnableAdministratorAccount(region string, administratorAccountRole string, rootRole string) error {
	rootAccountClient, rootSession := getSecurityHubClient(region, rootRole)
	adminAccountClient, adminSession := getSecurityHubClient(region, administratorAccountRole)
	rootAccountID := GetAccountID(rootSession, rootRole)
	adminAccountID := GetAccountID(adminSession, administratorAccountRole)

	logrus.Info("Enabling organization-wide AWS Security Hub with the following config:")
	logrus.Infof("    AWS Management Account %s", rootAccountID)
	logrus.Infof("    AWS Security Hub Administrator Account %s", adminAccountID)

	if !adminAccountAlreadyEnabled(rootAccountClient, adminAccountID) {
		enableAdminAccount(rootAccountClient, adminAccountID)
		enableAutoEnable(adminAccountClient)
	} else {
		logrus.Infof("%s already set as AWS Security Hub Administrator Account, skipping", adminAccountID)
	}

	logrus.Infof("Organization-wide AWS Security Hub complete")

	return nil
}

// AddMemberAccounts adds all the accounts in the AWS Organization as members to the Security Hub Administrator Account
func AddMemberAccounts(region string, administratorAccountRole string, rootRole string) {

}
