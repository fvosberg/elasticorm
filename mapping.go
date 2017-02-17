package elasticorm

import "reflect"

type MappingConfig struct {
	Type string `json:"type"`
}

func MappingFromStruct(i interface{}) map[string]map[string]MappingConfig {
	mapping := make(map[string]map[string]MappingConfig)
	mapping[`properties`] = make(map[string]MappingConfig)

	v := reflect.ValueOf(i).Elem()
	for n := 0; n < v.NumField(); n++ {
		mapping[`properties`][v.Type().Field(n).Name] = MappingConfig{
			Type: `text`,
		}
	}

	return mapping
}
