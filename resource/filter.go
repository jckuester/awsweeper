package resource

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	"gopkg.in/yaml.v2"
)

// Config represents the content of a yaml file that is used as a contract to filter resources for deletion.
type Config map[TerraformResourceType][]TypeFilter

// TypeFilter represents an entry in Config and selects the resources of a particular resource type.
type TypeFilter struct {
	ID   *StringFilter            `yaml:",omitempty"`
	Tags map[string]*StringFilter `yaml:",omitempty"`
	// select resources by creation time
	Created *Created `yaml:",omitempty"`
}

type StringMatcher interface {
	matches(string) (bool, error)
}

type StringFilter struct {
	Pattern string `yaml:",omitempty"`
	Negate  bool
}

type CreatedTime struct {
	time.Time `yaml:",omitempty"`
}

type Created struct {
	Before *CreatedTime `yaml:",omitempty"`
	After  *CreatedTime `yaml:",omitempty"`
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

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithError(err).Fatalf("failed to read config file: %s", filename)
	}

	err = yaml.UnmarshalStrict(data, &cfg)
	if err != nil {
		log.WithError(err).Fatalf("failed to unmarshal config: %s", filename)
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

	if ok, err := rtf.ID.matches(id); ok {
		if err != nil {
			log.WithError(err).Fatal("failed to match ID")
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
			if matched, err := regex.matches(tagVal); !matched {
				if err != nil {
					log.WithError(err).Fatal("failed to match tags")
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

func (f *StringFilter) matches(s string) (bool, error) {
	ok, err := regexp.MatchString(f.Pattern, s)
	if err != nil {
		return false, err
	}

	if f.Negate {
		return !ok, nil
	}

	return ok, err
}

func (f *StringFilter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}
	if strings.HasPrefix(v, "NOT(") && strings.HasSuffix(v, ")") {
		*f = StringFilter{strings.TrimSuffix(strings.TrimPrefix(v, "NOT("), ")"), true}
	} else {
		*f = StringFilter{v, false}
	}
	return nil
}

func (c *CreatedTime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}
	switch value := v.(type) {
	case time.Time:
		t, _ := v.(time.Time)
		*c = CreatedTime{t}
		return nil
	case string:
		d, err := time.ParseDuration(value)
		if err == nil {
			*c = CreatedTime{time.Now().UTC().Add(-d)}
			return nil
		}
		var t time.Time
		err = yaml.Unmarshal([]byte("!!timestamp "+value), &t)
		if err == nil {
			*c = CreatedTime{t}
			return nil
		}
		if strings.HasSuffix(value, "d") {
			d, err := strconv.ParseInt(value[0:len(value)-1], 10, 32)
			if err == nil {
				*c = CreatedTime{time.Now().UTC().AddDate(0, 0, -int(d))}
				return nil
			}
		}
		if strings.HasSuffix(value, "w") {
			w, err := strconv.ParseInt(value[0:len(value)-1], 10, 32)
			if err == nil {
				*c = CreatedTime{time.Now().UTC().AddDate(0, 0, -int(w*7))}
				return nil
			}
		}
		if strings.HasSuffix(value, "M") {
			m, err := strconv.ParseInt(value[0:len(value)-1], 10, 32)
			if err == nil {
				*c = CreatedTime{time.Now().UTC().AddDate(0, -int(m), 0)}
				return nil
			}
		}
		if strings.HasSuffix(value, "y") {
			y, err := strconv.ParseInt(value[0:len(value)-1], 10, 32)
			if err == nil {
				*c = CreatedTime{time.Now().UTC().AddDate(-int(y), 0, 0)}
				return nil
			}
		}
	}
	return errors.New("invalid created time")
}
