package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
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

	writeConfigID(t, terraformDir, "aws_lambda_function", id)
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

	env := InitEnv(t)

	terraformDir := "./test-fixtures/lambda-function"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")

	assertLambdaFunctionExists(t, env, id)

	writeConfigTag(t, terraformDir, "aws_lambda_function")
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
	_, err := env.AWSClient.Lambdaconn.GetFunction(
		context.Background(),
		&lambda.GetFunctionInput{
			FunctionName: &id,
		})

	if err != nil {
		var awsErr *types.ResourceNotFoundException
		if errors.As(err, &awsErr) {
			return false
		}
		t.Fatal(err)
	}

	return true
}
