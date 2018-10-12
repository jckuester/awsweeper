package resource

import (
	"regexp"

	"log"

	"errors"

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
	Ids  []*string         `yaml:",omitempty"`
	Tags map[string]string `yaml:",omitempty"`
}

// Filter selects resources for deletion.
type Filter interface {
	Apply(resType TerraformResourceType, res Resources, raw interface{}, aws *AWS) []Resources
	//Validate(as []APIDesc) error
	Matches(resType TerraformResourceType, id string, tags ...map[string]string) bool
	Types() []string
}

// YamlFilter selects resources
// stated in a yaml configuration for deletion.
type YamlFilter struct {
	file string
	cfg  YamlCfg
}

// NewFilter creates a new filter to select resources for deletion
// based on the path of yaml file.
func NewFilter(yamlFile string) *YamlFilter {
	return &YamlFilter{
		file: yamlFile,
		cfg:  read(yamlFile),
	}
}

// read reads a filter from a yaml file.
func read(file string) YamlCfg {
	var cfg YamlCfg

	data, err := afero.ReadFile(AppFs, file)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

// Validate checks if all resource types appearing in the config
// of the filter are currently supported.
//func (f YamlFilter) Validate(as []APIDesc) error {
//	for _, resType := range f.Types() {
//		isTerraformType := false
//		for _, a := range as {
//			if resType == a.Type {
//				isTerraformType = true
//			}
//		}
//		if !isTerraformType {
//			return fmt.Errorf("unsupported resource type '%s' found in '%s'", resType, f.file)
//		}
//	}
//	return nil
//}

// Types returns all the resource types stated in the yaml config.
// We use the same identifiers of resource types as the Terraform AWS provider.
func (f YamlFilter) Types() []TerraformResourceType {
	resTypes := make([]TerraformResourceType, 0, len(f.cfg))

	for k := range f.cfg {
		resTypes = append(resTypes, k)
	}

	return resTypes
}

// MatchID checks whether a resource (given by its type and id)
// matches the filter.
func (f YamlFilter) matchID(resType TerraformResourceType, id string) (bool, error) {
	cfgEntry, _ := f.cfg[resType]

	if len(cfgEntry.Ids) == 0 {
		return false, errors.New("no entries set in filter to match IDs")
	}

	for _, regex := range cfgEntry.Ids {
		if ok, err := regexp.MatchString(*regex, id); ok {
			if err != nil {
				log.Fatal(err)
			}
			return true, nil
		}
	}

	return false, nil
}

// MatchesTags checks whether a resource (given by its type and findTags)
// matches the filter. The keys must match exactly, whereas
// the tag value is checked against a regex.
func (f YamlFilter) matchTags(resType TerraformResourceType, tags map[string]string) (bool, error) {
	cfgEntry, _ := f.cfg[resType]

	if len(cfgEntry.Tags) == 0 {
		return false, errors.New("No entries set in filter to match findTags")
	}

	for cfgTagKey, regex := range cfgEntry.Tags {
		if tagVal, ok := tags[cfgTagKey]; ok {
			if res, err := regexp.MatchString(regex, tagVal); res {
				if err != nil {
					log.Fatal(err)
				}
				return true, nil
			}
		}
	}

	return false, nil
}

// Matches checks whether a resource (given by its type and findTags) matches
// the configured filter criteria for findTags and ids.
func (f YamlFilter) Matches(resType TerraformResourceType, id string, tags ...map[string]string) bool {
	var matchesTags = false
	var errTags error

	if tags != nil {
		matchesTags, errTags = f.matchTags(resType, tags[0])
	}
	matchesID, errID := f.matchID(resType, id)

	// if the filter has neither an entry to match ids nor findTags,
	// select all resources of that type
	if errID != nil && errTags != nil {
		return true
	}

	return matchesID || matchesTags
}
