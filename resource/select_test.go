package resource_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/stretchr/testify/assert"
)

func TestYamlFilter_Apply_EmptyConfig(t *testing.T) {
	//given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{},
	}
	res := []*resource.DeletableResource{
		{
			Type: resource.Instance,
			ID:   "foo",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result[0], 0)
}

func TestYamlFilter_Apply_FilterAll(t *testing.T) {
	//given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{
			resource.Instance: {},
		},
	}
	res := []*resource.DeletableResource{
		{
			Type: resource.Instance,
			ID:   "foo",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, len(res))
	assert.Equal(t, res[0].ID, result[0][0].ID)
}

func TestYamlFilter_Apply_FilterID(t *testing.T) {
	//given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{
			resource.Instance: {},
		},
	}
	res := []*resource.DeletableResource{
		{
			Type: resource.Instance,
			ID:   "foo",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, len(res))
	assert.Equal(t, res[0].ID, result[0][0].ID)
}

func TestYamlFilter_Apply_FilterByID(t *testing.T) {
	//given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{
			resource.Instance: {
				Ids: []*string{aws.String("^select")},
			},
		},
	}

	// when
	res := []*resource.DeletableResource{
		{
			Type: resource.Instance,
			ID:   "select-this",
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)
	assert.Len(t, result[0], 1)
	assert.Equal(t, "select-this", result[0][0].ID)
}

func TestYamlFilter_Apply_FilterByTag(t *testing.T) {
	//given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{
			resource.Instance: {
				Tags: map[string]string{
					"foo": "^bar",
				}},
		},
	}

	// when
	res := []*resource.DeletableResource{
		{
			Type: resource.Instance,
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this-either",
			Tags: map[string]string{
				"foo": "blub",
			},
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)
	assert.Len(t, result[0], 1)
	assert.Equal(t, "select-this", result[0][0].ID)
}

func TestYamlFilter_Apply_FilterByIDandTag(t *testing.T) {
	//given
	f := &resource.YamlFilter{
		Cfg: resource.YamlCfg{
			resource.Instance: {
				Ids: []*string{aws.String("^foo")},
				Tags: map[string]string{
					"foo": "^bar",
				}},
		},
	}

	// when
	res := []*resource.DeletableResource{
		{
			Type: resource.Instance,
			ID:   "foo",
			Tags: map[string]string{
				"foo": "blub",
			},
		},
		{
			Type: resource.Instance,
			ID:   "bar",
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)
	assert.Len(t, result[0], 2)
	assert.Equal(t, "foo", result[0][0].ID)
	assert.Equal(t, "bar", result[0][1].ID)

}
