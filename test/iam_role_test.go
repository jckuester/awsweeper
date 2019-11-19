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

func TestAccIamRole_deleteByIds(t *testing.T) {
	var r1, r2 iam.Role

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: sharedTfAwsProvider,
		Steps: []resource.TestStep{
			{
				Config:             testAccIamRoleConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamRoleExists("aws_iam_role.foo", &r1),
					testAccCheckIamRoleExists("aws_iam_role.bar", &r2),
					testMainIamRoleIds(argsDryRun, &r1),
					testIamRoleExists(&r1),
					testIamRoleExists(&r2),
					testMainIamRoleIds(argsForceDelete, &r1),
					testIamRoleDeleted(&r1),
					testIamRoleExists(&r2),
				),
			},
		},
	})
}

func testMainIamRoleIds(args []string, r *iam.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(testAWSweeperIdsConfig(res.IamRole, r.RoleName)), 0644)
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckIamRoleExists(name string, r *iam.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := sharedAwsClient.IAMAPI
		desc := &iam.GetRoleInput{
			RoleName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetRole(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return fmt.Errorf("IAM role has been deleted")
			}
			return err
		}

		*r = *resp.Role

		return nil
	}
}

func testIamRoleExists(r *iam.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := sharedAwsClient.IAMAPI
		desc := &iam.GetRoleInput{
			RoleName: r.RoleName,
		}
		_, err := conn.GetRole(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return fmt.Errorf("IAM role has been deleted")
			}
			return err
		}

		return nil
	}
}

func testIamRoleDeleted(r *iam.Role) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := sharedAwsClient.IAMAPI

		desc := &iam.GetRoleInput{
			RoleName: r.RoleName,
		}
		_, err := conn.GetRole(desc)
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
		return fmt.Errorf("IAM role hasn't been deleted")
	}
}

const testAccIamRoleConfig = `
resource "aws_iam_role" "foo" {
  name = "foo"
  path = "/awsweeper-testacc/"

  assume_role_policy = "${data.aws_iam_policy_document.test-assume-role-policy.json}"
}

resource "aws_iam_role" "bar" {
  name = "bar"
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

resource "aws_iam_role_policy" "test_role_policy" {
  name = "test_role_policy"
  role = "${aws_iam_role.foo.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test_policy" {
    name        = "test_policy"
    description = "A test policy"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test_attach" {
    role       = "${aws_iam_role.foo.name}"
    policy_arn = "${aws_iam_policy.test_policy.arn}"
}
`
