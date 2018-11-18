package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/cloudetc/awsweeper/command"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

func TestAccIamInstanceProfile_deleteByIds(t *testing.T) {
	var r1, r2 iam.InstanceProfile

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccIamInstanceProfileConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamInstanceProfileExists("aws_iam_instance_profile.foo", &r1),
					testAccCheckIamInstanceProfileExists("aws_iam_instance_profile.bar", &r2),
					testMainIamInstanceProfileIds(argsDryRun, &r1),
					testIamInstanceProfileExists(&r1),
					testIamInstanceProfileExists(&r2),
					testMainIamInstanceProfileIds(argsForceDelete, &r1),
					testIamInstanceProfileDeleted(&r1),
					testIamInstanceProfileExists(&r2),
				),
			},
		},
	})
}

func testMainIamInstanceProfileIds(args []string, r *iam.InstanceProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml",
			[]byte(testAWSweeperIdsConfig(res.IamInstanceProfile, r.InstanceProfileName)), 0644)
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckIamInstanceProfileExists(name string, r *iam.InstanceProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := client.IAMAPI
		desc := &iam.GetInstanceProfileInput{
			InstanceProfileName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetInstanceProfile(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return fmt.Errorf("IAM instance profile has been deleted")
			}
			return err
		}

		*r = *resp.InstanceProfile

		return nil
	}
}

func testIamInstanceProfileExists(r *iam.InstanceProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.IAMAPI
		desc := &iam.GetInstanceProfileInput{
			InstanceProfileName: r.InstanceProfileName,
		}
		_, err := conn.GetInstanceProfile(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return fmt.Errorf("IAM instance profile has been deleted")
			}
			return err
		}

		return nil
	}
}

func testIamInstanceProfileDeleted(r *iam.InstanceProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.IAMAPI

		desc := &iam.GetInstanceProfileInput{
			InstanceProfileName: r.InstanceProfileName,
		}
		_, err := conn.GetInstanceProfile(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return nil
			}
			return err
		}
		return fmt.Errorf("IAM instance profile hasn't been deleted")
	}
}

const testAccIamInstanceProfileConfig = `
resource "aws_iam_instance_profile" "foo" {
  name  = "awsweeper-testacc-foo"
  role = "${aws_iam_role.test_role.name}"
}

resource "aws_iam_instance_profile" "bar" {
  name  = "awsweeper-testacc-bar"
  role = "${aws_iam_role.test_role.name}"
}

resource "aws_iam_role" "test_role" {
  name = "test_role"
  path = "/awsweeper-testacc/"

  assume_role_policy = "${data.aws_iam_policy_document.test-assume-role-policy.json}"
}

data "aws_iam_policy_document" "test-assume-role-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}
`
