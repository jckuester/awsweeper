package test_acceptance

import (
	"os"

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
	"github.com/cloudetc/awsweeper/command_wipe"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

var client = initClient()
var argsDryRun = []string{"cmd", "--dry-run", "config.yml"}
var argsForceDelete = []string{"cmd", "--force", "config.yml"}

type AWSClient struct {
	ec2conn         *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn         *elb.ELB
	r53conn         *route53.Route53
	cfconn          *cloudformation.CloudFormation
	efsconn         *efs.EFS
	iamconn         *iam.IAM
	kmsconn         *kms.KMS
	s3conn          *s3.S3
	stsconn         *sts.STS
}

func initClient() *AWSClient {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &AWSClient{
		autoscalingconn: autoscaling.New(sess),
		ec2conn:         ec2.New(sess),
		elbconn:         elb.New(sess),
		r53conn:         route53.New(sess),
		cfconn:          cloudformation.New(sess),
		efsconn:         efs.New(sess),
		iamconn:         iam.New(sess),
		kmsconn:         kms.New(sess),
		s3conn:          s3.New(sess),
		stsconn:         sts.New(sess),
	}
}

func testMainTags(args []string, config string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		command_wipe.OsFs = afero.NewMemMapFs()
		afero.WriteFile(command_wipe.OsFs, "config.yml", []byte(config), 0644)
		os.Args = args

		command_wipe.WrappedMain()
		return nil
	}
}
