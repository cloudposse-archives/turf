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
	"github.com/aws/aws-sdk-go/service/organizations"
	common "github.com/cloudposse/posse-cli/common/error"
)

func getOrgClient(role string) *organizations.Organizations {
	sess := GetSession()
	creds := GetCreds(sess, role)
	return organizations.New(sess, &aws.Config{Credentials: creds})
}

// ListMemberAccountIDs provides a list of AWS Accounts that are members of the AWS Organization
func ListMemberAccountIDs(role string) []string {
	client := getOrgClient(role)
	accounts, err := client.ListAccounts(&organizations.ListAccountsInput{})
	common.AssertErrorNil(err)

	accountIDs := make([]string, 0)
	for i := range accounts.Accounts {
		accountIDs = append(accountIDs, *accounts.Accounts[i].Id)
	}

	return accountIDs
}
