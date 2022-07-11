package test

import (
	"fmt"
	"os"
	"regexp"
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

	logBuffer, err := runBinary(t, "", "", "--version")
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

func TestAcc_ProfilesAndRegions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test.")
	}

	testVars := InitEnv(t)

	terraformDir := "./test-fixtures/multiple-profiles-and-regions"

	terraformOptions := getTerraformOptions(terraformDir, testVars, map[string]interface{}{
		"profile1": testVars.AWSProfile1,
		"profile2": testVars.AWSProfile2,
		"region1":  testVars.AWSRegion1,
		"region2":  testVars.AWSRegion2,
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	vpcID1 := terraform.Output(t, terraformOptions, "id1")
	vpcID2 := terraform.Output(t, terraformOptions, "id2")
	vpcID3 := terraform.Output(t, terraformOptions, "id3")
	vpcID4 := terraform.Output(t, terraformOptions, "id4")

	writeConfigID(t, terraformDir, "aws_vpc", fmt.Sprintf("%s|%s|%s|%s", vpcID1, vpcID2, vpcID3, vpcID4))
	defer os.Remove(terraformDir + "/config.yml")

	tests := []struct {
		name            string
		args            []string
		envs            map[string]string
		expectedLogs    []string
		expectedErrCode int
	}{
		{
			name: "multiple profiles and regions via flag",
			args: []string{
				"-p", fmt.Sprintf("%s,%s", testVars.AWSProfile1, testVars.AWSProfile2),
				"-r", fmt.Sprintf("%s,%s", testVars.AWSRegion1, testVars.AWSRegion2),
				"--dry-run",
			},
			expectedLogs: []string{
				fmt.Sprintf("Id:\\s+%[1]s\\s+Profile:\\s+%[2]s\\s+Region:\\s+%[3]s", vpcID1, testVars.AWSProfile1, testVars.AWSRegion1),
				fmt.Sprintf("Id:\\s+%[1]s\\s+Profile:\\s+%[2]s\\s+Region:\\s+%[3]s", vpcID2, testVars.AWSProfile1, testVars.AWSRegion2),
				fmt.Sprintf("Id:\\s+%[1]s\\s+Profile:\\s+%[2]s\\s+Region:\\s+%[3]s", vpcID3, testVars.AWSProfile2, testVars.AWSRegion1),
				fmt.Sprintf("Id:\\s+%[1]s\\s+Profile:\\s+%[2]s\\s+Region:\\s+%[3]s", vpcID4, testVars.AWSProfile2, testVars.AWSRegion2),
				"TOTAL NUMBER OF RESOURCES THAT WOULD BE DELETED: 4",
			},
		},
		{
			name: "profile via env, multiple regions via flag",
			args: []string{
				"-r", fmt.Sprintf("%s,%s", testVars.AWSRegion1, testVars.AWSRegion2),
				"--dry-run",
			},
			envs: map[string]string{
				"AWS_PROFILE": testVars.AWSProfile1,
			},
			expectedLogs: []string{
				fmt.Sprintf("Id:\\s+%[1]s\\s+Profile:\\s+%[2]s\\s+Region:\\s+%[3]s", vpcID1, testVars.AWSProfile1, testVars.AWSRegion1),
				fmt.Sprintf("Id:\\s+%[1]s\\s+Profile:\\s+%[2]s\\s+Region:\\s+%[3]s", vpcID2, testVars.AWSProfile1, testVars.AWSRegion2),
				"TOTAL NUMBER OF RESOURCES THAT WOULD BE DELETED: 2",
			},
		},
		{
			name: "profile and region via env",
			envs: map[string]string{
				"AWS_PROFILE":        testVars.AWSProfile1,
				"AWS_DEFAULT_REGION": testVars.AWSRegion2,
			},
			args: []string{
				"--dry-run",
			},
			expectedLogs: []string{
				fmt.Sprintf("Id:\\s+%[1]s\\s+Profile:\\s+%[2]s\\s+Region:\\s+%[3]s", vpcID2, testVars.AWSProfile1, testVars.AWSRegion2),
				"TOTAL NUMBER OF RESOURCES THAT WOULD BE DELETED: 1",
			},
		},
		{
			name: "profile via env, using default region from AWS config file",
			envs: map[string]string{
				"AWS_PROFILE": testVars.AWSProfile1,
			},
			args: []string{
				"--dry-run",
			},
			expectedLogs: []string{
				fmt.Sprintf("Id:\\s+%[1]s\\s+Profile:\\s+%[2]s\\s+Region:\\s+%[3]s", vpcID1, testVars.AWSProfile1, testVars.AWSRegion1),
				"TOTAL NUMBER OF RESOURCES THAT WOULD BE DELETED: 1",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}
			logBuffer, err := runBinary(t, terraformDir, "", tc.args...)

			if tc.expectedErrCode > 0 {
				require.EqualError(t, err, "exit status 1")
			} else {
				require.NoError(t, err)
			}

			actualLogs := logBuffer.String()

			for _, expectedLogEntry := range tc.expectedLogs {
				assert.Regexp(t, regexp.MustCompile(expectedLogEntry), actualLogs)
			}

			fmt.Println(actualLogs)
		})
	}
}
