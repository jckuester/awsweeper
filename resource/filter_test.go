package resource_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/cloudetc/awsweeper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"
)

func TestYamlFilter_Validate(t *testing.T) {
	// given
	f := &resource.Filter{
		resource.IamRole:       {},
		resource.SecurityGroup: {},
		resource.Instance:      {},
		resource.Vpc:           {},
	}

	// when
	err := f.Validate()

	// then
	assert.NoError(t, err)
}

func TestYamlFilter_Validate_EmptyConfig(t *testing.T) {
	// given
	f := &resource.Filter{}

	// when
	err := f.Validate()

	// then
	assert.NoError(t, err)
}

func TestYamlFilter_Validate_UnsupportedType(t *testing.T) {
	// given
	f := &resource.Filter{
		resource.Instance:    {},
		"not_supported_type": {},
	}

	// when
	err := f.Validate()

	// then
	assert.EqualError(t, err, "unsupported resource type found in yaml config: not_supported_type")
}

func TestYamlFilter_Types(t *testing.T) {
	// given
	f := &resource.Filter{
		resource.Instance: {},
		resource.Vpc:      {},
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
		resource.Subnet: {},
		resource.Vpc:    {},
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
  - id: NOT(^foo.*)
    created:
      before: 5d
      after: 2018-10-28 12:28:39
  - id: ^foo.*
    created:
      before: 23h`)

	var cfg resource.Filter
	err := yaml.UnmarshalStrict(input, &cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg[resource.Instance])
	require.Len(t, cfg[resource.Instance], 2)
	require.NotNil(t, cfg[resource.Instance][0].ID)
	assert.Equal(t, "^foo.*", cfg[resource.Instance][0].ID.Pattern)
	assert.True(t, cfg[resource.Instance][0].ID.Negate)
	require.NotNil(t, cfg[resource.Instance][0].Created.Before)
	assert.True(t, cfg[resource.Instance][0].Created.Before.Before(time.Now().UTC().AddDate(0, 0, -4)))
	assert.True(t, cfg[resource.Instance][0].Created.Before.After(time.Now().UTC().AddDate(0, 0, -6)))
	require.NotNil(t, cfg[resource.Instance][0].Created.After)
	assert.Equal(t, resource.CreatedTime{Time: time.Date(2018, 10, 28, 12, 28, 39, 0000, time.UTC)}, *cfg[resource.Instance][0].Created.After)
	require.NotNil(t, cfg[resource.Instance][1].ID)
	assert.Equal(t, "^foo.*", cfg[resource.Instance][1].ID.Pattern)
	assert.False(t, cfg[resource.Instance][1].ID.Negate)
	require.NotNil(t, cfg[resource.Instance][1].Created.Before)
	assert.True(t, cfg[resource.Instance][1].Created.Before.Before(time.Now().UTC().Add(-22*time.Hour)))
	assert.True(t, cfg[resource.Instance][1].Created.Before.After(time.Now().UTC().Add(-24*time.Hour)))
	require.Nil(t, cfg[resource.Instance][1].Created.After)
}

func TestTypeFilter_MatchTags_Tagged(t *testing.T) {
	tests := []struct {
		name   string
		filter resource.TypeFilter
		tags   map[string]string
		want   bool
	}{
		{
			name:   "no tagged filter, resource has tags",
			filter: resource.TypeFilter{},
			tags:   map[string]string{"foo": "bar"},
			want:   true,
		},
		{
			name:   "no tagged filter, resource has no tags",
			filter: resource.TypeFilter{},
			want:   true,
		},
		{
			name: "tagged filter, resource has tags",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(true),
			},
			tags: map[string]string{"foo": "bar"},
			want: true,
		},
		{
			name: "tagged filter, resource has no tags",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(true),
			},
			want: false,
		},
		{
			name: "untagged filter, resource has tags",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(false),
			},
			tags: map[string]string{"foo": "bar"},
			want: false,
		},
		{
			name: "untagged filter, resource has no tags",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(false),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.MatchTags(tt.tags); got != tt.want {
				t.Errorf("MatchTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeFilter_MatchTags(t *testing.T) {
	tests := []struct {
		name   string
		filter resource.TypeFilter
		tags   map[string]string
		want   bool
	}{
		{
			name: "no matching key",
			filter: resource.TypeFilter{
				Tags: map[string]*resource.StringFilter{
					"foo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foz": "bar"},
			want: false,
		},
		{
			name: "matching key, but not value",
			filter: resource.TypeFilter{
				Tags: map[string]*resource.StringFilter{
					"foo": {Pattern: "^bar"},
				},
			},
			tags: map[string]string{"foo": "baz"},
			want: false,
		},
		{
			name: "matching key and value",
			filter: resource.TypeFilter{
				Tags: map[string]*resource.StringFilter{
					"foo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar"},
			want: true,
		},
		{
			name: "untagged filter, matching key and value",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(false),
				Tags: map[string]*resource.StringFilter{
					"foo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar"},
			want: false,
		},
		{
			name: "tagged filter, matching key and value",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(true),
				Tags: map[string]*resource.StringFilter{
					"foo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.MatchTags(tt.tags); got != tt.want {
				t.Errorf("MatchTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
