package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/smithy-go"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_DBInstance_DeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}
	t.Skip("Only running manually from time to time, as this test costs money.")

	env := InitEnv(t)

	terraformDir := "./test-fixtures/db-instance"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertDBInstanceExists(t, env, id)

	writeConfigID(t, terraformDir, "aws_db_instance", id)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n", "--timeout", "10m")
	require.NoError(t, err)

	assertDBInstanceDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func TestAcc_DBInstance_DeleteByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}
	t.Skip("Only running manually from time to time, as this test costs money.")

	env := InitEnv(t)

	terraformDir := "./test-fixtures/db-instance"

	terraformOptions := getTerraformOptions(terraformDir, env)

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	id := terraform.Output(t, terraformOptions, "id")
	assertDBInstanceExists(t, env, id)

	writeConfigTag(t, terraformDir, "aws_db_instance")
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n", "--timeout", "5m")
	require.NoError(t, err)

	assertDBInstanceDeleted(t, env, id)

	fmt.Println(logBuffer)
}

func assertDBInstanceExists(t *testing.T, env EnvVars, id string) {
	assert.True(t, dbInstanceExists(t, env, id))
}

func assertDBInstanceDeleted(t *testing.T, env EnvVars, id string) {
	assert.False(t, dbInstanceExists(t, env, id))
}

func dbInstanceExists(t *testing.T, env EnvVars, id string) bool {
	req, err := env.AWSClient.Rdsconn.DescribeDBInstances(context.Background(),
		&rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: &id,
		})

	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() == "DBInstanceNotFound" {
				return false
			}
			t.Fatal(err)
		}
		t.Fatal(err)
	}

	if len(req.DBInstances) == 0 {
		return false
	}

	return true
}
