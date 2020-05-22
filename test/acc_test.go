package test

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcc_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	tests := []struct {
		name                    string
		flags                   []string
		expectedLogs            []string
		unexpectedLogs          []string
		expectResourceIsDeleted bool
	}{
		{
			name:  "with dry-run flag",
			flags: []string{"--dry-run"},
			expectedLogs: []string{
				"SHOWING RESOURCES THAT WOULD BE DELETED (DRY RUN)",
				"TOTAL NUMBER OF RESOURCES THAT WOULD BE DELETED: 1",
			},
			unexpectedLogs: []string{
				"STARTING TO DELETE RESOURCES",
				"TOTAL NUMBER OF DELETED RESOURCES:",
			},
		},
		{
			name: "without dry-run flag",
			expectedLogs: []string{
				"SHOWING RESOURCES THAT WOULD BE DELETED (DRY RUN)",
				"TOTAL NUMBER OF RESOURCES THAT WOULD BE DELETED: 1",
				"STARTING TO DELETE RESOURCES",
				"TOTAL NUMBER OF DELETED RESOURCES: 1",
			},
			expectResourceIsDeleted: true,
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
			assertVpcExists(t, env, vpcID)

			writeConfigID(t, terraformDir, "aws_vpc", vpcID)
			defer os.Remove(terraformDir + "/config.yml")

			logBuffer, err := runBinary(t, terraformDir, "YES\n", tc.flags...)
			require.NoError(t, err)

			if tc.expectResourceIsDeleted {
				assertVpcDeleted(t, env, vpcID)
			} else {
				assertVpcExists(t, env, vpcID)
			}

			actualLogs := logBuffer.String()

			for _, expectedLogEntry := range tc.expectedLogs {
				assert.Contains(t, actualLogs, expectedLogEntry)
			}

			for _, unexpectedLogEntry := range tc.unexpectedLogs {
				assert.NotContains(t, actualLogs, unexpectedLogEntry)
			}

			fmt.Println(actualLogs)
		})
	}
}

func TestAcc_Version(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	logBuffer, err := runBinary(t, "", "", "-version")
	require.NoError(t, err)

	actualLogs := logBuffer.String()

	assert.Contains(t, actualLogs, fmt.Sprintf(`
version: dev
commit: ?
built at: ?
using: %s`, runtime.Version()))

	fmt.Println(actualLogs)
}

func TestAcc_WrongPathToFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	logBuffer, err := runBinary(t, "/does/not/exist", "")
	require.Error(t, err)

	actualLogs := logBuffer.String()

	assert.Contains(t, actualLogs, "Error: failed to create resource filter: open /does/not/exist/config.yml: no such file or directory")

	fmt.Println(actualLogs)
}
