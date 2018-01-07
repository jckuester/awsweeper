package test_integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudetc/awsweeper/command_wipe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

func TestSubnet_tags(t *testing.T) {
	var subnet ec2.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccSubnetConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_subnet.foo", &subnet),
					testMainTags(argsDryRun, testAccSubnetAWSweeperTagsConfig),
					testSubnetExists(&subnet),
					testMainTags(argsForceDelete, testAccSubnetAWSweeperTagsConfig),
					testSubnetDeleted(&subnet),
				),
			},
		},
	})
}

func TestSubnet_ids(t *testing.T) {
	var subnet ec2.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccSubnetConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_subnet.foo", &subnet),
					testMainSubnetIds(argsDryRun, &subnet),
					testSubnetExists(&subnet),
					testMainSubnetIds(argsForceDelete, &subnet),
					testSubnetDeleted(&subnet),
				),
			},
		},
	})
}

func testAccCheckSubnetExists(n string, subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No subnet ID is set")
		}

		conn := client.ec2conn
		DescribeSubnetOpts := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeSubnets(DescribeSubnetOpts)
		if err != nil {
			return err
		}
		if len(resp.Subnets) == 0 {
			return fmt.Errorf("Subnet not found")
		}

		*subnet = *resp.Subnets[0]

		return nil
	}
}

func testSubnetExists(subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.ec2conn
		DescribeSubnetOpts := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{subnet.SubnetId},
		}
		resp, err := conn.DescribeSubnets(DescribeSubnetOpts)
		if err != nil {
			return err
		}
		if len(resp.Subnets) == 0 {
			return fmt.Errorf("Subnet has been deleted")
		}

		return nil
	}
}

func testSubnetDeleted(subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.ec2conn
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
			return fmt.Errorf("Subnet hasn't been deleted")
		}

		return nil
	}
}

func testMainSubnetIds(args []string, subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		command_wipe.OsFs = afero.NewMemMapFs()
		afero.WriteFile(command_wipe.OsFs, "config.yml", []byte(testAccSubnetAWSweeperIdsConfig(subnet)), 0644)
		os.Args = args

		command_wipe.WrappedMain()
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
	cidr_block = "10.1.0.1/24"

	tags {
		foo = "bar"
		Name = "awsweeper-testacc"
	}
}
`

const testAccSubnetAWSweeperTagsConfig = `
aws_subnet:
  tags:
    foo: bar
`

func testAccSubnetAWSweeperIdsConfig(subnet *ec2.Subnet) string {
	id := subnet.SubnetId
	return fmt.Sprintf(`
aws_subnet:
  ids:
    - %s
`, *id)
}
