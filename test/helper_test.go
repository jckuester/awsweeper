package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"

	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/onsi/gomega/gexec"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
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

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/helper/resource"
)

const (
	packagePath       = "github.com/cloudetc/awsweeper"
	testTfStateBucket = "awsweeper-testacc-tfstate-492043"

	NoSuchEntity     = "NoSuchEntity"
	NoSuchHostedZone = "NoSuchHostedZone"
)

var (
	sharedAwsClient AWS
)

type AWS struct {
	autoscalingiface.AutoScalingAPI
	cloudformationiface.CloudFormationAPI
	ec2iface.EC2API
	ecsiface.ECSAPI
	efsiface.EFSAPI
	elbiface.ELBAPI
	iamiface.IAMAPI
	kmsiface.KMSAPI
	rdsiface.RDSAPI
	route53iface.Route53API
	s3iface.S3API
	stsiface.STSAPI
}

func NewAWS(s *session.Session) AWS {
	return AWS{
		AutoScalingAPI:    autoscaling.New(s),
		CloudFormationAPI: cloudformation.New(s),
		EC2API:            ec2.New(s),
		ECSAPI:            ecs.New(s),
		EFSAPI:            efs.New(s),
		ELBAPI:            elb.New(s),
		IAMAPI:            iam.New(s),
		KMSAPI:            kms.New(s),
		RDSAPI:            rds.New(s),
		Route53API:        route53.New(s),
		S3API:             s3.New(s),
		STSAPI:            sts.New(s),
	}
}

func init() {
	sharedAwsClient = initTests(nil)
}

func initTests(region *string) AWS {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: region},
		SharedConfigState: session.SharedConfigEnable,
	}))
	return NewAWS(sess)
}

// EnvVars contains environment variables for that must be set for tests.
type EnvVars struct {
	AWSRegion  string
	AWSProfile string
}

// InitEnv sets environment variables for acceptance tests.
func InitEnv(t *testing.T) EnvVars {
	t.Helper()

	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		t.Fatal("env variable AWS_PROFILE needs to be set for tests")
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		t.Fatal("env variable AWS_DEFAULT_REGION needs to be set for tests")
	}

	return EnvVars{
		AWSProfile: profile,
		AWSRegion:  region,
	}
}

func runBinary(t *testing.T, terraformDir, userInput string, flags ...string) (*bytes.Buffer, error) {
	defer gexec.CleanupBuildArtifacts()

	compiledPath, err := gexec.Build(packagePath)
	require.NoError(t, err)

	args := []string{terraformDir + "/config.yml"}
	if flags != nil {
		args = append(flags, args...)
	}

	logBuffer := &bytes.Buffer{}

	p := exec.Command(compiledPath, args...)
	p.Stdin = strings.NewReader(userInput)
	p.Stdout = logBuffer
	p.Stderr = logBuffer

	err = p.Run()

	return logBuffer, err
}

func writeConfigID(t *testing.T, terraformDir string, resType res.TerraformResourceType, id string) {
	config := fmt.Sprintf(`%s:
  - id: %s
`, resType, id)

	err := ioutil.WriteFile(terraformDir+"/config.yml",
		[]byte(config), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func writeConfigTag(t *testing.T, terraformDir string, resType res.TerraformResourceType) {
	config := fmt.Sprintf(`%s:
  - tags:
      awsweeper: test-acc
`, resType)

	err := ioutil.WriteFile(terraformDir+"/config.yml",
		[]byte(config), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func getTerraformOptions(terraformDir string, env EnvVars) *terraform.Options {
	name := "awsweeper-testacc-" + strings.ToLower(random.UniqueId())

	return &terraform.Options{
		TerraformDir: terraformDir,
		NoColor:      true,
		Vars: map[string]interface{}{
			"region":  env.AWSRegion,
			"profile": env.AWSProfile,
			"name":    name,
		},
		// BackendConfig defines where to store the Terraform state file of the test in S3
		BackendConfig: map[string]interface{}{
			"bucket":  testTfStateBucket,
			"key":     fmt.Sprintf("%s/%s.tfstate", terraformDir, name),
			"region":  env.AWSRegion,
			"profile": env.AWSProfile,
			"encrypt": true,
		},
	}
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
