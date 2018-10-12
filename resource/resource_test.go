package resource

//
//import (
//	"testing"
//
//	"reflect"
//
//	"fmt"
//
//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/service/ec2"
//	"github.com/aws/aws-sdk-go/service/sts"
//	"github.com/cloudetc/awsweeper/mocks"
//	"github.com/prometheus/common/log"
//	"github.com/stretchr/testify/mock"
//	"github.com/stretchr/testify/require"
//)
//
//var (
//	someVpcID = "some-vpc-id"
//	tagKey    = "bla"
//	tagValue  = "blub"
//
//	otherVpcID = "other-vpc-id"
//
//	vpcs = []*ec2.Vpc{
//		{
//			VpcId: aws.String(someVpcID),
//			Tags: []*ec2.Tag{
//				{
//					Key:   aws.String(tagKey),
//					Value: aws.String(tagValue),
//				},
//			},
//		},
//		{
//			VpcId: aws.String(otherVpcID),
//		},
//	}
//)
//
//func TestList_Vpc(t *testing.T) {
//	apiDesc := mockVpc(vpcs)
//
//	res, _ := DeletableResource(apiDesc)
//
//	result := []string{}
//	for _, r := range res {
//		result = append(result, r.ID)
//	}
//
//	require.Len(t, res, 2)
//	require.Contains(t, result, someVpcID)
//	require.Contains(t, result, otherVpcID)
//}
//
//func TestList_NestedDescribeOutput(t *testing.T) {
//	someExpectedID := "some-instance-id"
//	otherExpectedID := "other-instance-id"
//
//	rs := []*ec2.Reservation{
//		{
//			Instances: []*ec2.Instance{
//				{
//					InstanceId: aws.String(someExpectedID),
//				},
//			},
//		},
//		{
//			Instances: []*ec2.Instance{
//				{
//					InstanceId: aws.String(otherExpectedID),
//				},
//			},
//		},
//	}
//	apiDesc := mockInstance(rs)
//
//	res, _ := DeletableResource(apiDesc)
//
//	result := []string{}
//	for _, r := range res {
//		result = append(result, r.ID)
//	}
//
//	require.Len(t, res, 2)
//	require.Contains(t, result, otherExpectedID)
//	require.Contains(t, result, someExpectedID)
//}
//
//func TestList_OnlyTerminatedInstances(t *testing.T) {
//	// Filtering can not be tested via unit tests
//	// (it happens on AWS server side)
//	t.SkipNow()
//	availInstanceID := "id-of-available-instance"
//	termInstanceID := "id-of-terminated-instance"
//
//	rs := []*ec2.Reservation{
//		{
//			Instances: []*ec2.Instance{
//				{
//					InstanceId: aws.String(termInstanceID),
//					State: &ec2.InstanceState{
//						Code: aws.Int64(48),
//						Name: aws.String("terminated"),
//					},
//				},
//				{
//					InstanceId: aws.String(availInstanceID),
//					State: &ec2.InstanceState{
//						Code: aws.Int64(16),
//						Name: aws.String("running"),
//					},
//				},
//			},
//		},
//	}
//	apiDesc := mockInstance(rs)
//
//	res, _ := DeletableResource(apiDesc)
//
//	fmt.Println(res)
//
//	require.Len(t, res, 1)
//	require.Equal(t, availInstanceID, res[0].ID)
//}
//
//func TestInvoke(t *testing.T) {
//	apiDesc := mockVpc(vpcs)
//
//	describeOut := invoke(apiDesc.Describe, apiDesc.DescribeInput)
//	actualVpcs := describeOut.Elem().FieldByName("Vpcs")
//
//	result := []string{}
//	for i := 0; i < actualVpcs.Len(); i++ {
//		actualID := actualVpcs.Index(i).Elem().FieldByName("VpcId").Elem().String()
//		result = append(result, actualID)
//	}
//
//	require.Len(t, result, 2)
//	require.Contains(t, result, someVpcID)
//	require.Contains(t, result, otherVpcID)
//}
//
//func TestFindSlice(t *testing.T) {
//	apiDesc := mockVpc(vpcs)
//
//	desc := ec2.DescribeVpcsOutput{
//		Vpcs: vpcs,
//	}
//	expectedID := vpcs[0].VpcId
//
//	result, err := findSlice(apiDesc.DescribeOutputName[0], reflect.ValueOf(desc))
//	actualID := result.Index(0).Elem().FieldByName("VpcId").Elem().String()
//
//	require.Equal(t, *expectedID, actualID)
//	require.NoError(t, err)
//}
//
//func TestFindSlice_InvalidInput(t *testing.T) {
//	apiDesc := mockVpc(vpcs)
//	desc := "input is not a struct"
//
//	_, err := findSlice(apiDesc.DescribeOutputName[0], reflect.ValueOf(desc))
//
//	require.Error(t, err)
//}
//
//func TestTags_Vpc(t *testing.T) {
//	apiDesc := mockVpc(vpcs)
//
//	res, _ := DeletableResource(apiDesc)
//
//	require.Equal(t, tagValue, res[0].Tags[tagKey])
//}
//
//func mockVpc(vpcs []*ec2.Vpc) APIDesc {
//	mockAS := &mocks.AutoScalingAPI{}
//	mockCF := &mocks.CloudFormationAPI{}
//	mockEC2 := &mocks.EC2API{}
//	mockEFS := &mocks.EFSAPI{}
//	mockELB := &mocks.ELBAPI{}
//	mockIAM := &mocks.IAMAPI{}
//	mockKMS := &mocks.KMSAPI{}
//	mockR53 := &mocks.Route53API{}
//	mockS3 := &mocks.S3API{}
//	mockSTS := &mocks.STSAPI{}
//
//	c := &AWS{
//		ASconn:  mockAS,
//		CFconn:  mockCF,
//		EC2conn: mockEC2,
//		EFSconn: mockEFS,
//		ELBconn: mockELB,
//		IAMconn: mockIAM,
//		KMSconn: mockKMS,
//		R53conn: mockR53,
//		S3conn:  mockS3,
//		STSconn: mockSTS,
//	}
//
//	mockResultFn := func(input *ec2.DescribeVpcsInput) *ec2.DescribeVpcsOutput {
//		output := &ec2.DescribeVpcsOutput{}
//		output.SetVpcs(vpcs)
//		return output
//	}
//
//	mockEC2.On("DescribeVpcs", mock.MatchedBy(func(input *ec2.DescribeVpcsInput) bool {
//		return true
//	})).Return(mockResultFn, nil)
//
//	mockGetCallerIdentityFn := func(input *sts.GetCallerIdentityInput) *sts.GetCallerIdentityOutput {
//		output := &sts.GetCallerIdentityOutput{}
//		output.SetAccount("123456789")
//		return output
//	}
//
//	mockSTS.On("GetCallerIdentity", mock.MatchedBy(func(input *sts.GetCallerIdentityInput) bool {
//		return true
//	})).Return(mockGetCallerIdentityFn, nil)
//
//	a, err := getSupported("aws_vpc", c)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return a
//}
//
//func mockInstance(rs []*ec2.Reservation) APIDesc {
//	mockAS := &mocks.AutoScalingAPI{}
//	mockCF := &mocks.CloudFormationAPI{}
//	mockEC2 := &mocks.EC2API{}
//	mockEFS := &mocks.EFSAPI{}
//	mockELB := &mocks.ELBAPI{}
//	mockIAM := &mocks.IAMAPI{}
//	mockKMS := &mocks.KMSAPI{}
//	mockR53 := &mocks.Route53API{}
//	mockS3 := &mocks.S3API{}
//	mockSTS := &mocks.STSAPI{}
//
//	c := &AWS{
//		ASconn:  mockAS,
//		CFconn:  mockCF,
//		EC2conn: mockEC2,
//		EFSconn: mockEFS,
//		ELBconn: mockELB,
//		IAMconn: mockIAM,
//		KMSconn: mockKMS,
//		R53conn: mockR53,
//		S3conn:  mockS3,
//		STSconn: mockSTS,
//	}
//
//	mockResultFn := func(input *ec2.DescribeInstancesInput) *ec2.DescribeInstancesOutput {
//		output := &ec2.DescribeInstancesOutput{}
//		output.SetReservations(rs)
//		return output
//	}
//
//	mockEC2.On("DescribeInstances", mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
//		return true
//	})).Return(mockResultFn, nil)
//
//	mockGetCallerIdentityFn := func(input *sts.GetCallerIdentityInput) *sts.GetCallerIdentityOutput {
//		output := &sts.GetCallerIdentityOutput{}
//		output.SetAccount("123456789")
//		return output
//	}
//
//	mockSTS.On("GetCallerIdentity", mock.MatchedBy(func(input *sts.GetCallerIdentityInput) bool {
//		return true
//	})).Return(mockGetCallerIdentityFn, nil)
//
//	as, err := getSupported("aws_instance", c)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return as
//}
