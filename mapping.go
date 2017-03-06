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

func (m MappingConfig) elasticFieldName(structFieldName string) (string, error) {
	for propertyName, propertyMapping := range m.Properties {
		if propertyMapping.structFieldName == structFieldName {
			return propertyName, nil
		}
	}
	return ``, errors.New(`Mapping configuration has no mapping for struct field`)
}

// MappingFieldConfig is a struct which represents the elasticsearch mapping configuration of one field. It is used in the MappingConfig.
type MappingFieldConfig struct {
	Type            string                        `json:"type"`
	Analyzer        string                        `json:"analyzer,omitempty"`
	structFieldName string                        `json:"-"`
	Properties      map[string]MappingFieldConfig `json:"properties,omitempty"`
	Fields          map[string]MappingFieldConfig `json:"fields,omitempty"`
}

// MappingFromStruct returns the MappingConfig for a passed in struct (pointer). The mapping is configurable via json tags, which can change the name of the field, and elasticorm tags. The elasticorm tags can include
func MappingFromStruct(i interface{}) (MappingConfig, error) {
	mapping := MappingConfig{}
	var err error
	v := reflect.ValueOf(i).Elem()
	for n := 0; n < v.NumField(); n++ {
		field := v.Type().Field(n)
		if !shouldMapField(field) {
			continue
		}
		fieldMapping, propErr := mappingForField(field)
		if propErr != nil {
			err = propErr
		}
		name := nameForField(field)
		mapping.AddField(name, fieldMapping)
	}
	return mapping, err
}

func mappingForField(field reflect.StructField) (MappingFieldConfig, error) {
	var err error
	propMapping := MappingFieldConfig{
		Type:            typeForField(field),
		structFieldName: field.Name,
	}
	propMapping.Properties, err = propertiesForField(field)
	if err != nil {
		return propMapping, err
	}
	if tag := field.Tag.Get(`elasticorm`); tag != `` {
		options := optionsFromTag(tag)
		for name, value := range options {
			switch name {
			case `type`:
				propMapping.Type = value
			case `analyzer`:
				propMapping.Analyzer = value
			case `sortable`:
				propMapping.Fields = rawFieldForField(field)
			case `id`:
			default:
				err = errors.Wrap(ErrInvalidOption, fmt.Sprintf("parsing option %s=%s failed", name, value))
			}
		}
	}
	return propMapping, err
}

func typeForField(f reflect.StructField) string {
	if val, ok := optionValueForField(f, `type`); ok {
		return val
	}
	return elasticTypeForGoType(f.Type)
}

func elasticTypeForGoType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.PkgPath() == `time` && t.Name() == `Time` {
		return `date`
	}
	switch t.Kind() {
	case reflect.Slice:
		subtype := elasticTypeForGoType(t.Elem())
		if subtype == `object` {
			return `nested`
		}
		return subtype
	case reflect.Struct:
		return `object`
	case reflect.Bool:
		return `boolean`
	default:
		return `text`
	}
}

func optionValueForField(f reflect.StructField, name string) (string, bool) {
	o := optionsForField(f)
	v, ok := o[name]
	return v, ok
}

func optionsForField(f reflect.StructField) map[string]string {
	o := make(map[string]string, 2)
	tag := f.Tag.Get(`elasticorm`)
	if tag == `` {
		return o
	}
	return optionsFromTag(tag)
}

func propertiesForField(f reflect.StructField) (map[string]MappingFieldConfig, error) {
	t := f.Type
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	if typeForField(f) != `nested` && typeForField(f) != `object` {
		return nil, nil
	}
	properties := make(map[string]MappingFieldConfig, t.NumField())
	var err error
	for n := 0; n < t.NumField(); n++ {
		field := t.Field(n)
		if shouldMapField(field) {
			properties[nameForField(field)], err = mappingForField(field)
		}
	}
	return properties, err
}

func rawFieldForField(f reflect.StructField) map[string]MappingFieldConfig {
	cfg := make(map[string]MappingFieldConfig, 1)
	cfg[`raw`] = MappingFieldConfig{
		Type: `keyword`,
	}
	return cfg
}

func shouldMapField(f reflect.StructField) bool {
	_, isId := optionValueForField(f, `id`)
	return !(isId || f.Tag.Get(`json`) == `-`)
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
