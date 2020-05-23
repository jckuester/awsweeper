package resource_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudetc/awsweeper/pkg/resource"
	awsls "github.com/jckuester/awsls/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYamlFilter_Apply_EmptyConfig(t *testing.T) {
	//given
	f := &resource.Filter{}

	res := []awsls.Resource{
		{
			Type: resource.Instance,
			ID:   "foo",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, 0)
}

func TestYamlFilter_Apply_FilterAll(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {},
	}
	res := []awsls.Resource{
		{
			Type: resource.Instance,
			ID:   "foo",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, len(res))
	assert.Equal(t, res[0].ID, result[0].ID)
}

func TestYamlFilter_Apply_FilterByID(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				ID: &resource.StringFilter{Pattern: "^select"},
			},
		},
	}

	// when
	res := []awsls.Resource{
		{
			Type: resource.Instance,
			ID:   "select-this",
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this",
		},
	}

	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	require.Len(t, result, 1)
	assert.Equal(t, "select-this", result[0].ID)
}

func TestYamlFilter_Apply_FilterByTag(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar"},
				},
			},
		},
	}

	res := []awsls.Resource{
		{
			Type: resource.Instance,
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this",
			Tags: map[string]string{
				"foo": "blub",
			},
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this-either",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	require.Len(t, result, 1)
	assert.Equal(t, "select-this", result[0].ID)
}

func TestYamlFilter_Apply_FilterByMultipleTags(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar"},
					"bla": {Pattern: "^blub"},
				},
			},
		},
	}

	res := []awsls.Resource{
		{
			Type: resource.Instance,
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
				"bla": "blub",
			},
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "select-this", result[0].ID)
}

func TestYamlFilter_Apply_FilterByIDandTag(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				ID: &resource.StringFilter{Pattern: "^foo"},
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar"},
				},
			},
		},
	}

	res := []awsls.Resource{
		{
			Type: resource.Instance,
			ID:   "foo",
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this",
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		{
			Type: resource.Instance,
			ID:   "this-neither",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "foo", result[0].ID)
}

func TestYamlFilter_Apply_Created(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				Created: &resource.Created{
					After:  &resource.CreatedTime{Time: time.Date(2018, 11, 17, 0, 0, 0, 0, time.UTC)},
					Before: &resource.CreatedTime{Time: time.Date(2018, 11, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	res := []awsls.Resource{
		{
			Type:      resource.Instance,
			ID:        "foo",
			CreatedAt: aws.Time(time.Date(2018, 11, 17, 5, 0, 0, 0, time.UTC)),
		},
		{
			Type:      resource.Instance,
			ID:        "do-not-select-this1",
			CreatedAt: aws.Time(time.Date(2018, 11, 17, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type:      resource.Instance,
			ID:        "do-not-select-this2",
			CreatedAt: aws.Time(time.Date(2018, 11, 20, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type:      resource.Instance,
			ID:        "do-not-select-this3",
			CreatedAt: aws.Time(time.Date(2018, 11, 22, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this2",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "foo", result[0].ID)
}

func TestYamlFilter_Apply_CreatedBefore(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				Created: &resource.Created{
					Before: &resource.CreatedTime{Time: time.Date(2018, 11, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	res := []awsls.Resource{
		{
			Type:      resource.Instance,
			ID:        "foo",
			CreatedAt: aws.Time(time.Date(2018, 11, 17, 5, 0, 0, 0, time.UTC)),
		},
		{
			Type:      resource.Instance,
			ID:        "do-not-select-this",
			CreatedAt: aws.Time(time.Date(2018, 11, 22, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this2",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "foo", result[0].ID)
}

func TestYamlFilter_Apply_CreatedAfter(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				Created: &resource.Created{
					After: &resource.CreatedTime{Time: time.Date(2018, 11, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	res := []awsls.Resource{
		{
			Type:      resource.Instance,
			ID:        "foo",
			CreatedAt: aws.Time(time.Date(2018, 11, 22, 5, 0, 0, 0, time.UTC)),
		},
		{
			Type:      resource.Instance,
			ID:        "do-not-select-this",
			CreatedAt: aws.Time(time.Date(2018, 11, 17, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this2",
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "foo", result[0].ID)
}

func TestYamlFilter_Apply_MultipleFiltersPerResourceType(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				ID: &resource.StringFilter{Pattern: "^select"},
			},
			{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar"},
				},
			},
		},
	}

	res := []awsls.Resource{
		{
			Type: resource.Instance,
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
		{
			Type: resource.Instance,
			ID:   "select-this-too",
			Tags: map[string]string{
				"bla": "blub",
			},
		},
		{
			Type: resource.Instance,
			ID:   "do-not-select-this",
			Tags: map[string]string{
				"bla": "blub",
			},
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	require.Len(t, result, 2)
	assert.Equal(t, "select-this", result[0].ID)
	assert.Equal(t, "select-this-too", result[1].ID)
}

func TestYamlFilter_Apply_NegatedStringFilter(t *testing.T) {
	//given
	f := &resource.Filter{
		resource.Instance: {
			{
				ID: &resource.StringFilter{Pattern: "^select", Negate: true},
			},
			{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar", Negate: true},
				},
			},
		},
	}

	res := []awsls.Resource{
		{
			Type: resource.Instance,
			ID:   "select-this-not",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
		{
			Type: resource.Instance,
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "baz",
			},
		},
	}

	// when
	result := f.Apply(resource.Instance, res, testInstance, nil)

	// then
	require.Len(t, result, 1)
	assert.Equal(t, "select-this", result[0].ID)
}
