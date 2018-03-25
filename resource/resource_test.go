package resource

import (
	"testing"

	"fmt"

	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudetc/awsweeper/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	mockEC2 = &mocks.EC2API{}

	apiInfoVpc = ApiDesc{
		"aws_vpc",
		[]string{"Vpcs"},
		"VpcId",
		mockEC2.DescribeVpcs,
		&ec2.DescribeVpcsInput{},
		filterGeneric,
	}

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

	apiInfoInstance = ApiDesc{
		"aws_instance",
		[]string{"Reservations", "Instances"},
		"InstanceId",
		mockEC2.DescribeInstances,
		&ec2.DescribeInstancesInput{},
		filterInstances,
	}

	instanceId = "some-instance-id"

	reservations = []*ec2.Reservation{
		{
			OwnerId: aws.String("bla"),
			Instances: []*ec2.Instance{
				{
					InstanceId: aws.String(instanceId),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String(tagKey),
							Value: aws.String(tagValue),
						},
					},
				},
			},
		},
	}
)

func TestList_Vpc(t *testing.T) {
	mockVpc()

	res, _ := List(apiInfoVpc)

	require.Equal(t, vpcId, res[0].Id)
}

func TestList_Instance(t *testing.T) {
	mockInstance()

	res, _ := List(apiInfoInstance)

	var id string
	if len(res) > 0 {
		id = res[0].Id
	}
	require.Equal(t, instanceId, id)
}

func TestInvoke(t *testing.T) {
	t.SkipNow()
	mockVpc()

	describeOut := invoke(apiInfoVpc.DescribeFn, apiInfoVpc.DescribeFnInput)
	fmt.Println(describeOut)
	fmt.Println(ec2.DescribeVpcsOutput{
		Vpcs: vpcs,
	})

	expected := ec2.DescribeVpcsOutput{
		Vpcs: vpcs,
	}

	require.True(t, reflect.DeepEqual(describeOut, expected))
}

func TestTags_Vpc(t *testing.T) {
	mockVpc()

	res, _ := List(apiInfoVpc)

	require.Equal(t, tagValue, res[0].Tags[tagKey])
}

func mockVpc() {
	mockResultFn := func(input *ec2.DescribeVpcsInput) *ec2.DescribeVpcsOutput {
		output := &ec2.DescribeVpcsOutput{}
		output.SetVpcs(vpcs)
		return output
	}

	mockEC2.On("DescribeVpcs", mock.MatchedBy(func(input *ec2.DescribeVpcsInput) bool {
		return true
	})).Return(mockResultFn, nil)
}

func mockInstance() {
	mockResultFn := func(input *ec2.DescribeInstancesInput) *ec2.DescribeInstancesOutput {
		output := &ec2.DescribeInstancesOutput{}
		output.SetReservations(reservations)
		return output
	}

	mockEC2.On("DescribeInstances", mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
		return true
	})).Return(mockResultFn, nil)
}

func TestFindSlice(t *testing.T) {
	desc := ec2.DescribeVpcsOutput{
		Vpcs: vpcs,
	}
	expectedId := vpcs[0].VpcId

	result, err := findSlice(apiInfoVpc.DescribeOutputName[0], reflect.ValueOf(desc))
	actualId := result.Index(0).Elem().FieldByName("VpcId").Elem().String()

	require.Equal(t, *expectedId, actualId)
	require.NoError(t, err)
}

func TestFindSlice_InvalidInput(t *testing.T) {
	desc := "input is not a struct"

	_, err := findSlice(apiInfoVpc.DescribeOutputName[0], reflect.ValueOf(desc))

	require.Error(t, err)
}
