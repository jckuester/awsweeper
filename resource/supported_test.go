package resource

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/cloudetc/awsweeper/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSupported(t *testing.T) {
	apiDesc, err := getSupported("aws_vpc", mockAWSClient())

	require.Equal(t, apiDesc.TerraformType, "aws_vpc")
	require.Equal(t, apiDesc.DeleteID, "VpcId")

	require.NoError(t, err)
}

func TestGetSupported_InvalidType(t *testing.T) {
	_, err := getSupported("some_type", mockAWSClient())

	require.Error(t, err)
}

func mockAWSClient() *AWSClient {
	mockAS := &mocks.AutoScalingAPI{}
	mockCF := &mocks.CloudFormationAPI{}
	mockEC2 := &mocks.EC2API{}
	mockEFS := &mocks.EFSAPI{}
	mockELB := &mocks.ELBAPI{}
	mockIAM := &mocks.IAMAPI{}
	mockKMS := &mocks.KMSAPI{}
	mockR53 := &mocks.Route53API{}
	mockS3 := &mocks.S3API{}
	mockSTS := &mocks.STSAPI{}

	c := &AWSClient{
		ASconn:  mockAS,
		CFconn:  mockCF,
		EC2conn: mockEC2,
		EFSconn: mockEFS,
		ELBconn: mockELB,
		IAMconn: mockIAM,
		KMSconn: mockKMS,
		R53conn: mockR53,
		S3conn:  mockS3,
		STSconn: mockSTS,
	}

	mockGetCallerIdentityFn := func(input *sts.GetCallerIdentityInput) *sts.GetCallerIdentityOutput {
		output := &sts.GetCallerIdentityOutput{}
		output.SetAccount("123456789")
		return output
	}

	mockSTS.On("GetCallerIdentity", mock.MatchedBy(func(input *sts.GetCallerIdentityInput) bool {
		return true
	})).Return(mockGetCallerIdentityFn, nil)

	return c
}
