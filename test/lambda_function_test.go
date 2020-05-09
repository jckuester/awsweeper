package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/lambda"

	res "github.com/cloudetc/awsweeper/resource"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_LambdaFunction_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/lambda-function"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")

	assertLambdaFunctionExists(t, env, id)

	writeConfigID(t, terraformDir, res.LambdaFunction, id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertLambdaFunctionDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_LambdaFunction_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}
	t.Skip("Tags not supported yet.")

	env := InitEnv(t)

	terraformDir := "./test-fixtures/lambda-function"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")

	assertLambdaFunctionExists(t, env, id)

	writeConfigTag(t, terraformDir, res.LambdaFunction)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertLambdaFunctionDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertLambdaFunctionExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, lambdaFunctionExists(t, env, id))
}

func assertLambdaFunctionDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, lambdaFunctionExists(t, env, id))
}

func lambdaFunctionExists(t *testing.T, env EnvVars, id string) bool {
	opts := &lambda.GetFunctionInput{
		FunctionName: &id,
	}

	_, err := env.AWSClient.GetFunction(opts)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if !ok {
			t.Fatal(err)
		}
		if awsErr.Code() == "ResourceNotFoundException" {
			return false
		}
		t.Fatal(err)
	}

	return true
}
