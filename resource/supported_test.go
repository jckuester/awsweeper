package resource_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/cloudetc/awsweeper/resource/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testImageId = "test-ami"
	testAmi     = &ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			{
				ImageId: &testImageId,
			},
		},
	}

	testInstanceID = "test-instance"
	testInstance   = &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
					{
						ImageId: &testInstanceID,
					},
				},
			},
		},
	}

	testRegion               = "us-east-1"
	testAutoscalingGroupName = "test-auto-scaling-group"
	testTags                 = map[string]string{
		"test-tag-key": "test-tag-value",
	}
	testAutoscalingGroup = &autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []*autoscaling.Group{
			{
				AutoScalingGroupName: &testAutoscalingGroupName,
				Tags:                 convertTags(testTags),
			},
		},
	}

	testLaunchConfigurationName = "test-launch-configuration-name"
	testLaunchConfiguration     = &autoscaling.DescribeLaunchConfigurationsOutput{
		LaunchConfigurations: []*autoscaling.LaunchConfiguration{
			{
				LaunchConfigurationName: &testLaunchConfigurationName,
			},
		},
	}
)

func TestAWS_Resources_Amis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// given
	awsMock := createAmiMock(mockCtrl)

	// when
	resources, err := awsMock.RawResources(resource.Ami)
	require.NoError(t, err)
	res := resources.([]*ec2.Image)

	// then
	assert.Len(t, res, 1)
	assert.Equal(t, *res[0].ImageId, testImageId)
}

func TestAWS_Resources_AutoScalingGroups(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// given
	awsMock := createAutoScalingGroupMock(mockCtrl)

	// when
	resources, err := awsMock.RawResources(resource.AutoscalingGroup)
	require.NoError(t, err)
	groups := resources.([]*autoscaling.Group)

	// then
	assert.Len(t, groups, 1)
	assert.Equal(t, *groups[0].AutoScalingGroupName, testAutoscalingGroupName)
}

func TestAWS_Resources_LaunchConfigurations(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// given
	awsMock := createLaunchConfigurationMock(mockCtrl)

	// when
	resources, err := awsMock.RawResources(resource.LaunchConfiguration)
	require.NoError(t, err)
	lc := resources.([]*autoscaling.LaunchConfiguration)

	// then
	assert.Len(t, lc, 1)
	assert.Equal(t, *lc[0].LaunchConfigurationName, testLaunchConfigurationName)
}

func createAmiMock(mockCtrl *gomock.Controller) *resource.AWS {
	mockObj := mocks.NewMockEC2API(mockCtrl)
	mockObjSts := mocks.NewMockSTSAPI(mockCtrl)

	awsMock := &resource.AWS{
		EC2API: mockObj,
		STSAPI: mockObjSts,
	}

	mockObj.EXPECT().DescribeImages(gomock.Any()).Return(
		testAmi, nil)

	mockObjSts.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Return(
		&sts.GetCallerIdentityOutput{
			Account: aws.String("123456789"),
		}, nil)

	return awsMock
}

func createAutoScalingGroupMock(mockCtrl *gomock.Controller) *resource.AWS {
	mockObj := mocks.NewMockAutoScalingAPI(mockCtrl)
	awsMock := &resource.AWS{
		AutoScalingAPI: mockObj,
	}

	mockObj.EXPECT().DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{}).Return(
		testAutoscalingGroup, nil)

	return awsMock
}

func createLaunchConfigurationMock(mockCtrl *gomock.Controller) *resource.AWS {
	mockObj := mocks.NewMockAutoScalingAPI(mockCtrl)
	awsMock := &resource.AWS{
		AutoScalingAPI: mockObj,
	}

	mockObj.EXPECT().DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{}).Return(
		testLaunchConfiguration, nil)

	return awsMock
}

func convertTags(tags map[string]string) []*autoscaling.TagDescription {
	var tagDescriptions = make([]*autoscaling.TagDescription, 0, len(tags))

	for key, value := range tags {
		tagDescriptions = append(tagDescriptions, &autoscaling.TagDescription{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return tagDescriptions
}
