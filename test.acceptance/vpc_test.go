package test_acceptance

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
					testMainTags(argsDryRun, testAccVpcAWSweeperTagsConfig),
					testVpcExists(&vpc1),
					testVpcExists(&vpc2),
					testMainTags(argsForceDelete, testAccVpcAWSweeperTagsConfig),
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
					testMainVpcIds(argsDryRun, &vpc1),
					testVpcExists(&vpc1),
					testVpcExists(&vpc2),
					testMainVpcIds(argsForceDelete, &vpc1),
					testVpcDeleted(&vpc1),
					testVpcExists(&vpc2),
				),
			},
		},
	})
}

func testAccCheckVpcExists(n string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := client.ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
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

func testMainVpcIds(args []string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		command_wipe.OsFs = afero.NewMemMapFs()
		afero.WriteFile(command_wipe.OsFs, "config.yml", []byte(testAccVpcAWSweeperIdsConfig(vpc)), 0644)
		os.Args = args

		command_wipe.WrappedMain()
		return nil
	}
}

func testVpcExists(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{vpc.VpcId},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
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
		conn := client.ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{vpc.VpcId},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
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

const testAccVpcAWSweeperTagsConfig = `
aws_vpc:
  tags:
    foo: bar
`

func testAccVpcAWSweeperIdsConfig(vpc *ec2.Vpc) string {
	id := vpc.VpcId
	return fmt.Sprintf(`
aws_vpc:
  ids:
    - %s
`, *id)
}
