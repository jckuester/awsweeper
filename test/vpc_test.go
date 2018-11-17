package test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVpc_deleteByTags(t *testing.T) {
	var vpc1, vpc2 ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccVpcConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc1),
					testAccCheckVpcExists("aws_vpc.bar", &vpc2),
					testMainTags(argsDryRun, testAWSweeperTagsConfig(res.Vpc)),
					testVpcExists(&vpc1),
					testVpcExists(&vpc2),
					testMainTags(argsForceDelete, testAWSweeperTagsConfig(res.Vpc)),
					testVpcDeleted(&vpc1),
					testVpcExists(&vpc2),
				),
			},
		},
	})
}

func TestAccVpc_deleteByIds(t *testing.T) {
	var vpc1, vpc2 ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccVpcConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc1),
					testAccCheckVpcExists("aws_vpc.bar", &vpc2),
					testMainIds(argsDryRun, vpc1.VpcId),
					testVpcExists(&vpc1),
					testVpcExists(&vpc2),
					testMainIds(argsForceDelete, vpc1.VpcId),
					testVpcDeleted(&vpc1),
					testVpcExists(&vpc2),
				),
			},
		},
	})
}

func testAccCheckVpcExists(name string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := client.EC2API
		desc := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(desc)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 {
			return fmt.Errorf("VPC not found")
		}

		*vpc = *resp.Vpcs[0]

		return nil
	}
}

func testVpcExists(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.EC2API
		desc := &ec2.DescribeVpcsInput{
			VpcIds: []*string{vpc.VpcId},
		}
		resp, err := conn.DescribeVpcs(desc)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 {
			return fmt.Errorf("VPC has been deleted")
		}

		return nil
	}
}

func testVpcDeleted(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.EC2API
		desc := &ec2.DescribeVpcsInput{
			VpcIds: []*string{vpc.VpcId},
		}
		resp, err := conn.DescribeVpcs(desc)
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() == "InvalidVpcID.NotFound" {
				return nil
			}
			return err
		}

		if len(resp.Vpcs) != 0 {
			return fmt.Errorf("VPC hasn't been deleted")

		}

		return nil
	}
}

const testAccVpcConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		foo = "bar"
		Name = "awsweeper-testacc"
	}
}

resource "aws_vpc" "bar" {
	cidr_block = "10.2.0.0/16"

	tags {
		bar = "baz"
		Name = "awsweeper-testacc"
	}
}
`
