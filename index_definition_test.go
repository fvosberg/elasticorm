package elasticorm_test

import (
	"encoding/json"
	"testing"

	"github.com/fvosberg/elasticorm"
)

func TestIndexDefinition(t *testing.T) {
	tests := []struct {
		title        string
		def          elasticorm.IndexDefinition
		expectedJSON string
	}{
		{
			title:        `Empty Index definition`,
			def:          elasticorm.NewIndexDefinition(),
			expectedJSON: `{}`,
		},
		{
			title: `Index definition with number of shards setting`,
			def: elasticorm.NewIndexDefinition(
				elasticorm.SetNumberOfShards(3),
			),
			expectedJSON: `{"settings":{"number_of_shards":3}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			actualJSON, err := json.Marshal(tt.def)
			ok(t, err)
			equals(t, tt.expectedJSON, string(actualJSON))
		})
	}
}
