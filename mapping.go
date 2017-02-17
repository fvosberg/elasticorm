package elasticorm

import (
	"reflect"
	"strings"
)

type MappingConfig struct {
	Type string `json:"type"`
}

func MappingFromStruct(i interface{}) map[string]map[string]MappingConfig {
	mapping := make(map[string]map[string]MappingConfig)
	mapping[`properties`] = make(map[string]MappingConfig)

	v := reflect.ValueOf(i).Elem()
	for n := 0; n < v.NumField(); n++ {
		name := v.Type().Field(n).Name
		if jn := v.Type().Field(n).Tag.Get(`json`); jn != `` {
			name = jn
		}

		mappingType := `text`
		if et := v.Type().Field(n).Tag.Get(`elasticorm`); et != `` {
			opts := strings.Split(et, `=`)
			if opts[0] == `type` {
				mappingType = opts[1]
			}
		}

		mapping[`properties`][name] = MappingConfig{
			Type: mappingType,
		}
	}

	return mapping
}
