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

func TestAccInternetGateways_deleteByTags(t *testing.T) {
	t.SkipNow()
	// TODO tag support

	var ig1, ig2 ec2.InternetGateway

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccInternetGatewayConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists("aws_internet_gateway.foo", &ig1),
					testAccCheckInternetGatewayExists("aws_internet_gateway.bar", &ig2),
					testMainTags(argsDryRun, testAWSweeperTagsConfig(res.InternetGateway)),
					testInternetGatewayExists(&ig1),
					testInternetGatewayExists(&ig2),
					testMainTags(argsForceDelete, testAWSweeperTagsConfig(res.InternetGateway)),
					testInternetGatewayDeleted(&ig1),
				),
			},
		},
	})
}

func TestAccInternetGateway_deleteByIds(t *testing.T) {
	var ig1, ig2 ec2.InternetGateway

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccInternetGatewayConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists("aws_internet_gateway.foo", &ig1),
					testAccCheckInternetGatewayExists("aws_internet_gateway.bar", &ig2),
					testMainInternetGatewayIds(argsDryRun, &ig1),
					testInternetGatewayExists(&ig1),
					testInternetGatewayExists(&ig2),
					testMainInternetGatewayIds(argsForceDelete, &ig1),
					testInternetGatewayDeleted(&ig1),
					testInternetGatewayExists(&ig2),
				),
			},
		},
	})
}

func testMainInternetGatewayIds(args []string, ig *ec2.InternetGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml",
			[]byte(testAWSweeperIdsConfig(res.InternetGateway, ig.InternetGatewayId)), 0644)
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckInternetGatewayExists(n string, ig *ec2.InternetGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := client.EC2API
		resp, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.InternetGateways) == 0 {
			return fmt.Errorf("InternetGateway not found")
		}

		*ig = *resp.InternetGateways[0]

		return nil
	}
}

func testInternetGatewayExists(ig *ec2.InternetGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.EC2API
		opts := &ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{ig.InternetGatewayId},
		}
		desc, err := conn.DescribeInternetGateways(opts)
		if err != nil {
			return err
		}

		if len(desc.InternetGateways) == 0 {
			return fmt.Errorf("InternetGateway has been deleted")
		}

		return nil
	}
}

func testInternetGatewayDeleted(ig *ec2.InternetGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.EC2API
		desc := &ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{ig.InternetGatewayId},
		}
		resp, err := conn.DescribeInternetGateways(desc)
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() == "InvalidInternetGatewayID.NotFound" {
				return nil
			}
			return err
		}

		if len(resp.InternetGateways) != 0 {
			return fmt.Errorf("InternetGateway hasn't been deleted")
		}

		return nil
	}
}

const testAccInternetGatewayConfig = `
resource "aws_internet_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"

	tags {
		key = "foo"
		value = "bar"
	}

	tags {
		key = "Name"
		value = "awsweeper-testacc"
	}
}

resource "aws_internet_gateway" "bar" {
  vpc_id = "${aws_vpc.bar.id}"

	tags {
		key = "foo"
		value = "baz"
	}

	tags {
		key = "Name"
		value = "awsweeper-testacc"
	}
}

resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		Name = "awsweeper-testacc"
	}
}

resource "aws_vpc" "bar" {
	cidr_block = "10.2.0.0/16"

	tags {
		Name = "awsweeper-testacc"
	}
}
`
