package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	res "github.com/cloudetc/awsweeper/resource"
)

func TestAcc_Vpc(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	tests := []struct {
		name           string
		configTemplate string
	}{
		{
			name:           "delete by ID",
			configTemplate: configTemplateID,
		},
		{
			name:           "delete by tag",
			configTemplate: configTemplateTag,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := InitEnv(t)

			terraformDir := "./test-fixtures/vpc"

			terraformOptions := getTerraformOptions(terraformDir, env)

			defer terraform.Destroy(t, terraformOptions)

			terraform.InitAndApply(t, terraformOptions)

			vpcID := terraform.Output(t, terraformOptions, "id")
			assertVpcExists(t, vpcID)

			writeConfig(t, tc.configTemplate, terraformDir, res.Vpc, vpcID)
			defer os.Remove(terraformDir + "/config.yml")

			logBuffer, err := runBinary(t, terraformDir, "yes\n")
			require.NoError(t, err)

			assertVpcDeleted(t, vpcID)

			fmt.Println(logBuffer)
		})
	}
}

func assertVpcExists(t *testing.T, id string) {
	assert.True(t, vpcExists(t, id))
}

func assertVpcDeleted(t *testing.T, id string) {
	assert.False(t, vpcExists(t, id))
}

func vpcExists(t *testing.T, id string) bool {
	conn := sharedAwsClient.EC2API

	desc := &ec2.DescribeVpcsInput{
		VpcIds: []*string{&id},
	}
	resp, err := conn.DescribeVpcs(desc)
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
