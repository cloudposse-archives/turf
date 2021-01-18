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
	"github.com/aws/aws-sdk-go/service/ec2"
	common "github.com/cloudposse/posse-cli/common/error"
	"github.com/sirupsen/logrus"
)

func getEC2Client(region string, role string) *ec2.EC2 {
	sess := GetSession()
	creds := GetCreds(sess, role)
	return ec2.New(sess, &aws.Config{Credentials: creds, Region: &region})
}

func getDefaultVPC(client *ec2.EC2) string {
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("isDefault"),
			Values: []*string{aws.String("true")},
		},
	}
	describeInput := &ec2.DescribeVpcsInput{Filters: filters}
	defaultVpc, err := client.DescribeVpcs(describeInput)
	common.AssertErrorNil(err)

	if len(defaultVpc.Vpcs) == 0 {
		logrus.Info("      no default VPC found")
		return ""
	}
	return *defaultVpc.Vpcs[0].VpcId
}

// GetEnabledRegions provides a list of AWS Regions that are enabled
func GetEnabledRegions(region string, role string) []string {
	client := getEC2Client(region, role)
	regions, err := client.DescribeRegions(&ec2.DescribeRegionsInput{AllRegions: aws.Bool(false)})
	common.AssertErrorNil(err)

	regionsList := make([]string, 0)
	for i := range regions.Regions {
		regionsList = append(regionsList, *regions.Regions[i].RegionName)
	}

	return regionsList
}
