package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudetc/awsweeper/command"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVpc_deleteByTags(t *testing.T) {
	var vpc1, vpc2 ec2.Vpc

	awsClient, tfAwsProvider := initTests(aws.String("us-west-2"))

	resource.Test(t, resource.TestCase{
		Providers: tfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccVpcConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					awsClient.testAccCheckVpcExists("aws_vpc.foo", &vpc1),
					awsClient.testAccCheckVpcExists("aws_vpc.bar", &vpc2),
					testMainTags(argsDryRun, testAWSweeperTagsConfig(res.Vpc)),
					awsClient.testVpcExists(&vpc1),
					awsClient.testVpcExists(&vpc2),
					testMainTags(argsForceDelete, testAWSweeperTagsConfig(res.Vpc)),
					awsClient.testVpcDeleted(&vpc1),
					awsClient.testVpcExists(&vpc2),
				),
			},
		},
	})
}

func TestAccVpc_deleteByIds(t *testing.T) {
	var vpc1, vpc2 ec2.Vpc

	awsClient, tfAwsProvider := initTests(aws.String("us-west-2"))

	resource.Test(t, resource.TestCase{
		Providers: tfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccVpcConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					awsClient.testAccCheckVpcExists("aws_vpc.foo", &vpc1),
					awsClient.testAccCheckVpcExists("aws_vpc.bar", &vpc2),
					testMainVpcIds(argsDryRun, &vpc1),
					awsClient.testVpcExists(&vpc1),
					awsClient.testVpcExists(&vpc2),
					testMainVpcIds(argsForceDelete, &vpc1),
					awsClient.testVpcDeleted(&vpc1),
					awsClient.testVpcExists(&vpc2),
				),
			},
		},
	})
}

func testMainVpcIds(args []string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(testAWSweeperIdsConfig(res.Vpc, vpc.VpcId)), 0644)
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func (a AWS) testAccCheckVpcExists(name string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		desc := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := a.DescribeVpcs(desc)
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

func (a AWS) testVpcExists(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		desc := &ec2.DescribeVpcsInput{
			VpcIds: []*string{vpc.VpcId},
		}
		resp, err := a.DescribeVpcs(desc)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 {
			return fmt.Errorf("VPC has been deleted")
		}

		return nil
	}
}

func (a AWS) testVpcDeleted(vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		desc := &ec2.DescribeVpcsInput{
			VpcIds: []*string{vpc.VpcId},
		}
		resp, err := a.DescribeVpcs(desc)
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
