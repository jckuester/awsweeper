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

func TestAutoscalingGroup_tags(t *testing.T) {
	var asg autoscaling.Group

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccAutoscalingGroupConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.foo", &asg),
					testMainTags(argsDryRun, testAccAutoscalingGroupAWSweeperTagsConfig),
					testAutoscalingGroupExists(&asg),
					testMainTags(argsForceDelete, testAccAutoscalingGroupAWSweeperTagsConfig),
					testAutoscalingGroupDeleted(&asg),
				),
			},
		},
	})
}

func TestAutoscalingGroup_ids(t *testing.T) {
	var asg autoscaling.Group

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccAutoscalingGroupConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.foo", &asg),
					testMainAutoscalingGroupIds(argsDryRun, &asg),
					testAutoscalingGroupExists(&asg),
					testMainAutoscalingGroupIds(argsForceDelete, &asg),
					testAutoscalingGroupDeleted(&asg),
				),
			},
		},
	})
}

func testAccCheckAWSAutoScalingGroupExists(n string, group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AutoScaling Group ID is set")
		}

		conn := client.autoscalingconn

		describeGroups, err := conn.DescribeAutoScalingGroups(
			&autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{aws.String(rs.Primary.ID)},
			})

		if err != nil {
			return err
		}

		if len(describeGroups.AutoScalingGroups) != 1 ||
			*describeGroups.AutoScalingGroups[0].AutoScalingGroupName != rs.Primary.ID {
			return fmt.Errorf("AutoScaling Group not found")
		}

		*group = *describeGroups.AutoScalingGroups[0]

		return nil
	}
}

func testAutoscalingGroupExists(asg *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.autoscalingconn
		DescribeAutoscalingGroupOpts := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{asg.AutoScalingGroupName},
		}
		resp, err := conn.DescribeAutoScalingGroups(DescribeAutoscalingGroupOpts)
		if err != nil {
			return err
		}

		if len(resp.AutoScalingGroups) == 0 {
			return fmt.Errorf("Autoscaling Group has been deleted")
		}

		return nil
	}
}

func testAutoscalingGroupDeleted(asg *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.autoscalingconn
		DescribeAutoscalingGroupOpts := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{asg.AutoScalingGroupName},
		}
		resp, err := conn.DescribeAutoScalingGroups(DescribeAutoscalingGroupOpts)
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() == "InvalidGroup.NotFound" {
				return nil
			}
			return err
		}

		if len(resp.AutoScalingGroups) != 0 {
			return fmt.Errorf("Autoscaling Group hasn't been deleted")
		}

		return nil
	}
}

func testMainAutoscalingGroupIds(args []string, group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		command_wipe.OsFs = afero.NewMemMapFs()
		afero.WriteFile(command_wipe.OsFs, "config.yml", []byte(testAccAutoscalingGroupAWSweeperIdsConfig(group)), 0644)
		os.Args = args

		command_wipe.WrappedMain()
		return nil
	}
}

const testAccAutoscalingGroupConfig = `
resource "aws_autoscaling_group" "foo" {
	name_prefix = "awsweeper-testacc-"
	max_size = "1"
	min_size = "0"

	launch_configuration = "${aws_launch_configuration.foo.id}"
	availability_zones = [ "us-west-2a" ]

	tag {
		key = "foo"
		value = "bar"
		propagate_at_launch = false
	}

	tag {
		key = "Name"
		value = "awsweeper-testacc"
		propagate_at_launch = false
	}
}

resource "aws_launch_configuration" "foo" {
	name_prefix = "awsweeper-testacc-tags-"
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

const testAccAutoscalingGroupAWSweeperTagsConfig = `
aws_autoscaling_group:
  tags:
    foo: bar
`

func testAccAutoscalingGroupAWSweeperIdsConfig(group *autoscaling.Group) string {
	name := group.AutoScalingGroupName

	return fmt.Sprintf(`
aws_autoscaling_group:
  ids:
    - %s
`, *name)
}
