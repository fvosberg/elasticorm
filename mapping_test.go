package elasticorm_test

import (
	"testing"

	"encoding/json"

	"github.com/fvosberg/elasticorm"
)

func TestNewMappingFromStruct(t *testing.T) {
	tests := []struct {
		Title        string
		Input        interface{}
		ExpectedJSON string
	}{
		{
			Title: `For a struct without tags`,
			Input: func() interface{} {
				type User struct {
					FirstName string
					LastName  string
				}
				return &User{}
			}(),
			ExpectedJSON: `{"properties":{"FirstName":{"type":"text"},"LastName":{"type":"text"}}}`,
		},
		{
			Title: `For a struct with json tags`,
			Input: func() interface{} {
				type User struct {
					FirstName string `json:"first_name"`
					LastName  string `json:"last_name"`
				}
				return &User{}
			}(),
			ExpectedJSON: `{"properties":{"first_name":{"type":"text"},"last_name":{"type":"text"}}}`,
		},
		/*
			TODO
			{
				Input:        User{},
				ExpectedJSON: `{"properties":{"FirstName":{"type":"text"},"LastName":{"type":"text"}}}`,
			},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.Title, func(t *testing.T) {
			mapping := elasticorm.MappingFromStruct(tt.Input)

			actualJSON, err := json.Marshal(mapping)
			ok(t, err)
			equals(t, tt.ExpectedJSON, string(actualJSON))
		})
	}
}
