package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudetc/awsweeper/command"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

func TestAccSubnet_deleteByTags(t *testing.T) {
	var subnet1, subnet2 ec2.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccSubnetConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_subnet.foo", &subnet1),
					testAccCheckSubnetExists("aws_subnet.bar", &subnet2),
					testMainTags(argsDryRun, testAWSweeperTagsConfig(res.Subnet)),
					testSubnetExists(&subnet1),
					testSubnetExists(&subnet2),
					testMainTags(argsForceDelete, testAWSweeperTagsConfig(res.Subnet)),
					testSubnetDeleted(&subnet1),
					testSubnetExists(&subnet2),
				),
			},
		},
	})
}

func TestAccSubnet_deleteByIds(t *testing.T) {
	var subnet1, subnet2 ec2.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccSubnetConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_subnet.foo", &subnet1),
					testAccCheckSubnetExists("aws_subnet.bar", &subnet2),
					testMainSubnetIds(argsDryRun, &subnet1),
					testSubnetExists(&subnet1),
					testSubnetExists(&subnet2),
					testMainSubnetIds(argsForceDelete, &subnet1),
					testSubnetDeleted(&subnet1),
					testSubnetExists(&subnet2),
				),
			},
		},
	})
}

func testAccCheckSubnetExists(n string, subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no subnet ID is set")
		}

		conn := client.EC2API
		DescribeSubnetOpts := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeSubnets(DescribeSubnetOpts)
		if err != nil {
			return err
		}
		if len(resp.Subnets) == 0 {
			return fmt.Errorf("subnet not found")
		}

		*subnet = *resp.Subnets[0]

		return nil
	}
}

func testSubnetExists(subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.EC2API
		DescribeSubnetOpts := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{subnet.SubnetId},
		}
		resp, err := conn.DescribeSubnets(DescribeSubnetOpts)
		if err != nil {
			return err
		}
		if len(resp.Subnets) == 0 {
			return fmt.Errorf("subnet has been deleted")
		}

		return nil
	}
}

func testSubnetDeleted(subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.EC2API
		DescribeSubnetOpts := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{subnet.SubnetId},
		}
		resp, err := conn.DescribeSubnets(DescribeSubnetOpts)
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() == "InvalidSubnetID.NotFound" {
				return nil
			}
			return err
		}

		if len(resp.Subnets) != 0 {
			return fmt.Errorf("subnet hasn't been deleted")
		}

		return nil
	}
}

func testMainSubnetIds(args []string, subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(testAWSweeperIdsConfig(res.Subnet, subnet.SubnetId)), 0644)
		os.Args = args

		command.WrappedMain()
		return nil
	}
}

const testAccSubnetConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		Name = "awsweeper-testacc"
	}
}

resource "aws_subnet" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
	cidr_block = "10.1.1.0/24"

	tags {
		foo = "bar"
		Name = "awsweeper-testacc"
	}
}

resource "aws_subnet" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
	cidr_block = "10.1.2.0/24"

	tags {
		bar = "baz"
		Name = "awsweeper-testacc"
	}
}
`
