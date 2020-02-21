package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	res "github.com/cloudetc/awsweeper/resource"
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

	elbID := terraform.Output(t, terraformOptions, "id")
	assertElbExists(t, elbID)

	writeConfigID(t, terraformDir, res.Elb, elbID)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertElbDeleted(t, elbID)

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

	elbID := terraform.Output(t, terraformOptions, "id")
	assertElbExists(t, elbID)

	writeConfigTag(t, terraformDir, res.Elb)
	defer os.Remove(terraformDir + "/config.yml")

	logBuffer, err := runBinary(t, terraformDir, "YES\n")
	require.NoError(t, err)

	assertElbDeleted(t, elbID)

	fmt.Println(logBuffer)
}

func assertElbExists(t *testing.T, id string) {
	assert.True(t, elbExists(t, id))
}

func assertElbDeleted(t *testing.T, id string) {
	assert.False(t, elbExists(t, id))
}

func elbExists(t *testing.T, id string) bool {
	conn := sharedAwsClient.ELBAPI

	DescribeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{&id},
	}
	resp, err := conn.DescribeLoadBalancers(DescribeElbOpts)
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
