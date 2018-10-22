package resource

import (
	"regexp"

	"log"

	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// AppFs is an abstraction of the file system
// to allow mocking in tests.
var AppFs = afero.NewOsFs()

// YamlCfg represents the data structure of a yaml
// file that is used as a contract to select resources.
// Each yamlEntry selects the resources of a particular resource type.
type YamlCfg map[TerraformResourceType]yamlEntry

// yamlEntry represents an entry in YamlCfg
// i.e., regexps to select
// a subset of resources by ids or findTags.
type yamlEntry struct {
	ID   string            `yaml:",omitempty"`
	Tags map[string]string `yaml:",omitempty"`
}

// YamlFilter selects resources
// stated in a yaml configuration for deletion.
type YamlFilter struct {
	file string
	Cfg  YamlCfg
}

// NewFilter creates a new filter based on a config given via a yaml file.
func NewFilter(yamlFile string) *YamlFilter {
	return &YamlFilter{
		Cfg: read(yamlFile),
	}
}

// read reads a filter from a yaml file.
func read(filename string) YamlCfg {
	var cfg YamlCfg

	data, err := afero.ReadFile(AppFs, filename)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

// Validate checks if all resource types appearing in the config are currently supported.
func (f YamlFilter) Validate() error {
	for _, resType := range f.Types() {
		if !SupportedResourceType(resType) {
			return fmt.Errorf("unsupported resource type found in yaml config: %s", resType)
		}
	}
	return nil
}

// Types returns all the resource types in the config.
func (f YamlFilter) Types() []TerraformResourceType {
	resTypes := make([]TerraformResourceType, 0, len(f.Cfg))

	for k := range f.Cfg {
		resTypes = append(resTypes, k)
	}

	return resTypes
}

// MatchID checks whether a resource (given by its type and id) matches the filter.
func (f YamlFilter) matchID(resType TerraformResourceType, id string) bool {
	cfgEntry, found := f.Cfg[resType]
	if !found {
		return false
	}

	if cfgEntry.ID == "" {
		return true
	}

	if ok, err := regexp.MatchString(cfgEntry.ID, id); ok {
		if err != nil {
			log.Fatal(err)
		}
		return true
	}

	return false
}

// MatchesTags checks whether a resource (given by its type and findTags)
// matches the filter. The keys must match exactly, whereas the tag value is checked against a regex.
func (f YamlFilter) matchTags(resType TerraformResourceType, tags map[string]string) bool {
	cfgEntry, found := f.Cfg[resType]
	if !found {
		return false
	}

	if len(cfgEntry.Tags) == 0 {
		return true
	}

	for cfgTagKey, regex := range cfgEntry.Tags {
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

// matches checks whether a resource matches the filter criteria.
func (f YamlFilter) matches(r *DeletableResource) bool {
	return f.matchTags(r.Type, r.Tags) && f.matchID(r.Type, r.ID)
}
