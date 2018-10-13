package resource_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudetc/awsweeper/mocks"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAWS_DeletableResources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// given
	mockObj := mocks.NewMockAutoScalingAPI(mockCtrl)
	awsMock := &resource.AWS{
		AutoScalingAPI: mockObj,
	}
	// when
	res, err := awsMock.DeletableResources(resource.AutoscalingGroup, []*autoscaling.Group{
		{
			AutoScalingGroupName: &testAutoscalingGroupName,
			Tags:                 convertTags(testTags),
		},
	})
	require.NoError(t, err)

	require.Len(t, res, 1)
	require.Equal(t, res[0].ID, testAutoscalingGroupName)
	require.Equal(t, res[0].Tags, testTags)
}
