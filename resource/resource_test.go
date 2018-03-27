package resource

import (
	"testing"

	"reflect"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudetc/awsweeper/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	vpcId    = "some-vpc-id"
	tagKey   = "bla"
	tagValue = "blub"

	vpcs = []*ec2.Vpc{
		{
			VpcId: aws.String(vpcId),
			Tags: []*ec2.Tag{
				{
					Key:   aws.String(tagKey),
					Value: aws.String(tagValue),
				},
			},
		},
	}
)

func TestList_Vpc(t *testing.T) {
	apiDesc := mockVpc()

	res, _ := List(apiDesc)

	require.Equal(t, vpcId, res[0].Id)
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
	apiDesc := mockVpc()

	describeOut := invoke(apiDesc.DescribeFn, apiDesc.DescribeFnInput)
	actualId := describeOut.Elem().FieldByName("Vpcs").Index(0).Elem().FieldByName("VpcId").Elem().String()

	require.Equal(t, vpcId, actualId)
}

func TestFindSlice(t *testing.T) {
	apiDesc := mockVpc()

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
	apiDesc := mockVpc()
	desc := "input is not a struct"

	_, err := findSlice(apiDesc.DescribeOutputName[0], reflect.ValueOf(desc))

	require.Error(t, err)
}

func TestTags_Vpc(t *testing.T) {
	apiDesc := mockVpc()

	res, _ := List(apiDesc)

	require.Equal(t, tagValue, res[0].Tags[tagKey])
}

func mockVpc() ApiDesc {
	mockEC2 := &mocks.EC2API{}

	a := ApiDesc{
		"aws_vpc",
		[]string{"Vpcs"},
		"VpcId",
		mockEC2.DescribeVpcs,
		&ec2.DescribeVpcsInput{},
		filterGeneric,
	}

	mockResultFn := func(input *ec2.DescribeVpcsInput) *ec2.DescribeVpcsOutput {
		output := &ec2.DescribeVpcsOutput{}
		output.SetVpcs(vpcs)
		return output
	}

	mockEC2.On("DescribeVpcs", mock.MatchedBy(func(input *ec2.DescribeVpcsInput) bool {
		return true
	})).Return(mockResultFn, nil)

	return a
}

func mockInstance(rs []*ec2.Reservation) ApiDesc {
	mockEC2 := &mocks.EC2API{}

	a := ApiDesc{
		"aws_instance",
		[]string{"Reservations", "Instances"},
		"InstanceId",
		mockEC2.DescribeInstances,
		&ec2.DescribeInstancesInput{},
		filterGeneric,
	}

	mockResultFn := func(input *ec2.DescribeInstancesInput) *ec2.DescribeInstancesOutput {
		output := &ec2.DescribeInstancesOutput{}
		output.SetReservations(rs)
		return output
	}

	mockEC2.On("DescribeInstances", mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
		return true
	})).Return(mockResultFn, nil)

	return a
}
