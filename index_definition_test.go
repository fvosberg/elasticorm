package elasticorm_test

import (
	"encoding/json"
	"testing"

	"github.com/fvosberg/elasticorm"
)

func TestIndexDefinition(t *testing.T) {
	tests := []struct {
		title        string
		defFuncs     []elasticorm.IndexDefinitionFunc
		expectedJSON string
	}{
		{
			title:        `Empty Index definition`,
			defFuncs:     []elasticorm.IndexDefinitionFunc{},
			expectedJSON: `{"settings":{}}`,
		},
		{
			title: `Index definition with number of shards setting`,
			defFuncs: []elasticorm.IndexDefinitionFunc{
				elasticorm.SetNumberOfShards(3),
			},
			expectedJSON: `{"settings":{"number_of_shards":3}}`,
		},
		{
			title: `Index definition with number of replicas setting`,
			defFuncs: []elasticorm.IndexDefinitionFunc{
				elasticorm.SetNumberOfReplicas(2),
			},
			expectedJSON: `{"settings":{"number_of_replicas":2}}`,
		},
		{
			title: `Index definition with a setting and a customer mapping`,
			defFuncs: []elasticorm.IndexDefinitionFunc{
				elasticorm.SetNumberOfReplicas(2),
				elasticorm.AddMappingFromStruct(
					`customer`,
					(func() interface{} {
						type User struct {
							FirstName string `json:"first_name,omitempty"`
							LastName  string `json:"last_name"`
						}
						return &User{}
					})(),
				),
			},
			expectedJSON: `{"settings":{"number_of_replicas":2},"mappings":{"customer":{"properties":{"first_name":{"type":"text"},"last_name":{"type":"text"}}}}}`,
		},
		{
			title: `Index definition with a setting and a customer mapping with an case insensitive reference ID`,
			defFuncs: []elasticorm.IndexDefinitionFunc{
				elasticorm.SetNumberOfReplicas(2),
				elasticorm.AddMappingFromStruct(
					`customer`,
					(func() interface{} {
						type User struct {
							FirstName string `json:"first_name,omitempty"`
							Email     string `json:"email" elasticorm:"ref_id,case_sensitive=false"`
						}
						return &User{}
					})(),
				),
			},
			expectedJSON: `{"settings":{"number_of_replicas":2,"analysis":{"analyzer":{"case_insensitive_ref_id":{"type":"custom","tokenizer":"keyword","filter":["lowercase"]}}}},"mappings":{"customer":{"properties":{"email":{"type":"text","analyzer":"case_insensitive_ref_id"},"first_name":{"type":"text"}}}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			def, err := elasticorm.NewIndexDefinition(tt.defFuncs...)
			ok(t, err)
			actualJSON, err := json.Marshal(def)
			ok(t, err)
			equals(t, tt.expectedJSON, string(actualJSON))
		})
	}
}
