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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	res "github.com/cloudetc/awsweeper/resource"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/onsi/gomega/gexec"
	"github.com/stretchr/testify/require"
)

const (
	packagePath       = "github.com/cloudetc/awsweeper"
	testTfStateBucket = "awsweeper-testacc-tfstate-492043"

	NoSuchEntity     = "NoSuchEntity"
	NoSuchHostedZone = "NoSuchHostedZone"
)

// EnvVars contains environment variables for tests.
type EnvVars struct {
	AWSRegion  string
	AWSProfile string
	AWSClient  *res.AWS
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

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	}))

	return EnvVars{
		AWSProfile: profile,
		AWSRegion:  region,
		AWSClient:  res.NewAWS(sess),
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

func writeConfigID(t *testing.T, terraformDir string, resType string, id string) {
	config := fmt.Sprintf(`%s:
  - id: %s
`, resType, id)

	err := ioutil.WriteFile(terraformDir+"/config.yml",
		[]byte(config), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func writeConfigTag(t *testing.T, terraformDir string, resType string) {
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
		// BackendConfig defines where to store the Terraform state files of tests
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
