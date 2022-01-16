package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_IamPolicy_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/iam-policy"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	arn := terraform.Output(t, terraformOptions, "arn")
	assertIamPolicyExists(t, env, arn)

	id := terraform.Output(t, terraformOptions, "id")
	writeConfigID(t, terraformDir, "aws_iam_policy", id)

	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertIamPolicyDeleted(t, env, arn)

	fmt.Println(logBuffer)
}

func assertIamPolicyExists(t *testing.T, env EnvVars, arn string) {
	assert.True(t, iamPolicyExists(t, env, arn))
}

func assertIamPolicyDeleted(t *testing.T, env EnvVars, arn string) {
	assert.False(t, iamPolicyExists(t, env, arn))
}

func iamPolicyExists(t *testing.T, env EnvVars, arn string) bool {
	_, err := env.AWSClient.Iamconn.GetPolicy(
		context.Background(),
		&iam.GetPolicyInput{
			PolicyArn: &arn,
		})

	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() == "NoSuchEntity" {
				return false
			}
			t.Fatal(err)
		}
		t.Fatal(err)
	}

	return true
}
