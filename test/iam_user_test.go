package test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccIamUser_deleteByIds(t *testing.T) {
	var u1, u2 iam.User

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccIamUserConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIamUserExists("aws_iam_user.foo", &u1),
					testAccCheckIamUserExists("aws_iam_user.bar", &u2),
					testMainIds(argsDryRun, u1.UserName),
					testIamUserExists(&u1),
					testIamUserExists(&u2),
					testMainIds(argsForceDelete, u1.UserName),
					testIamUserDeleted(&u1),
					testIamUserExists(&u2),
				),
			},
		},
	})
}

func testAccCheckIamUserExists(name string, u *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := client.IAMAPI
		desc := &iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetUser(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return fmt.Errorf("IAM user has been deleted")
			}
			return err
		}

		*u = *resp.User

		return nil
	}
}

func testIamUserExists(u *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.IAMAPI
		desc := &iam.GetUserInput{
			UserName: u.UserName,
		}
		_, err := conn.GetUser(desc)
		if err != nil {
			iamErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if iamErr.Code() == NoSuchEntity {
				return fmt.Errorf("IAM user has been deleted")
			}
			return err
		}

		return nil
	}
}

func testIamUserDeleted(u *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.IAMAPI

		desc := &iam.GetUserInput{
			UserName: u.UserName,
		}
		_, err := conn.GetUser(desc)
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
		return fmt.Errorf("IAM user hasn't been deleted")
	}
}

const testAccIamUserConfig = `
resource "aws_iam_user" "foo" {
  name = "foo"
  path = "/awsweeper-testacc/"
}

resource "aws_iam_access_key" "foo" {
  user = "${aws_iam_user.foo.name}"
}

resource "aws_iam_user" "bar" {
  name = "bar"
  path = "/awsweeper-testacc/"
}

resource "aws_iam_user_policy" "test_user_policy" {
  name = "test_user_policy"
  user = "${aws_iam_user.foo.id}"
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

resource "aws_iam_user_policy_attachment" "test_attach" {
    user       = "${aws_iam_user.foo.name}"
    policy_arn = "${aws_iam_policy.test_policy.arn}"
}
`
