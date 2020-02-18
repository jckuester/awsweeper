package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cloudetc/awsweeper/command"
	"github.com/spf13/afero"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAmi_deleteByIds(t *testing.T) {
	var image1, image2 ec2.Image
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: sharedTfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccAmiConfig(rName),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists("aws_ami.foo", &image1),
					testAccCheckAmiExists("aws_ami.bar", &image2),
					testMainAmiIds(argsDryRun, &image1),
					testAmiExists(&image1),
					testAmiExists(&image2),
					testMainAmiIds(argsForceDelete, &image1),
					testAmiDeleted(&image1),
					testAmiExists(&image2),
				),
			},
		},
	})
}

func testMainAmiIds(args []string, image *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		err := afero.WriteFile(res.AppFs, "config.yml", []byte(testAWSweeperIdsConfig(res.Ami, image.ImageId)), 0644)
		if err != nil {
			return err
		}
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckAmiExists(n string, ami *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("AMI Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AMI ID is set")
		}

		conn := sharedAwsClient.EC2API

		var resp *ec2.DescribeImagesOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			opts := &ec2.DescribeImagesInput{
				ImageIds: []*string{aws.String(rs.Primary.ID)},
			}
			var err error
			resp, err = conn.DescribeImages(opts)
			if err != nil {
				// This can be just eventual consistency
				awsErr, ok := err.(awserr.Error)
				if ok && awsErr.Code() == "InvalidAMIID.NotFound" {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("Unable to find AMI after retries: %s", err)
		}

		if len(resp.Images) == 0 {
			return fmt.Errorf("AMI not found")
		}
		*ami = *resp.Images[0]
		return nil
	}
}

func testAmiExists(image *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := sharedAwsClient.EC2API
		opts := &ec2.DescribeImagesInput{
			ImageIds: []*string{image.ImageId},
		}
		resp, err := conn.DescribeImages(opts)
		if err != nil {
			return err
		}
		if len(resp.Images) == 0 {
			return fmt.Errorf("image has been deleted")
		}

		return nil
	}
}

func testAmiDeleted(image *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := sharedAwsClient.EC2API
		opts := &ec2.DescribeImagesInput{
			ImageIds: []*string{image.ImageId},
		}
		resp, err := conn.DescribeImages(opts)
		if err != nil {
			// This can be just eventual consistency
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == "InvalidAMIID.NotFound" {
				return nil
			}

			return err
		}

		if len(resp.Images) != 0 {
			return fmt.Errorf("AMI still exists")
		}

		return nil
	}
}

func testAccAmiConfig(rName string) string {
	return fmt.Sprintf(`

resource "aws_ami" "foo" {
  name = "awsweeper-testacc-foo-%s"
  root_device_name = "/dev/xvda"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/xvda"
    snapshot_id = "${aws_ebs_snapshot.foo.id}"
    volume_size = 8
  }
}

resource "aws_ami" "bar" {
  name = "awsweeper-testacc-bar-%s"
  root_device_name = "/dev/xvda"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/xvda"
    snapshot_id = "${aws_ebs_snapshot.foo.id}"
    volume_size = 8
  }
}

resource "aws_ebs_volume" "foo" {
  availability_zone = "us-west-2a"
  size              = 5

	tags {
		Name = "awsweeper-testacc"
	}
}

resource "aws_ebs_snapshot" "foo" {
  volume_id = "${aws_ebs_volume.foo.id}"

	tags {
		Name = "awsweeper-testacc"
	}
}
`, rName, rName)
}
