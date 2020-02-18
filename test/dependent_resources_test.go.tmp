package test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAcc_deleteDependentResourcesFirst(t *testing.T) {
	var vpc ec2.Vpc
	var subnet ec2.Subnet

	awsClient, tfAwsProvider := initTests(aws.String("us-west-2"))

	resource.Test(t, resource.TestCase{
		Providers: tfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccDependentResources,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					awsClient.testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckSubnetExists("aws_subnet.bar", &subnet),
					testMainTags(argsForceDelete, testAccAwsweeperConfig),
					awsClient.testVpcDeleted(&vpc),
					testSubnetDeleted(&subnet),
				),
			},
		},
	})
}

const testAccDependentResources = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		foo = "bar"
		Name = "awsweeper-testacc"
	}
}

resource "aws_subnet" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
	cidr_block = "10.1.1.0/24"

	tags {
		foo = "bar"
		Name = "awsweeper-testacc"
	}
}
`

const testAccAwsweeperConfig = `
aws_vpc:
  - tags:
      foo: bar
aws_subnet:
  - tags:
      foo: bar
`
