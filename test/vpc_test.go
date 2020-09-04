package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_Vpc_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/vpc"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertVpcExists(t, env, id)

	writeConfigID(t, terraformDir, "aws_vpc", id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertVpcDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_Vpc_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/vpc"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertVpcExists(t, env, id)

	writeConfigTag(t, terraformDir, "aws_vpc")
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertVpcDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertVpcExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, vpcExists(t, env, id))
}

func assertVpcDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, vpcExists(t, env, id))
}

func vpcExists(t *testing.T, env EnvVars, id string) bool {
	req := env.AWSClient.Ec2conn.DescribeVpcsRequest(
		&ec2.DescribeVpcsInput{
			VpcIds: []string{id},
		})

	resp, err := req.Send(context.Background())

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			t.Fatal()
		}
		if ec2err.Code() == "InvalidVpcID.NotFound" {
			return false
		}
		t.Fatal()
	}

	if len(resp.Vpcs) == 0 {
		return false
	}

	return true
}
