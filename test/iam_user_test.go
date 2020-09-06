package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
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
	req := env.AWSClient.Iamconn.GetUserRequest(&iam.GetUserInput{
		UserName: &id,
	})

	_, err := req.Send(context.Background())

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			t.Fatal()
		}
		if ec2err.Code() == "NoSuchEntity" {
			return false
		}
		t.Fatal()
	}

	return true
}
