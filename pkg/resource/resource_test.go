package resource_test

import (
	"testing"
	"time"

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
	testTags                 = map[string]string{
		"test-tag-key": "test-tag-value",
	}
)

func TestDeletableResources_CreationDateIsTypeTime(t *testing.T) {
	// given
	testLaunchTime := aws.Time(time.Date(2018, 11, 17, 5, 0, 0, 0, time.UTC))

	rawResources := []*autoscaling.AutoScalingGroup{
		{
			AutoScalingGroupName: &testAutoscalingGroupName,
			Tags:                 convertTags(testTags),
			CreatedTime:          testLaunchTime,
		},
	}

	// when
	res, err := resource.DeletableResources(resource.AutoscalingGroup, rawResources)
	require.NoError(t, err)
	require.Len(t, res, 1)

	// then
	assert.Equal(t, testAutoscalingGroupName, res[0].ID)
	assert.Equal(t, testTags, res[0].Tags)
	assert.Equal(t, testLaunchTime, res[0].CreatedAt)
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
	res, err := resource.DeletableResources(resource.Ami, rawResources)
	require.NoError(t, err)

	// then
	require.Len(t, res, 1)
	require.Equal(t, testCreationDate, res[0].CreatedAt.Format("2006-01-02T15:04:05.000Z0700"))
}

func convertTags(tags map[string]string) []autoscaling.TagDescription {
	var tagDescriptions = make([]autoscaling.TagDescription, 0, len(tags))

	for key, value := range tags {
		tagDescriptions = append(tagDescriptions, autoscaling.TagDescription{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return tagDescriptions
}
