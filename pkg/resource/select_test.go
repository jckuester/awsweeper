package resource_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/jckuester/awstools-lib/terraform"
	"github.com/jckuester/awsweeper/pkg/resource"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestYamlFilter_Apply_EmptyConfig(t *testing.T) {
	//given
	f := &resource.Filter{}

	res := []terraform.Resource{
		{
			Type: "aws_instance",
			ID:   "foo",
		},
	}

	// when
	result := f.Apply(res)

	// then
	assert.Len(t, result, 0)
}

func TestYamlFilter_Apply_FilterAll(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {},
	}
	res := []terraform.Resource{
		{
			Type: "aws_instance",
			ID:   "foo",
		},
	}

	// when
	result := f.Apply(res)

	// then
	assert.Len(t, result, len(res))
	assert.Equal(t, res[0].ID, result[0].ID)
}

func TestYamlFilter_Apply_FilterByID(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
			{
				ID: &resource.StringFilter{Pattern: "^select"},
			},
		},
	}

	// when
	res := []terraform.Resource{
		{
			Type: "aws_instance",
			ID:   "select-this",
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this",
		},
	}

	result := f.Apply(res)

	// then
	require.Len(t, result, 1)
	assert.Equal(t, "select-this", result[0].ID)
}

func TestYamlFilter_Apply_FilterByTag(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
			{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar"},
				},
			},
		},
	}

	res := []terraform.Resource{
		{
			Type: "aws_instance",
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this",
			Tags: map[string]string{
				"foo": "blub",
			},
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this-either",
		},
	}

	// when
	result := f.Apply(res)

	// then
	require.Len(t, result, 1)
	assert.Equal(t, "select-this", result[0].ID)
}

func TestYamlFilter_Apply_FilterByMultipleTags(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
			{
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar"},
					"bla": {Pattern: "^blub"},
				},
			},
		},
	}

	res := []terraform.Resource{
		{
			Type: "aws_instance",
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
				"bla": "blub",
			},
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
	}

	// when
	result := f.Apply(res)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "select-this", result[0].ID)
}

func TestYamlFilter_Apply_FilterByIDandTag(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
			{
				ID: &resource.StringFilter{Pattern: "^foo"},
				Tags: map[string]resource.StringFilter{
					"foo": {Pattern: "^bar"},
				},
			},
		},
	}

	res := []terraform.Resource{
		{
			Type: "aws_instance",
			ID:   "foo",
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this",
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		{
			Type: "aws_instance",
			ID:   "this-neither",
		},
	}

	// when
	result := f.Apply(res)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "foo", result[0].ID)
}

func TestYamlFilter_Apply_Created(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
			{
				Created: &resource.Created{
					After:  &resource.CreatedTime{Time: time.Date(2018, 11, 17, 0, 0, 0, 0, time.UTC)},
					Before: &resource.CreatedTime{Time: time.Date(2018, 11, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	res := []terraform.Resource{
		{
			Type:      "aws_instance",
			ID:        "foo",
			CreatedAt: aws.Time(time.Date(2018, 11, 17, 5, 0, 0, 0, time.UTC)),
		},
		{
			Type:      "aws_instance",
			ID:        "do-not-select-this1",
			CreatedAt: aws.Time(time.Date(2018, 11, 17, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type:      "aws_instance",
			ID:        "do-not-select-this2",
			CreatedAt: aws.Time(time.Date(2018, 11, 20, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type:      "aws_instance",
			ID:        "do-not-select-this3",
			CreatedAt: aws.Time(time.Date(2018, 11, 22, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this2",
		},
	}

	// when
	result := f.Apply(res)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "foo", result[0].ID)
}

func TestYamlFilter_Apply_CreatedBefore(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
			{
				Created: &resource.Created{
					Before: &resource.CreatedTime{Time: time.Date(2018, 11, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	res := []terraform.Resource{
		{
			Type:      "aws_instance",
			ID:        "foo",
			CreatedAt: aws.Time(time.Date(2018, 11, 17, 5, 0, 0, 0, time.UTC)),
		},
		{
			Type:      "aws_instance",
			ID:        "do-not-select-this",
			CreatedAt: aws.Time(time.Date(2018, 11, 22, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this2",
		},
	}

	// when
	result := f.Apply(res)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "foo", result[0].ID)
}

func TestYamlFilter_Apply_CreatedAfter(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
			{
				Created: &resource.Created{
					After: &resource.CreatedTime{Time: time.Date(2018, 11, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	res := []terraform.Resource{
		{
			Type:      "aws_instance",
			ID:        "foo",
			CreatedAt: aws.Time(time.Date(2018, 11, 22, 5, 0, 0, 0, time.UTC)),
		},
		{
			Type:      "aws_instance",
			ID:        "do-not-select-this",
			CreatedAt: aws.Time(time.Date(2018, 11, 17, 0, 0, 0, 0, time.UTC)),
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this2",
		},
	}

	// when
	result := f.Apply(res)

	// then
	assert.Len(t, result, 1)
	assert.Equal(t, "foo", result[0].ID)
}

func TestYamlFilter_Apply_MultipleFiltersPerResourceType(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
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

	res := []terraform.Resource{
		{
			Type: "aws_instance",
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
		{
			Type: "aws_instance",
			ID:   "select-this-too",
			Tags: map[string]string{
				"bla": "blub",
			},
		},
		{
			Type: "aws_instance",
			ID:   "do-not-select-this",
			Tags: map[string]string{
				"bla": "blub",
			},
		},
	}

	// when
	result := f.Apply(res)

	// then
	require.Len(t, result, 2)
	assert.Equal(t, "select-this", result[0].ID)
	assert.Equal(t, "select-this-too", result[1].ID)
}

func TestYamlFilter_Apply_NegatedStringFilter(t *testing.T) {
	//given
	f := &resource.Filter{
		"aws_instance": {
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

	res := []terraform.Resource{
		{
			Type: "aws_instance",
			ID:   "select-this-not",
			Tags: map[string]string{
				"foo": "bar-bab",
			},
		},
		{
			Type: "aws_instance",
			ID:   "select-this",
			Tags: map[string]string{
				"foo": "baz",
			},
		},
	}

	// when
	result := f.Apply(res)

	// then
	require.Len(t, result, 1)
	assert.Equal(t, "select-this", result[0].ID)
}

func TestGetTags(t *testing.T) {
	tests := []struct {
		name    string
		arg     *terraform.Resource
		want    map[string]string
		wantErr string
	}{
		{
			name:    "resource is nil",
			wantErr: "resource is nil: <nil>",
		},
		{
			name:    "embedded updatable resource is nil",
			arg:     &terraform.Resource{},
			wantErr: "resource is nil: &{Type: ID: Region: Profile: AccountID: Tags:map[] CreatedAt:<nil> UpdatableResource:<nil>}",
		},
		{
			name: "state is nil",
			arg: &terraform.Resource{
				UpdatableResource: &terradozerRes.Resource{},
			},
			wantErr: "state is nil: <nil>",
		},
		{
			name: "state is nil value",
			arg: &terraform.Resource{
				UpdatableResource: terradozerRes.NewWithState("aws_foo", "1234", nil, &cty.NilVal),
			},
			wantErr: "state is nil: &{ty:{typeImpl:<nil>} v:<nil>}",
		},
		{
			name: "null map",
			arg: &terraform.Resource{
				UpdatableResource: terradozerRes.NewWithState("aws_foo", "1234",
					nil, ctyValuePtr(cty.NullVal(cty.Map(cty.String)))),
			},
			wantErr: "state is nil: &{ty:{typeImpl:{typeImplSigil:{} ElementTypeT:{typeImpl:{typeImplSigil:{} Kind:83}}}} v:<nil>}",
		},
		{
			name: "unhandled type",
			arg: &terraform.Resource{
				UpdatableResource: terradozerRes.NewWithState("aws_foo", "1234",
					nil, ctyValuePtr(cty.ObjectVal(map[string]cty.Value{
						"tags": cty.StringVal("foo"),
					}))),
			},
			wantErr: "currently unhandled type: string",
		},
		{
			name: "tags attribute not found",
			arg: &terraform.Resource{
				UpdatableResource: terradozerRes.NewWithState("aws_foo", "1234",
					nil, ctyValuePtr(cty.ObjectVal(map[string]cty.Value{
						"tag": cty.StringVal("foo"),
					}))),
			},
			wantErr: "attribute not found: tags",
		},
		{
			name: "cannot iterate element",
			arg: &terraform.Resource{
				UpdatableResource: terradozerRes.NewWithState("aws_foo", "1234",
					nil, ctyValuePtr(cty.StringVal("foo"))),
			},
			wantErr: "cannot iterate: cty.StringVal(\"foo\")",
		},
		{
			name: "empty map of tags",
			arg: &terraform.Resource{
				UpdatableResource: terradozerRes.NewWithState("aws_foo", "1234",
					nil, ctyValuePtr(cty.ObjectVal(map[string]cty.Value{
						"tags": cty.MapValEmpty(cty.String),
					}))),
			},
			want: map[string]string{},
		},
		{
			name: "some tags",
			arg: &terraform.Resource{
				UpdatableResource: terradozerRes.NewWithState("aws_foo", "1234",
					nil, ctyValuePtr(cty.ObjectVal(map[string]cty.Value{
						"tags": cty.MapVal(map[string]cty.Value{"foo": cty.StringVal("bar")}),
					}))),
			},
			want: map[string]string{"foo": "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resource.GetTags(tt.arg)

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetTags() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func ctyValuePtr(v cty.Value) *cty.Value {
	return &v
}
