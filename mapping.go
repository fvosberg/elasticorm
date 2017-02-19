package elasticorm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// MappingConfig is a struct which marshals to a valid elasticsearch mapping configuration
type MappingConfig struct {
	Properties map[string]MappingFieldConfig `json:"properties,omitempty"`
}

// AddField adds a new field to the mapping
func (m *MappingConfig) AddField(name string, cfg MappingFieldConfig) {
	if m.Properties == nil {
		m.Properties = make(map[string]MappingFieldConfig)
	}
	m.Properties[name] = cfg
}

// MappingFieldConfig is a struct which represents the elasticsearch mapping configuration of one field. It is used in the MappingConfig.
type MappingFieldConfig struct {
	Type     string `json:"type"`
	Analyzer string `json:"analyzer,omitempty"`
}

// MappingFromStruct returns the MappingConfig for a passed in struct (pointer). The mapping is configurable via json tags, which can change the name of the field, and elasticorm tags. The elasticorm tags can include
func MappingFromStruct(i interface{}) (MappingConfig, error) {
	mapping := MappingConfig{}
	var err error

	v := reflect.ValueOf(i).Elem()
	for n := 0; n < v.NumField(); n++ {
		fieldMapping, propErr := mappingForField(v.Type().Field(n).Tag.Get(`elasticorm`))
		if propErr == errIdField {
			continue
		} else if propErr != nil {
			err = propErr
		}
		name := nameForField(v.Type().Field(n))
		mapping.AddField(name, fieldMapping)
	}

	return mapping, err
}

func mappingForField(tag string) (MappingFieldConfig, error) {
	var err error
	propMapping := MappingFieldConfig{
		Type: `text`,
	}

	if tag != `` {
		options := optionsFromTag(tag)
		for name, value := range options {
			switch name {
			case `type`:
				propMapping.Type = value
			case `analyzer`:
				propMapping.Analyzer = value
			case `id`:
				return MappingFieldConfig{}, errIdField
			default:
				err = errors.Wrap(ErrInvalidOption, fmt.Sprintf("parsing option %s=%s failed", name, value))
			}
		}
	}
	return propMapping, err
}

func optionsFromTag(tag string) map[string]string {
	options := make(map[string]string, 2)
	definitions := strings.Split(tag, `,`)
	for _, definition := range definitions {
		kv := strings.Split(definition, `=`)
		if len(kv) > 1 {
			options[kv[0]] = kv[1]
		} else {
			options[kv[0]] = ``
		}
	}
	return options
}

func nameForField(field reflect.StructField) string {
	name := field.Name
	if json := field.Tag.Get(`json`); json != `` {
		if i := strings.Index(json, `,`); i > -1 {
			json = json[:strings.Index(json, `,`)]
		}
		name = json
	}
	return name
}
