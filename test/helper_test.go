package test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/jckuester/awstools-lib/aws"
	"github.com/onsi/gomega/gexec"
	"github.com/stretchr/testify/require"
)

const (
	packagePath = "github.com/jckuester/awsweeper"

	// tfstateBucket is the S3 bucket that stores all Terraform states of acceptance tests.
	// Note: the bucket must be located in "profile1" and "region1".
	testTfStateBucket = "awsweeper-testacc-tfstate-492043"
	// profile1 is used as profile for the 1st test account if not overwritten by TEST_AWS_PROFILE1.
	profile1 = "myaccount1"
	// region1 is used as the 1st test region if not overwritten by TEST_AWS_REGION1.
	region1 = "us-west-2"
	// profile2 is used as profile for the 2nd test account if not overwritten by TEST_AWS_PROFILE2.
	profile2 = "myaccount2"
	// region1 is used as the 2nd test region if not overwritten by TEST_AWS_REGION2.
	region2 = "us-east-1"
)

// EnvVars contains environment variables for tests.
type EnvVars struct {
	AWSProfile1 string
	AWSProfile2 string
	AWSRegion1  string
	AWSRegion2  string
	AWSClient   *aws.Client
}

// InitEnv sets environment variables for acceptance tests.
func InitEnv(t *testing.T) EnvVars {
	t.Helper()

	profile1 := getEnvOrDefault(t, "TEST_AWS_PROFILE1", profile1)
	profile2 := getEnvOrDefault(t, "TEST_AWS_PROFILE2", profile2)
	region1 := getEnvOrDefault(t, "TEST_AWS_REGION1", region1)
	region2 := getEnvOrDefault(t, "TEST_AWS_REGION2", region2)

	client, err := aws.NewClient(
		context.Background(),
		config.WithSharedConfigProfile(profile1),
		config.WithRegion(region1))
	require.NoError(t, err)

	return EnvVars{
		AWSProfile1: profile1,
		AWSRegion1:  region1,
		AWSProfile2: profile2,
		AWSRegion2:  region2,
		AWSClient:   client,
	}
}

func getEnvOrDefault(t *testing.T, envName, defaultValue string) string {
	varValue := os.Getenv(envName)
	if varValue == "" {
		varValue = defaultValue

		t.Logf("env %s not set, therefore using the following default value: %s",
			envName, defaultValue)
	}
	return varValue
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

func getTerraformOptions(terraformDir string, env EnvVars, overrideVars ...map[string]interface{}) *terraform.Options {
	name := "awsweeper-testacc-" + strings.ToLower(random.UniqueId())

	vars := map[string]interface{}{
		"profile": env.AWSProfile1,
		"region":  env.AWSRegion1,
		"name":    name,
	}

	if len(overrideVars) > 0 {
		vars = overrideVars[0]
	}

	return &terraform.Options{
		TerraformDir: terraformDir,
		NoColor:      true,
		Vars:         vars,
		// BackendConfig defines where to store the Terraform state files of tests
		BackendConfig: map[string]interface{}{
			"bucket":  testTfStateBucket,
			"key":     fmt.Sprintf("%s.tfstate", name),
			"profile": env.AWSProfile1,
			"region":  env.AWSRegion1,
			"encrypt": true,
		},
	}
}
