package test

import (
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/cloudetc/awsweeper/command"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

var client = initClient()

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

var argsDryRun = []string{"cmd", "--dry-run", "config.yml"}
var argsForceDelete = []string{"cmd", "--force", "config.yml"}

func initClient() *res.AWSClient {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &res.AWSClient{
		ASconn:  autoscaling.New(sess),
		CFconn:  cloudformation.New(sess),
		EC2conn: ec2.New(sess),
		EFSconn: efs.New(sess),
		ELBconn: elb.New(sess),
		IAMconn: iam.New(sess),
		KMSconn: kms.New(sess),
		R53conn: route53.New(sess),
		S3conn:  s3.New(sess),
		STSconn: sts.New(sess),
	}
}

func init() {
	testAccProvider = aws.Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"aws": testAccProvider,
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
	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func testMainTags(args []string, config string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res.AppFs = afero.NewMemMapFs()
		afero.WriteFile(res.AppFs, "config.yml", []byte(config), 0644)
		os.Args = args

		command.WrappedMain()
		return nil
	}
}