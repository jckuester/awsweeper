package resource_test

import (
	"testing"

	"github.com/cloudetc/awsweeper/resource"
	"github.com/stretchr/testify/assert"
)

func TestYamlFilter_Validate(t *testing.T) {
	// given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{
			resource.IamRole:       {},
			resource.SecurityGroup: {},
			resource.Instance:      {},
			resource.Vpc:           {},
		},
	}

	// when
	err := f.Validate()

	// then
	assert.NoError(t, err)
}

func TestYamlFilter_Validate_EmptyConfig(t *testing.T) {
	// given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{},
	}

	// when
	err := f.Validate()

	// then
	assert.NoError(t, err)
}

func TestYamlFilter_Validate_UnsupportedType(t *testing.T) {
	// given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{
			resource.Instance:    {},
			"not_supported_type": {},
		},
	}

	// when
	err := f.Validate()

	// then
	assert.EqualError(t, err, "unsupported resource type found in yaml config: not_supported_type")
}

func TestYamlFilter_Types(t *testing.T) {
	// given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{
			resource.Instance: {},
			resource.Vpc:      {},
		},
	}

	// when
	resTypes := f.Types()

	// then
	assert.Len(t, resTypes, 2)
	assert.Contains(t, resTypes, resource.Vpc)
	assert.Contains(t, resTypes, resource.Instance)
}
