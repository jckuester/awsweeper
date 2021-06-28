package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_IamUser_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/iam-user"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertIamUserExists(t, env, id)

	writeConfigID(t, terraformDir, "aws_iam_user", id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertIamUserDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_IamUser_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/iam-user"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertIamUserExists(t, env, id)

	writeConfigTag(t, terraformDir, "aws_iam_user")
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertIamUserDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertIamUserExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, iamUserExists(t, env, id))
}

func assertIamUserDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, iamUserExists(t, env, id))
}

func iamUserExists(t *testing.T, env EnvVars, id string) bool {
	_, err := env.AWSClient.Iamconn.GetUser(
		context.Background(),
		&iam.GetUserInput{
			UserName: &id,
		})

	if err != nil {
		var awsErr *types.NoSuchEntityException
		if errors.As(err, &awsErr) {
			return false
		}
		t.Fatal(err)
	}

	return true
}
