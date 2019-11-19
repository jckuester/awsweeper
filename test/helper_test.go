package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"

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

// sharedAwsClient is an AWS client instance that is shared amongst all acceptance tests
var sharedAwsClient AWS

// sharedTfAwsProvider is a Terraform AWS provider that is shared amongst all acceptance tests
var sharedTfAwsProvider map[string]terraform.ResourceProvider

var argsDryRun = []string{"cmd", "--dry-run", "config.yml"}
var argsForceDelete = []string{"cmd", "--force", "config.yml"}

type AWS struct {
	ec2iface.EC2API
	autoscalingiface.AutoScalingAPI
	elbiface.ELBAPI
	route53iface.Route53API
	cloudformationiface.CloudFormationAPI
	efsiface.EFSAPI
	iamiface.IAMAPI
	kmsiface.KMSAPI
	s3iface.S3API
	stsiface.STSAPI
}

func NewAWS(s *session.Session) AWS {
	return AWS{
		AutoScalingAPI:    autoscaling.New(s),
		CloudFormationAPI: cloudformation.New(s),
		EC2API:            ec2.New(s),
		EFSAPI:            efs.New(s),
		ELBAPI:            elb.New(s),
		IAMAPI:            iam.New(s),
		KMSAPI:            kms.New(s),
		Route53API:        route53.New(s),
		S3API:             s3.New(s),
		STSAPI:            sts.New(s),
	}
}

func init() {
	awsClient, tfAwsProvider := initTests(nil)

	sharedAwsClient = awsClient
	sharedTfAwsProvider = tfAwsProvider
}

func initTests(region *string) (AWS, map[string]terraform.ResourceProvider) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: region},
		SharedConfigState: session.SharedConfigEnable,
	}))

	awsClient := NewAWS(sess)

	tfProvider := map[string]terraform.ResourceProvider{
		"aws": terraformProviderAWS.Provider(),
	}

	err := os.Setenv("AWS_DEFAULT_REGION", *sess.Config.Region)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Setenv("AWS_REGION", *sess.Config.Region)
	if err != nil {
		log.Fatal(err)
	}

	return awsClient, tfProvider
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
