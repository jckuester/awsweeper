package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_ECSCluster_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/ecs-cluster"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")

	assertEcsClusterExists(t, env, id)

	writeConfigID(t, terraformDir, "aws_ecs_cluster", id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertEcsClusterDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_ECSCluster_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/ecs-cluster"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")

	assertEcsClusterExists(t, env, id)

	writeConfigTag(t, terraformDir, "aws_ecs_cluster")
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertEcsClusterDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertEcsClusterExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, ecsClusterExists(t, env, id))
}

func assertEcsClusterDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, ecsClusterExists(t, env, id))
}

func ecsClusterExists(t *testing.T, env EnvVars, id string) bool {
	req := env.AWSClient.Ecsconn.DescribeClustersRequest(
		&ecs.DescribeClustersInput{
			Clusters: []string{id},
		})

	resp, err := req.Send(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Clusters) == 0 {
		return false
	}

	if *resp.Clusters[0].Status == "INACTIVE" {
		return false
	}

	return true
}
