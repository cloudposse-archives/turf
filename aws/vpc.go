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
	"github.com/sirupsen/logrus"
)

// Vpc is a struct that represents an AWS VPC and attaches methods to delete subordanate resources
type Vpc struct {
	VpcID  string
	client ec2.EC2
}

func (vpc Vpc) deleteInternetGateways() {
	gws, err := vpc.client.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: []*string{aws.String(vpc.VpcID)},
			},
		},
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	if len(gws.InternetGateways) == 1 {
		logrus.Infof("      deleting internet gateways for %s", vpc.VpcID)
		for _, gw := range gws.InternetGateways {
			_, err := vpc.client.DetachInternetGateway(&ec2.DetachInternetGatewayInput{InternetGatewayId: gw.InternetGatewayId, VpcId: &vpc.VpcID})
			if err != nil {
				logrus.Error(err)
			} else {
				_, err := vpc.client.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{InternetGatewayId: gw.InternetGatewayId})
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	} else {
		logrus.Infof("      no internet gateways found for %s", vpc.VpcID)
	}
}

func (vpc Vpc) deleteSubnets() {
	subnets, err := vpc.client.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpc.VpcID)},
			},
		},
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	if len(subnets.Subnets) > 0 {
		logrus.Infof("      deleting subnets for %s", vpc.VpcID)
		for _, subnet := range subnets.Subnets {
			_, err := vpc.client.DeleteSubnet(&ec2.DeleteSubnetInput{SubnetId: subnet.SubnetId})
			if err != nil {
				logrus.Error(err)
			}
		}
	} else {
		logrus.Infof("      no subnets found for %s", vpc.VpcID)
	}
}

func (vpc Vpc) deleteRouteTables() {
	routeTables, err := vpc.client.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpc.VpcID)},
			},
		},
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	if len(routeTables.RouteTables) > 0 {
		logrus.Infof("      deleting route tables for %s", vpc.VpcID)

		for _, routeTable := range routeTables.RouteTables {
			if len(routeTable.Associations) > 0 && *routeTable.Associations[0].Main {
				continue
			}

			_, err := vpc.client.DeleteRouteTable(&ec2.DeleteRouteTableInput{RouteTableId: routeTable.RouteTableId})
			if err != nil {
				logrus.Error(err)
			}
		}
	} else {
		logrus.Infof("      no route tables found for %s", vpc.VpcID)
	}
}

func (vpc Vpc) deleteNACLs() {
	nacls, err := vpc.client.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpc.VpcID)},
			},
		},
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	if len(nacls.NetworkAcls) > 0 {
		logrus.Infof("      deleting nacls for %s", vpc.VpcID)
		for _, nacl := range nacls.NetworkAcls {
			if !*nacl.IsDefault {
				_, err := vpc.client.DeleteNetworkAcl(&ec2.DeleteNetworkAclInput{NetworkAclId: nacl.NetworkAclId})
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	} else {
		logrus.Infof("      no subnets found for %s", vpc.VpcID)
	}
}

func (vpc Vpc) deleteSecurityGroups() {
	sgs, err := vpc.client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpc.VpcID)},
			},
		},
	})
	if err != nil {
		logrus.Error(err)
		return
	}

	if len(sgs.SecurityGroups) > 0 {
		logrus.Infof("      deleting security groups for %s", vpc.VpcID)
		for _, sg := range sgs.SecurityGroups {
			if *sg.GroupName != "default" {
				_, err := vpc.client.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{GroupId: sg.GroupId})
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	} else {
		logrus.Infof("      no security groups found for %s", vpc.VpcID)
	}
}

func (vpc Vpc) deleteVpc() {
	logrus.Infof("      deleting %s", vpc.VpcID)
	_, err := vpc.client.DeleteVpc(&ec2.DeleteVpcInput{VpcId: &vpc.VpcID})
	if err != nil {
		logrus.Error(err)
	}
}

func (vpc Vpc) delete() {
	vpc.deleteInternetGateways()
	vpc.deleteSubnets()
	vpc.deleteRouteTables()
	vpc.deleteNACLs()
	vpc.deleteSecurityGroups()
	vpc.deleteVpc()
}

// DeleteDefaultVPCs deletes all of the default VPCs in all regions of an account
func DeleteDefaultVPCs(region string, role string, deleteFlag bool) error {
	enabledRegions := GetEnabledRegions(region, role)

	logrus.Infof("Deleting default VPCs")

	if !deleteFlag {
		logrus.Infof("Dry-run mode is active. Run again with %s flag to delete VPCs", "--delete")
	}

	logrus.Info("Identifying VPCs to delete:")

	for r := range enabledRegions {
		currentRegion := enabledRegions[r]
		logrus.Infof("  Processing region %s", currentRegion)

		client := getEC2Client(currentRegion, role)
		vpc := getDefaultVPC(client)
		if vpc != "" {
			logrus.Infof("    found %s", vpc)

			if deleteFlag {
				vpcInfo := Vpc{
					VpcID:  vpc,
					client: *client,
				}

				vpcInfo.delete()
			}
		}
	}

	logrus.Infof("Deleting default VPCs complete")

	return nil
}
