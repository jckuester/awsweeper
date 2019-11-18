package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cloudetc/awsweeper/command"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	terraformProviderAWS "github.com/terraform-providers/terraform-provider-aws/aws"
)

const (
	NoSuchEntity     = "NoSuchEntity"
	NoSuchHostedZone = "NoSuchHostedZone"
)

var client *res.AWS
var testAccProviders map[string]terraform.ResourceProvider

var argsDryRun = []string{"cmd", "--dry-run", "config.yml"}
var argsForceDelete = []string{"cmd", "--force", "config.yml"}

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client = res.NewAWS(sess)

	testAccProviders = map[string]terraform.ResourceProvider{
		"aws": terraformProviderAWS.Provider(),
	}
	err := os.Setenv("AWS_DEFAULT_REGION", *sess.Config.Region)
	if err != nil {
		log.Fatal(err)
	}
}

func initWithRegion(region string) map[string]terraform.ResourceProvider {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	}))

	client = res.NewAWS(sess)

	err := os.Setenv("AWS_DEFAULT_REGION", *sess.Config.Region)
	if err != nil {
		log.Fatal(err)
	}

	return map[string]terraform.ResourceProvider{
		"aws": terraformProviderAWS.Provider(),
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("AWS_PROFILE"); v == "" {
		if v := os.Getenv("AWS_ACCESS_KEY_ID"); v == "" {
			t.Fatal("AWS_ACCESS_KEY_ID must be set for acceptance tests")
		}
		if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v == "" {
			t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
		}
	}
	if v := os.Getenv("AWS_DEFAULT_REGION"); v == "" {
		log.Println("[INFO] Test: Using us-west-2 as test region")
		os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	}
}

func testMainTags(args []string, config string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		err := afero.WriteFile(res.AppFs, "config.yml", []byte(config), 0644)
		if err != nil {
			return err
		}
		os.Args = args

		command.WrappedMain()
		return nil
	}
}

func testAWSweeperIdsConfig(resType res.TerraformResourceType, id *string) string {
	return fmt.Sprintf(`
%s:
  - id: %s
`, resType, *id)
}

func testAWSweeperTagsConfig(resType res.TerraformResourceType) string {
	return fmt.Sprintf(`
%s:
  - tags:
      foo: bar
`, resType)
}

func retryOnAwsCode(code string, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == code {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	return resp, err
}
