package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_AutoscalingGroup_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/autoscaling-group"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertAutoscalingGroupExists(t, env, id)

	writeConfigID(t, terraformDir, "aws_autoscaling_group", id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n", "--timeout", "5m")
	require.NoError(t, err)

	assertAutoscalingGroupDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_AutoscalingGroup_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/autoscaling-group"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertAutoscalingGroupExists(t, env, id)

	idTag := terraform.Output(t, terraformOptions, "id_tag")
	assertAutoscalingGroupExists(t, env, idTag)

	idTags := terraform.Output(t, terraformOptions, "id_tags")
	assertAutoscalingGroupExists(t, env, idTags)

	writeConfigTag(t, terraformDir, "aws_autoscaling_group")
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n", "--timeout", "5m")
	require.NoError(t, err)

	assertAutoscalingGroupDeleted(t, env, idTag)
	assertAutoscalingGroupDeleted(t, env, idTags)
	assertAutoscalingGroupExists(t, env, id)

	fmt.Println(logBuffer)
}

func assertAutoscalingGroupExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, autoscalingGroupExists(t, env, id))
}

func assertAutoscalingGroupDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, autoscalingGroupExists(t, env, id))
}

func autoscalingGroupExists(t *testing.T, env EnvVars, id string) bool {
	req := env.AWSClient.Autoscalingconn.DescribeAutoScalingGroupsRequest(
		&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []string{id},
		})

	resp, err := req.Send(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.AutoScalingGroups) == 0 {
		return false
	}

	return true
}
