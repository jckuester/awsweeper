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

func TestDeletableResources(t *testing.T) {
	// given
	rawResources := []*autoscaling.Group{
		{
			AutoScalingGroupName: &testAutoscalingGroupName,
			Tags:                 convertTags(testTags),
		},
	}

	// when
	res, err := resource.DeletableResources(testRegion, resource.AutoscalingGroup, rawResources)
	require.NoError(t, err)

	// then
	require.Len(t, res, 1)
	require.Equal(t, testAutoscalingGroupName, res[0].ID)
	require.Equal(t, testRegion, res[0].Region)
	require.Equal(t, testTags, res[0].Tags)
}

func TestDeletableResources_Created(t *testing.T) {
	// given
	testLaunchTime := aws.Time(time.Date(2018, 11, 17, 5, 0, 0, 0, time.UTC))
	rawResources := []*ec2.Instance{
		{
			InstanceId: &testInstanceID,
			LaunchTime: testLaunchTime,
		},
	}

	// when
	res, err := resource.DeletableResources(testRegion, resource.Instance, rawResources)
	require.NoError(t, err)

	// then
	require.Len(t, res, 1)
	require.Equal(t, testInstanceID, res[0].ID)
	require.Equal(t, testLaunchTime, res[0].Created)

}

func TestDeletableResources_CreatedFieldIsTypeString(t *testing.T) {
	// given
	testCreationDate := "2018-12-16T19:40:28.000Z"
	rawResources := []*ec2.Image{
		{
			ImageId:      &testImageId,
			CreationDate: &testCreationDate,
		},
	}

	// when
	res, err := resource.DeletableResources(testRegion, resource.Ami, rawResources)
	require.NoError(t, err)

	// then
	require.Len(t, res, 1)
	require.Equal(t, testCreationDate, res[0].Created.Format("2006-01-02T15:04:05.000Z0700"))
}
