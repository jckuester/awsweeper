package resource

import (
	"testing"

	"reflect"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/cloudetc/awsweeper/mocks"
	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	someVpcId = "some-vpc-id"
	tagKey    = "bla"
	tagValue  = "blub"

	otherVpcId = "other-vpc-id"

	vpcs = []*ec2.Vpc{
		{
			VpcId: aws.String(someVpcId),
			Tags: []*ec2.Tag{
				{
					Key:   aws.String(tagKey),
					Value: aws.String(tagValue),
				},
			},
		},
		{
			VpcId: aws.String(otherVpcId),
		},
	}
)

func TestList_Vpc(t *testing.T) {
	apiDesc := mockVpc(vpcs)

	res, _ := List(apiDesc)

	result := []string{}
	for _, r := range res {
		result = append(result, r.Id)
	}

	require.Len(t, res, 2)
	require.Contains(t, result, someVpcId)
	require.Contains(t, result, otherVpcId)
}

func TestList_NestedDescribeOutput(t *testing.T) {
	someExpectedId := "some-instance-id"
	otherExpectedId := "other-instance-id"

	rs := []*ec2.Reservation{
		{
			Instances: []*ec2.Instance{
				{
					InstanceId: aws.String(someExpectedId),
				},
			},
		},
		{
			Instances: []*ec2.Instance{
				{
					InstanceId: aws.String(otherExpectedId),
				},
			},
		},
	}
	apiDesc := mockInstance(rs)

	res, _ := List(apiDesc)

	result := []string{}
	for _, r := range res {
		result = append(result, r.Id)
	}

	require.Len(t, res, 2)
	require.Contains(t, result, otherExpectedId)
	require.Contains(t, result, someExpectedId)
}

func TestList_OnlyTerminatedInstances(t *testing.T) {
	// Filtering can not be tested via unit tests
	// (it happens on AWS server side)
	t.SkipNow()
	availInstanceId := "id-of-available-instance"
	termInstanceId := "id-of-terminated-instance"

	rs := []*ec2.Reservation{
		{
			Instances: []*ec2.Instance{
				{
					InstanceId: aws.String(termInstanceId),
					State: &ec2.InstanceState{
						Code: aws.Int64(48),
						Name: aws.String("terminated"),
					},
				},
				{
					InstanceId: aws.String(availInstanceId),
					State: &ec2.InstanceState{
						Code: aws.Int64(16),
						Name: aws.String("running"),
					},
				},
			},
		},
	}
	apiDesc := mockInstance(rs)

	res, _ := List(apiDesc)

	fmt.Println(res)

	require.Len(t, res, 1)
	require.Equal(t, availInstanceId, res[0].Id)
}

func TestInvoke(t *testing.T) {
	apiDesc := mockVpc(vpcs)

	describeOut := invoke(apiDesc.DescribeFn, apiDesc.DescribeFnInput)
	actualVpcs := describeOut.Elem().FieldByName("Vpcs")

	result := []string{}
	for i := 0; i < actualVpcs.Len(); i++ {
		actualId := actualVpcs.Index(i).Elem().FieldByName("VpcId").Elem().String()
		result = append(result, actualId)
	}

	require.Len(t, result, 2)
	require.Contains(t, result, someVpcId)
	require.Contains(t, result, otherVpcId)
}

func TestFindSlice(t *testing.T) {
	apiDesc := mockVpc(vpcs)

	desc := ec2.DescribeVpcsOutput{
		Vpcs: vpcs,
	}
	expectedId := vpcs[0].VpcId

	result, err := findSlice(apiDesc.DescribeOutputName[0], reflect.ValueOf(desc))
	actualId := result.Index(0).Elem().FieldByName("VpcId").Elem().String()

	require.Equal(t, *expectedId, actualId)
	require.NoError(t, err)
}

func TestFindSlice_InvalidInput(t *testing.T) {
	apiDesc := mockVpc(vpcs)
	desc := "input is not a struct"

	_, err := findSlice(apiDesc.DescribeOutputName[0], reflect.ValueOf(desc))

	require.Error(t, err)
}

func TestTags_Vpc(t *testing.T) {
	apiDesc := mockVpc(vpcs)

	res, _ := List(apiDesc)

	require.Equal(t, tagValue, res[0].Tags[tagKey])
}

func mockVpc(vpcs []*ec2.Vpc) ApiDesc {
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

	mockResultFn := func(input *ec2.DescribeVpcsInput) *ec2.DescribeVpcsOutput {
		output := &ec2.DescribeVpcsOutput{}
		output.SetVpcs(vpcs)
		return output
	}

	mockEC2.On("DescribeVpcs", mock.MatchedBy(func(input *ec2.DescribeVpcsInput) bool {
		return true
	})).Return(mockResultFn, nil)

	mockGetCallerIdentityFn := func(input *sts.GetCallerIdentityInput) *sts.GetCallerIdentityOutput {
		output := &sts.GetCallerIdentityOutput{}
		output.SetAccount("123456789")
		return output
	}

	mockSTS.On("GetCallerIdentity", mock.MatchedBy(func(input *sts.GetCallerIdentityInput) bool {
		return true
	})).Return(mockGetCallerIdentityFn, nil)

	a, err := getSupported("aws_vpc", c)
	if err != nil {
		log.Fatal(err)
	}

	return a
}

func mockInstance(rs []*ec2.Reservation) ApiDesc {
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

	mockResultFn := func(input *ec2.DescribeInstancesInput) *ec2.DescribeInstancesOutput {
		output := &ec2.DescribeInstancesOutput{}
		output.SetReservations(rs)
		return output
	}

	mockEC2.On("DescribeInstances", mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
		return true
	})).Return(mockResultFn, nil)

	mockGetCallerIdentityFn := func(input *sts.GetCallerIdentityInput) *sts.GetCallerIdentityOutput {
		output := &sts.GetCallerIdentityOutput{}
		output.SetAccount("123456789")
		return output
	}

	mockSTS.On("GetCallerIdentity", mock.MatchedBy(func(input *sts.GetCallerIdentityInput) bool {
		return true
	})).Return(mockGetCallerIdentityFn, nil)

	as, err := getSupported("aws_instance", c)
	if err != nil {
		log.Fatal(err)
	}

	return as
}
