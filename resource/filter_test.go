package resource_test

import (
	"testing"
	"time"

	"github.com/cloudetc/awsweeper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"
)

func TestYamlFilter_Validate(t *testing.T) {
	// given
	f := &resource.Filter{
		Cfg: resource.Config{
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
	f := &resource.Filter{
		Cfg: resource.Config{},
	}

	// when
	err := f.Validate()

	// then
	assert.NoError(t, err)
}

func TestYamlFilter_Validate_UnsupportedType(t *testing.T) {
	// given
	f := &resource.Filter{
		Cfg: resource.Config{
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
	f := &resource.Filter{
		Cfg: resource.Config{
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

func TestYamlFilter_Types_DependencyOrder(t *testing.T) {
	// given
	f := &resource.Filter{
		Cfg: resource.Config{
			resource.Subnet: {},
			resource.Vpc:    {},
		},
	}

	// when
	resTypes := f.Types()

	// then
	assert.Len(t, resTypes, 2)
	assert.Equal(t, resTypes[0], resource.Subnet)
	assert.Equal(t, resTypes[1], resource.Vpc)
}

func Test_ParseFile(t *testing.T) {
	input := []byte(`aws_instance:
  - id: ^foo.*
    created:
      before: 5d
      after: 2018-10-28 12:28:39
  - id: ^foo.*`)
    created:
      before: 23h`)

	var cfg resource.Config
	err := yaml.UnmarshalStrict(input, &cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg[resource.Instance])
	require.Len(t, cfg[resource.Instance], 2)
	require.NotNil(t, cfg[resource.Instance][0].ID)
	assert.Equal(t, "^foo.*", cfg[resource.Instance][0].ID.Pattern)
	assert.True(t, cfg[resource.Instance][0].ID.Negate)
	require.NotNil(t, cfg[resource.Instance][1].ID)
	require.NotNil(t, cfg[resource.Instance][0].Created.Before)
	assert.True(t, cfg[resource.Instance][0].Created.Before.Before(time.Now().UTC().AddDate(0, 0, -4)))
	assert.True(t, cfg[resource.Instance][0].Created.Before.After(time.Now().UTC().AddDate(0, 0, -6)))
	require.NotNil(t, cfg[resource.Instance][0].Created.After)
	assert.Equal(t, resource.CreatedTime{Time: time.Date(2018, 10, 28, 12, 28, 39, 0000, time.UTC)}, *cfg[resource.Instance][0].Created.After)
	assert.Equal(t, "^foo.*", cfg[resource.Instance][1].ID.Pattern)
	assert.False(t, cfg[resource.Instance][1].ID.Negate)
	require.NotNil(t, cfg[resource.Instance][1].Created.Before)
	assert.True(t, cfg[resource.Instance][1].Created.Before.Before(time.Now().UTC().Add(-22 * time.Hour)))
	assert.True(t, cfg[resource.Instance][1].Created.Before.After(time.Now().UTC().Add(-24 * time.Hour)))
	require.Nil(t, cfg[resource.Instance][1].Created.After)
}