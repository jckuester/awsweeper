package test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccCLIArguments_region_UsEast1(t *testing.T) {
	var vpc ec2.Vpc

	region := "us-east-1"
	awsClient, tfAwsProvider := initTests(&region)

	argsRegionUsEast1 := []string{"cmd", "--force", "--region", region, "config.yml"}

	resource.Test(t, resource.TestCase{
		Providers: tfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccVpcConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					awsClient.testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testMainVpcIds(argsRegionUsEast1, &vpc),
					awsClient.testVpcDeleted(&vpc),
				),
			},
		},
	})
}

func TestAccCLIArguments_region_UsWest2(t *testing.T) {
	var vpc ec2.Vpc

	region := "us-west-2"
	awsClient, tfAwsProvider := initTests(&region)

	argsRegionUsWest2 := []string{"cmd", "--force", "--region", region, "config.yml"}

	resource.Test(t, resource.TestCase{
		Providers: tfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccVpcConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					awsClient.testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testMainVpcIds(argsRegionUsWest2, &vpc),
					awsClient.testVpcDeleted(&vpc),
				),
			},
		},
	})
}
