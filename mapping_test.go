package elasticorm_test

import (
	"testing"

	"encoding/json"

	"github.com/fvosberg/elasticorm"
)

func TestNewMappingFromStruct(t *testing.T) {
	type User struct {
		FirstName string
		LastName  string
	}

	mapping := elasticorm.MappingFromStruct(&User{})

	expectedJSON := `{"properties":{"FirstName":{"type":"text"},"LastName":{"type":"text"}}}`
	actualJSON, err := json.Marshal(mapping)
	ok(t, err)
	equals(t, expectedJSON, string(actualJSON))
}
