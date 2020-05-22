package resource_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/cloudetc/awsweeper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"
)

func TestFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		f       resource.Filter
		wantErr string
	}{
		{
			name: "empty filter",
			f:    resource.Filter{},
		},
		{
			name: "unsupported type",
			f: resource.Filter{
				resource.Instance:    {},
				"not_supported_type": {},
			},
			wantErr: "unsupported resource type: not_supported_type",
		},
		{
			name: "valid filter",
			f: resource.Filter{
				"aws_iam_role":       {},
				"aws_security_group": {},
				resource.Instance:    {},
				"aws_vpc":            {},
			},
		},
		{
			name: "valid filter includes awsls supported resources",
			f: resource.Filter{
				"aws_iam_role":       {},
				"aws_security_group": {},
				resource.Instance:    {},
				"aws_glue_job":       {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestFilter_Types(t *testing.T) {
	tests := []struct {
		name string
		f    resource.Filter
		want []string
	}{
		{
			name: "dependency order",
			f: resource.Filter{
				"aws_vpc":         {},
				resource.Instance: {},
			},
			want: []string{resource.Instance, "aws_vpc"},
		},
		{
			name: "dependency order not specified",
			f: resource.Filter{
				"aws_vpc":      {},
				"aws_glue_job": {},
			},
			want: []string{"aws_vpc", "aws_glue_job"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Types(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Types() = %v, want %v", got, tt.want)
			}
		})
	}
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

func TestTypeFilter_MatchTagged(t *testing.T) {
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
			name: "filter tagged resources, resource has tags",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(true),
			},
			tags: map[string]string{"foo": "bar"},
			want: true,
		},
		{
			name: "filter tagged resources, resource has no tags",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(true),
			},
			want: false,
		},
		{
			name: "filter untagged resources, resource has tags",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(false),
			},
			tags: map[string]string{"foo": "bar"},
			want: false,
		},
		{
			name: "filter untagged resources, resource has no tags",
			filter: resource.TypeFilter{
				Tagged: aws.Bool(false),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.MatchTagged(tt.tags); got != tt.want {
				t.Errorf("MatchTagged() = %v, want %v", got, tt.want)
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
			name:   "no tags filter, resources has no tags",
			filter: resource.TypeFilter{},
			want:   true,
		},
		{
			name:   "no tags filter, resources has tags",
			filter: resource.TypeFilter{},
			tags:   map[string]string{"foo": "bar"},
			want:   true,
		},
		{
			name: "filter one tag, resource has no tags",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^ba"},
				},
			},
			want: false,
		},
		{
			name: "filter one tag, resource tags have no matching key",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foz": "bar"},
			want: false,
		},
		{
			name: "filter one tag, one resource tag's key matches, but not value",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar"},
				},
			},
			tags: map[string]string{"foo": "baz"},
			want: false,
		},
		{
			name: "filter one tag, resource tag's key and value match",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar"},
			want: true,
		},
		{
			name: "filter one tag, one out of multiple resource tag's key and value match",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boo": "baz"},
			want: true,
		},
		{
			name: "filter multiple tags, all match",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^ba"},
					"boo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boo": "baz"},
			want: true,
		},
		{
			name: "filter multiple tags, one doesn't match (key)",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^ba"},
					"boo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boz": "baz"},
			want: false,
		},
		{
			name: "filter multiple tags, one doesn't match (value)",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^ba"},
					"boo": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boo": "boz"},
			want: false,
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

func TestTypeFilter_MatchNoTags(t *testing.T) {
	tests := []struct {
		name   string
		filter resource.TypeFilter
		tags   map[string]string
		want   bool
	}{
		{
			name:   "no notags filter, resource has no tags",
			filter: resource.TypeFilter{},
			want:   true,
		},
		{
			name:   "no notags filter, resource has tags",
			filter: resource.TypeFilter{},
			tags:   map[string]string{"foo": "bar"},
			want:   true,
		},
		{
			name: "resource has no tags",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"NOT(foo)": {Pattern: "^ba"},
				},
			},
			want: true,
		},
		{
			name: "no matching key",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"NOT(foo)": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foz": "bar"},
			want: true,
		},
		{
			name: "matching key, but not value",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"NOT(foo)": {Pattern: "^bar"},
				},
			},
			tags: map[string]string{"foo": "baz"},
			want: true,
		},
		{
			name: "matching key and value",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"NOT(foo)": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boo": "baz"},
			want: false,
		},
		{
			name: "matching key and value, multiple tags",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"NOT(foo)": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boo": "baz"},
			want: false,
		},
		{
			name: "multiple filter match",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"NOT(foo)": {Pattern: "^ba"},
					"NOT(boo)": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boo": "baz"},
			want: false,
		},
		{
			name: "one of multiple filter rules doesn't match key",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"NOT(foo)": {Pattern: "^ba"},
					"NOT(boo)": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boz": "baz"},
			want: true,
		},
		{
			name: "one of multiple filter rules doesn't match value",
			filter: resource.TypeFilter{
				Tags: map[string]resource.StringFilter{
					"NOT(foo)": {Pattern: "^ba"},
					"NOT(boo)": {Pattern: "^ba"},
				},
			},
			tags: map[string]string{"foo": "bar", "boo": "boz"},
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
