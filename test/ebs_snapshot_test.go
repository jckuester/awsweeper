package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_EbsSnapshot_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/ebs-snapshot"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertEbsSnapshotExists(t, env, id)

	writeConfigID(t, terraformDir, "aws_ebs_snapshot", id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n", "-debug")
	require.NoError(t, err)

	assertEbsSnapshotDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_EbsSnapshot_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	env := InitEnv(t)

	terraformDir := "./test-fixtures/ebs-snapshot"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertEbsSnapshotExists(t, env, id)

	writeConfigTag(t, terraformDir, "aws_ebs_snapshot")
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertEbsSnapshotDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertEbsSnapshotExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, ebsSnapshotExists(t, env, id))
}

func assertEbsSnapshotDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, ebsSnapshotExists(t, env, id))
}

func ebsSnapshotExists(t *testing.T, env EnvVars, id string) bool {
	opts := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{&id},
	}

	resp, err := env.AWSClient.DescribeSnapshots(opts)
	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			t.Fatal(err)
		}
		if ec2err.Code() == "InvalidSnapshot.NotFound" {
			return false
		}
		t.Fatal(err)
	}

	if len(resp.Snapshots) == 0 {
		return false
	}

	return true
}
