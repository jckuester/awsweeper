package resource_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudetc/awsweeper/mocks"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	securityGroupType = resource.SecurityGroup
	iamRoleType       = resource.IamRole
	instanceType      = resource.Instance

	yml = resource.YamlCfg{
		iamRoleType: {
			Ids: []*string{aws.String("^foo.*")},
		},
		securityGroupType: {},
		instanceType: {
			Tags: map[string]string{
				"foo": "bar",
				"bla": "blub",
			},
		},
		resource.AutoscalingGroup: {
			Ids: []*string{aws.String("^foo.*")},
			Tags: map[string]string{
				"foo": "bar",
			},
		},
	}

	f = &resource.YamlFilter{
		Cfg: yml,
	}
)

func TestYamlFilter_Apply_DoNotFilter(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// given
	mockObj := mocks.NewMockAutoScalingAPI(mockCtrl)
	awsMock := &resource.AWS{
		AutoScalingAPI: mockObj,
	}

	// when
	deletableResources := []*resource.DeletableResource{
		{
			Type: resource.AutoscalingGroup,
			ID:   "do-not-filter",
		},
	}

	// when
	filteredResources := f.Apply(resource.AutoscalingGroup, deletableResources, testAutoscalingGroup, awsMock)
	assert.Len(t, filteredResources[0], 0)
}

func TestYamlFilter_Apply(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// given
	mockObj := mocks.NewMockAutoScalingAPI(mockCtrl)
	awsMock := &resource.AWS{
		AutoScalingAPI: mockObj,
	}

	// when
	deletableResources := []*resource.DeletableResource{
		{
			Type: resource.AutoscalingGroup,
			ID:   "foo",
		},
	}

	// when
	filteredResources := f.Apply(resource.AutoscalingGroup, deletableResources, testAutoscalingGroup, awsMock)
	assert.Len(t, filteredResources[0], 1)
}
