package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/cloudetc/awsweeper/command"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

func TestAccKmsKey_deleteByTags(t *testing.T) {
	// TODO implement tag support
	t.Skip("Costs money even in free tier")
	var k1, k2 kms.KeyMetadata

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccKmsKeyConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsKeyExists("aws_kms_key.foo", &k1),
					testAccCheckKmsKeyExists("aws_kms_key.bar", &k2),
					testMainTags(argsDryRun, testAWSweeperTagsConfig(res.KmsKey)),
					testKmsKeyExists(&k1),
					testKmsKeyExists(&k2),
					testMainTags(argsForceDelete, testAWSweeperTagsConfig(res.KmsKey)),
					testKmsKeyDeleted(&k1),
					testKmsKeyExists(&k2),
				),
			},
		},
	})
}

func TestAccKmsKey_deleteByIds(t *testing.T) {
	t.Skip("Costs money even in free tier")
	var k1, k2 kms.KeyMetadata

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:             testAccKmsKeyConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKmsKeyExists("aws_kms_key.foo", &k1),
					testAccCheckKmsKeyExists("aws_kms_key.bar", &k2),
					testMainKmsKeyIds(argsDryRun, &k1),
					testKmsKeyExists(&k1),
					testKmsKeyExists(&k2),
					testMainKmsKeyIds(argsForceDelete, &k1),
					testKmsKeyDeleted(&k1),
					testKmsKeyExists(&k2),
				),
			},
		},
	})
}

func testAccCheckKmsKeyExists(name string, k *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := client.KMSAPI

		o, err := retryOnAwsCode("NotFoundException", func() (interface{}, error) {
			return conn.DescribeKey(&kms.DescribeKeyInput{
				KeyId: aws.String(rs.Primary.ID),
			})
		})
		if err != nil {
			return err
		}
		out := o.(*kms.DescribeKeyOutput)

		*k = *out.KeyMetadata

		return nil
	}
}

func testMainKmsKeyIds(args []string, k *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(testAWSweeperIdsConfig(res.KmsKey, k.KeyId)), 0644)
		os.Args = args

		command.WrappedMain()
		return nil
	}
}

func testKmsKeyExists(k *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.KMSAPI
		_, err := retryOnAwsCode("NotFoundException", func() (interface{}, error) {
			return conn.DescribeKey(&kms.DescribeKeyInput{
				KeyId: k.KeyId,
			})
		})
		if err != nil {
			kmsErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if kmsErr.Code() == "NotFoundException" {
				return fmt.Errorf("KMS key not found")
			}
			return err
		}

		return nil
	}
}

func testKmsKeyDeleted(k *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := client.KMSAPI

		resp, err := conn.DescribeKey(&kms.DescribeKeyInput{
			KeyId: k.KeyId,
		})
		if err != nil {
			kmsErr, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if kmsErr.Code() == "NotFoundException" {
				return nil
			}
			return err
		}
		if *resp.KeyMetadata.KeyState == "PendingDeletion" {
			return nil
		}
		return fmt.Errorf("KMS key hasn't been deleted")
	}
}

const testAccKmsKeyConfig = `
resource "aws_kms_key" "foo" {
    description = "AWSweeper acc test"
    deletion_window_in_days = 7

    tags {
		foo = "bar"
		Name = "awsweeper-testacc"
	}
}

resource "aws_kms_key" "bar" {
    description = "AWSweeper acc test"
    deletion_window_in_days = 7

    tags {
		bar = "baz"
		Name = "awsweeper-testacc"
	}
}
`
