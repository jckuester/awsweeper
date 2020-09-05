package resource_test

import (
	"testing"
	"time"

	awsls "github.com/jckuester/awsls/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/jckuester/awsweeper/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testImageId              = "test-ami"
	testAutoscalingGroupName = "test-auto-scaling-group"
)

func TestDeletableResources_CreationDateIsTypeTime(t *testing.T) {
	// given
	testCreationDate := aws.Time(time.Date(2018, 11, 17, 5, 0, 0, 0, time.UTC))
	rawResources := []*autoscaling.AutoScalingGroup{
		{
			AutoScalingGroupName: &testAutoscalingGroupName,
			CreatedTime:          testCreationDate,
		},
	}

	// when
	res, err := resource.DeletableResources(resource.AutoscalingGroup, rawResources, awsls.Client{})
	require.NoError(t, err)
	require.Len(t, res, 1)

	// then
	assert.Equal(t, testAutoscalingGroupName, res[0].ID)
	assert.Equal(t, testCreationDate, res[0].CreatedAt)
}

func TestDeletableResources_CreationDateIsTypeString(t *testing.T) {
	// given
	testCreationDate := "2018-12-16T19:40:28.000Z"
	rawResources := []*ec2.Image{
		{
			ImageId:      &testImageId,
			CreationDate: &testCreationDate,
		},
	}

	// when
	res, err := resource.DeletableResources(resource.Ami, rawResources, awsls.Client{})
	require.NoError(t, err)

	// then
	require.Len(t, res, 1)
	require.Equal(t, testCreationDate, res[0].CreatedAt.Format("2006-01-02T15:04:05.000Z0700"))
}
