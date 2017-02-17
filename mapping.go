package elasticorm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type MappingConfig struct {
	Type string `json:"type"`
}

func MappingFromStruct(i interface{}) (map[string]map[string]MappingConfig, error) {
	mapping := make(map[string]map[string]MappingConfig)
	mapping[`properties`] = make(map[string]MappingConfig)
	var err error

	v := reflect.ValueOf(i).Elem()
	for n := 0; n < v.NumField(); n++ {
		name := v.Type().Field(n).Name
		if jn := v.Type().Field(n).Tag.Get(`json`); jn != `` {
			name = jn
		}

		mappingType := `text`
		if et := v.Type().Field(n).Tag.Get(`elasticorm`); et != `` {
			opts := strings.Split(et, `=`)
			switch opts[0] {
			case `type`:
				mappingType = opts[1]
			default:
				err = errors.Wrap(InvalidOptionErr, fmt.Sprintf("parsing option %s failed", et))
			}
		}

		mapping[`properties`][name] = MappingConfig{
			Type: mappingType,
		}
	}

	return mapping, err
}
