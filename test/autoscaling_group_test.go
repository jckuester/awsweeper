package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudetc/awsweeper/command"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

func TestAccAutoscalingGroup_deleteByTags(t *testing.T) {
	var asg1, asg2 autoscaling.Group

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccAutoscalingGroupConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.foo", &asg1),
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &asg2),
					testMainTags(argsDryRun, testAWSweeperTagsConfig(res.AutoscalingGroup)),
					testAutoscalingGroupExists(&asg1),
					testAutoscalingGroupExists(&asg2),
					testMainTags(argsForceDelete, testAWSweeperTagsConfig(res.AutoscalingGroup)),
					testAutoscalingGroupDeleted(&asg1),
				),
			},
		},
	})
}

func TestAccAutoscalingGroup_deleteByIds(t *testing.T) {
	var asg1, asg2 autoscaling.Group

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccAutoscalingGroupConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.foo", &asg1),
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &asg2),
					testMainAutoscalingGroupIds(argsDryRun, &asg1),
					testAutoscalingGroupExists(&asg1),
					testAutoscalingGroupExists(&asg2),
					testMainAutoscalingGroupIds(argsForceDelete, &asg1),
					testAutoscalingGroupDeleted(&asg1),
					testAutoscalingGroupExists(&asg2),
				),
			},
		},
	})
}

func testMainAutoscalingGroupIds(args []string, group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		err := afero.WriteFile(res.AppFs, "config.yml",
			[]byte(testAWSweeperIdsConfig(res.AutoscalingGroup, group.AutoScalingGroupName)), 0644)
		if err != nil {
			return err
		}
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckAWSAutoScalingGroupExists(n string, group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no AutoScaling Group ID is set")
		}

		conn := client.AutoScalingAPI

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
		conn := client.AutoScalingAPI
		DescribeAutoscalingGroupOpts := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{asg.AutoScalingGroupName},
		}
		resp, err := conn.DescribeAutoScalingGroups(DescribeAutoscalingGroupOpts)
		if err != nil {
			return err
		}

		if len(resp.AutoScalingGroups) == 0 {
			return fmt.Errorf("autoscaling Group has been deleted")
		}

		return nil
	}
}

func testAutoscalingGroupDeleted(asg *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.AutoScalingAPI
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
			return fmt.Errorf("autoscaling Group hasn't been deleted")
		}

		return nil
	}
}

const testAccAutoscalingGroupConfig = `
resource "aws_autoscaling_group" "foo" {
	name_prefix = "awsweeper-testacc-foo-"
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

resource "aws_autoscaling_group" "bar" {
	name_prefix = "awsweeper-testacc-bar-"
	max_size = "1"
	min_size = "0"

	launch_configuration = "${aws_launch_configuration.foo.id}"
	availability_zones = [ "us-west-2a" ]

	tag {
		key = "foo"
		value = "baz"
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
