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
	packagePath       = "github.com/jckuester/awsweeper"
	testTfStateBucket = "awsweeper-testacc-tfstate-492043"
)

// EnvVars contains environment variables for tests.
type EnvVars struct {
	AWSRegion  string
	AWSProfile string
	AWSClient  *aws.Client
}

// InitEnv sets environment variables for acceptance tests.
func InitEnv(t *testing.T) EnvVars {
	t.Helper()

	profile := getEnvOrDefault(t, "AWS_PROFILE", "myaccount1")
	region := getEnvOrDefault(t, "AWS_DEFAULT_REGION", "us-west-2")

	client, err := aws.NewClient(
		context.Background(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region))
	require.NoError(t, err)

	return EnvVars{
		AWSProfile: profile,
		AWSRegion:  region,
		AWSClient:  client,
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
