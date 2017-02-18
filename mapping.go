package elasticorm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type MappingConfig struct {
	Properties map[string]MappingFieldConfig `json:"properties,omitempty"`
}

func (m *MappingConfig) AddField(name string, cfg MappingFieldConfig) {
	if m.Properties == nil {
		m.Properties = make(map[string]MappingFieldConfig)
	}
	m.Properties[name] = cfg
}

type MappingFieldConfig struct {
	Type     string `json:"type"`
	Analyzer string `json:"analyzer,omitempty"`
}

func MappingFromStruct(i interface{}) (MappingConfig, error) {
	mapping := MappingConfig{}
	var err error

	v := reflect.ValueOf(i).Elem()
	for n := 0; n < v.NumField(); n++ {
		fieldMapping, propErr := mappingForField(v.Type().Field(n).Tag.Get(`elasticorm`))
		if propErr != nil {
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
		opts := strings.Split(tag, `=`)
		switch opts[0] {
		case `type`:
			propMapping.Type = opts[1]
		case `analyzer`:
			propMapping.Analyzer = opts[1]
		default:
			err = errors.Wrap(InvalidOptionErr, fmt.Sprintf("parsing option %s failed", tag))
		}
	}
	return propMapping, err
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
