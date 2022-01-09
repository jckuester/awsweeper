package resource

import (
	"fmt"

	"github.com/apex/log"
	"github.com/jckuester/awstools-lib/terraform"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Apply applies the filter to the given resources.
func (f Filter) Apply(res []terraform.Resource) []terraform.Resource {
	for i, r := range res {
		tags, err := GetTags(&r)
		if err != nil {
			log.WithFields(log.Fields{
				"type": r.Type,
				"id":   r.ID,
			}).WithError(err).Debug("failed to get tags")

			continue
		}

		res[i].Tags = tags
	}

	var result []terraform.Resource

	for _, r := range res {
		if f.Match(r) {
			result = append(result, r)
		}
	}

	return result
}

func GetTags(r *terraform.Resource) (map[string]string, error) {
	if r == nil || r.UpdatableResource == nil {
		return nil, fmt.Errorf("resource is nil: %+v", r)
	}

	state := r.State()

	if state == nil || state.IsNull() {
		return nil, fmt.Errorf("state is nil: %+v", state)
	}

	if !state.CanIterateElements() {
		return nil, fmt.Errorf("cannot iterate: %s", state.GoString())
	}

	attrValue, ok := state.AsValueMap()["tags"]
	if !ok {
		return nil, fmt.Errorf("attribute not found: tags")
	}

	if attrValue.IsNull() {
		return nil, fmt.Errorf("attribute is null value")
	}

	switch attrValue.Type() {
	case cty.Map(cty.String):
		var v map[string]string
		err := gocty.FromCtyValue(attrValue, &v)
		if err != nil {
			return nil, err
		}

		return v, nil
	default:
		return nil, fmt.Errorf("currently unhandled type: %s", attrValue.Type().FriendlyName())
	}
}
