package resource_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/stretchr/testify/require"
)

func TestAWS_DeletableResources(t *testing.T) {
	// given
	rawResources := []*autoscaling.Group{
		{
			AutoScalingGroupName: &testAutoscalingGroupName,
			Tags:                 convertTags(testTags),
		},
	}

	// when
	res, err := resource.DeletableResources(resource.AutoscalingGroup, rawResources)
	require.NoError(t, err)

	// then
	require.Len(t, res, 1)
	require.Equal(t, testAutoscalingGroupName, res[0].ID)
	require.Equal(t, testTags, res[0].Tags)
}

func TestAWS_DeletableResources_Created(t *testing.T) {
	// given
	testLaunchTime := aws.Time(time.Date(2018, 11, 17, 5, 0, 0, 0, time.UTC))
	rawResources := []*ec2.Instance{
		{
			InstanceId: &testInstanceID,
			LaunchTime: testLaunchTime,
		},
	}

	// when
	res, err := resource.DeletableResources(resource.Instance, rawResources)
	require.NoError(t, err)

	// then
	require.Len(t, res, 1)
	require.Equal(t, testInstanceID, res[0].ID)
	require.Equal(t, testLaunchTime, res[0].Created)

}
