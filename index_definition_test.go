package elasticorm_test

import (
	"encoding/json"
	"testing"

	"github.com/fvosberg/elasticorm"
)

func TestIndexDefinition(t *testing.T) {
	def := elasticorm.NewIndexDefinition(`Customers`)

	actualJSON, err := json.Marshal(def)
	ok(t, err)
	equals(t, `{}`, string(actualJSON))
}
