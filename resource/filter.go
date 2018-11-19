package resource

import (
	"regexp"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"log"

	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// AppFs is an abstraction of the file system to allow mocking in tests.
var AppFs = afero.NewOsFs()

// Config represents the content of a yaml file that is used as a contract to filter resources for deletion.
type Config map[TerraformResourceType][]TypeFilter

// TypeFilter represents an entry in Config and selects the resources of a particular resource type.
type TypeFilter struct {
	ID   *string           `yaml:",omitempty"`
	Tags map[string]string `yaml:",omitempty"`
	// select resources by creation time
	Created *Created `yaml:",omitempty"`
}

type Created struct {
	Before *time.Time `yaml:",omitempty"`
	After  *time.Time `yaml:",omitempty"`
}

// Filter selects resources based on a given yaml config.
type Filter struct {
	Cfg Config
}

// NewFilter creates a new filter based on a config given via a yaml file.
func NewFilter(yamlFile string) *Filter {
	return &Filter{
		Cfg: read(yamlFile),
	}
}

// read reads a filter from a yaml file.
func read(filename string) Config {
	var cfg Config

	data, err := afero.ReadFile(AppFs, filename)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to read config file: %s", filename)
	}

	err = yaml.UnmarshalStrict(data, &cfg)
	if err != nil {
		logrus.WithError(err).Fatalf("Cannot unmarshal config: %s", filename)
	}

	return cfg
}

// Validate checks if all resource types appearing in the config are currently supported.
func (f Filter) Validate() error {
	for _, resType := range f.Types() {
		if !SupportedResourceType(resType) {
			return fmt.Errorf("unsupported resource type found in yaml config: %s", resType)
		}
	}
	return nil
}

// Types returns all the resource types in the config in their dependency order.
func (f Filter) Types() []TerraformResourceType {
	resTypes := make([]TerraformResourceType, 0, len(f.Cfg))

	for k := range f.Cfg {
		resTypes = append(resTypes, k)
	}

	sort.Slice(resTypes, func(i, j int) bool {
		return DependencyOrder[resTypes[i]] > DependencyOrder[resTypes[j]]
	})

	return resTypes
}

// MatchID checks whether a resource ID matches the filter.
func (rtf TypeFilter) matchID(id string) bool {
	if rtf.ID == nil {
		return true
	}

	if ok, err := regexp.MatchString(*rtf.ID, id); ok {
		if err != nil {
			log.Fatal(err)
		}
		return true
	}

	return false
}

// MatchesTags checks whether a resource's tags
// match the filter. The keys must match exactly, whereas the tag value is checked against a regex.
func (rtf TypeFilter) matchTags(tags map[string]string) bool {
	if rtf.Tags == nil {
		return true
	}

	for cfgTagKey, regex := range rtf.Tags {
		if tagVal, ok := tags[cfgTagKey]; ok {
			if matched, err := regexp.MatchString(regex, tagVal); !matched {
				if err != nil {
					log.Fatal(err)
				}
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func (rtf TypeFilter) matchCreated(creationTime *time.Time) bool {
	if rtf.Created == nil {
		return true
	}

	if creationTime == nil {
		return false
	}

	createdAfter := true
	if rtf.Created.After != nil {
		createdAfter = creationTime.Unix() > rtf.Created.After.Unix()
	}

	createdBefore := true
	if rtf.Created.Before != nil {
		createdBefore = creationTime.Unix() < rtf.Created.Before.Unix()
	}

	return createdAfter && createdBefore
}

// matches checks whether a resource matches the filter criteria.
func (f Filter) matches(r *Resource) bool {
	resTypeFilters, found := f.Cfg[r.Type]
	if !found {
		return false
	}

	if len(resTypeFilters) == 0 {
		return true
	}

	for _, rtf := range resTypeFilters {
		if rtf.matchTags(r.Tags) && rtf.matchID(r.ID) && rtf.matchCreated(r.Created) {
			return true
		}
	}
	return false
}
