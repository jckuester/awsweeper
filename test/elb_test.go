package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_Elb_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/elb"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertElbExists(t, env, id)

	writeConfigID(t, terraformDir, "aws_elb", id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertElbDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_Elb_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/elb"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertElbExists(t, env, id)

	writeConfigTag(t, terraformDir, "aws_elb")
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertElbDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertElbExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, elbExists(t, env, id))
}

func assertElbDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, elbExists(t, env, id))
}

func elbExists(t *testing.T, env EnvVars, id string) bool {
	req := env.AWSClient.Elasticloadbalancingconn.DescribeLoadBalancersRequest(
		&elasticloadbalancing.DescribeLoadBalancersInput{
			LoadBalancerNames: []string{id},
		})

	resp, err := req.Send(context.Background())

	if err != nil {
		elbErr, ok := err.(awserr.Error)
		if !ok {
			t.Fatal(err)
		}
		if elbErr.Code() == "LoadBalancerNotFound" {
			return false
		}
		t.Fatal(err)
	}

	if len(resp.LoadBalancerDescriptions) == 0 {
		return false
	}

	return true
}
