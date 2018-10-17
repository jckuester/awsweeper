package resource_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/stretchr/testify/require"
)

func TestAWS_DeletableResources(t *testing.T) {
	aws := &resource.AWS{}

	// given
	rawResources := []*autoscaling.Group{
		{
			AutoScalingGroupName: &testAutoscalingGroupName,
			Tags:                 convertTags(testTags),
		},
	}

	// when
	res, err := aws.DeletableResources(resource.AutoscalingGroup, rawResources)
	require.NoError(t, err)

	// then
	require.Len(t, res, 1)
	require.Equal(t, res[0].ID, testAutoscalingGroupName)
	require.Equal(t, res[0].Tags, testTags)
}
