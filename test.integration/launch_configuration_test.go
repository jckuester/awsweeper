package test_integration

import (
	"fmt"
	"testing"

	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudetc/awsweeper/command_wipe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

func TestLaunchConfiguration_ids(t *testing.T) {
	var lc autoscaling.LaunchConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccLaunchConfigurationConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchConfigurationExists("aws_launch_configuration.foo", &lc),
					testMainLaunchConfigurationIds(argsDryRun, &lc),
					testLaunchConfigurationExists(&lc),
					testMainLaunchConfigurationIds(argsForceDelete, &lc),
					testLaunchConfigurationDeleted(&lc),
				),
			},
		},
	})
}

func testAccCheckLaunchConfigurationExists(n string, lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Launch Configuration name is set")
		}

		conn := client.autoscalingconn
		DescribeLaunchConfigurationOpts := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeLaunchConfigurations(DescribeLaunchConfigurationOpts)
		if err != nil {
			return err
		}
		if len(resp.LaunchConfigurations) == 0 {
			return fmt.Errorf("Launch Configuration not found")
		}

		*lc = *resp.LaunchConfigurations[0]

		return nil
	}
}

func testLaunchConfigurationDeleted(lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.autoscalingconn
		DescribeLaunchConfigurationOpts := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{lc.LaunchConfigurationName},
		}
		resp, err := conn.DescribeLaunchConfigurations(DescribeLaunchConfigurationOpts)
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() == "InvalidLaunchConfiguration.NotFound" {
				return nil
			}
			return err
		}

		if len(resp.LaunchConfigurations) != 0 {
			return fmt.Errorf("Launch Configuration hasn't been deleted")
		}

		return nil
	}
}

func testLaunchConfigurationExists(lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.autoscalingconn
		DescribeLaunchConfigurationOpts := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{lc.LaunchConfigurationName},
		}
		resp, err := conn.DescribeLaunchConfigurations(DescribeLaunchConfigurationOpts)
		if err != nil {
			return err
		}
		if len(resp.LaunchConfigurations) == 0 {
			return fmt.Errorf("Launch Configuration has been deleted")
		}

		return nil
	}
}

func testMainLaunchConfigurationIds(args []string, lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		command_wipe.OsFs = afero.NewMemMapFs()
		afero.WriteFile(command_wipe.OsFs, "config.yml", []byte(testAccLaunchConfigurationAWSweeperIdsConfig(lc)), 0644)
		os.Args = args

		command_wipe.WrappedMain()
		return nil
	}
}

const testAccLaunchConfigurationConfig = `
resource "aws_launch_configuration" "foo" {
	name_prefix = "awsweeper-testacc-"
	image_id = "${data.aws_ami.foo.id}"
	instance_type = "t2.micro"

	lifecycle {
		create_before_destroy = true
	}
}

data "aws_ami" "foo" {
	most_recent = true
	owners = ["099720109477"]

	filter {
		name = "name"
		values = ["*ubuntu-trusty-14.04-amd64-server-*"]
	}

	filter {
		name = "state"
		values = ["available"]
	}

	filter {
		name = "virtualization-type"
		values = ["hvm"]
	}

	filter {
		name = "is-public"
		values = ["true"]
	}
}
`

func testAccLaunchConfigurationAWSweeperIdsConfig(lc *autoscaling.LaunchConfiguration) string {
	name := lc.LaunchConfigurationName

	return fmt.Sprintf(`
aws_launch_configuration:
  ids:
    - %s
`, *name)
}
