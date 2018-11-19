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

func TestAccKeyPair_deleteByIds(t *testing.T) {
	var kp1, kp2 ec2.KeyPairInfo

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccKeyPairConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyPairExists("aws_key_pair.foo", &kp1),
					testAccCheckKeyPairExists("aws_key_pair.bar", &kp2),
					testMainKeyPairIds(argsDryRun, &kp1),
					testKeyPairExists(&kp1),
					testKeyPairExists(&kp2),
					testMainKeyPairIds(argsForceDelete, &kp1),
					testKeyPairDeleted(&kp1),
					testKeyPairExists(&kp2),
				),
			},
		},
	})
}

func testMainKeyPairIds(args []string, kp *ec2.KeyPairInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(testAWSweeperIdsConfig(res.KeyPair, kp.KeyName)), 0644)
		os.Args = args
		command.WrappedMain()
		return nil
	}
}

func testAccCheckKeyPairExists(n string, kp *ec2.KeyPairInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no key pair ID is set")
		}

		conn := client.EC2API
		opts := &ec2.DescribeKeyPairsInput{
			KeyNames: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeKeyPairs(opts)
		if err != nil {
			return err
		}
		if len(resp.KeyPairs) == 0 {
			return fmt.Errorf("key pair not found")
		}

		*kp = *resp.KeyPairs[0]

		return nil
	}
}

func testKeyPairExists(kp *ec2.KeyPairInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.EC2API
		opts := &ec2.DescribeKeyPairsInput{
			KeyNames: []*string{kp.KeyName},
		}
		resp, err := conn.DescribeKeyPairs(opts)
		if err != nil {
			return err
		}
		if len(resp.KeyPairs) == 0 {
			return fmt.Errorf("key pair has been deleted")
		}

		return nil
	}
}

func testKeyPairDeleted(kp *ec2.KeyPairInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.EC2API
		opts := &ec2.DescribeKeyPairsInput{
			KeyNames: []*string{kp.KeyName},
		}
		resp, err := conn.DescribeKeyPairs(opts)
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() == "InvalidKeyPair.NotFound" {
				return nil
			}
			return err
		}

		if len(resp.KeyPairs) != 0 {
			return fmt.Errorf("key pair hasn't been deleted")

		}

		return nil
	}
}

const testAccKeyPairConfig = `
resource "aws_key_pair" "foo" {
  key_name   = "awsweeper-testacc-foo"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com"
}

resource "aws_key_pair" "bar" {
  key_name   = "awsweeper-testacc-bar"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 email@example.com"
}
`
