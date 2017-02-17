package elasticorm_test

import (
	"testing"

	"encoding/json"

	"github.com/fvosberg/elasticorm"
)

type mappingTestCase struct {
	Title        string
	Input        interface{}
	ExpectedJSON string
}

var mappingTestCases []mappingTestCase = []mappingTestCase{
	mappingTestCase{
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
	mappingTestCase{
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
	mappingTestCase{
		Title: `For a struct with json and elasticorm tags`,
		Input: func() interface{} {
			type User struct {
				FirstName   string `json:"first_name"`
				LastName    string `json:"last_name"`
				DateOfBirth string `json:"date" elasticorm:"type=date"`
			}
			return &User{}
		}(),
		ExpectedJSON: `{"properties":{"date":{"type":"date"},"first_name":{"type":"text"},"last_name":{"type":"text"}}}`,
	},
	/*
		TODO
		{
			Input:        User{},
			ExpectedJSON: `{"properties":{"FirstName":{"type":"text"},"LastName":{"type":"text"}}}`,
		},
	*/
}

func TestNewMappingFromStruct(t *testing.T) {
	for _, tt := range mappingTestCases {
		t.Run(tt.Title, func(t *testing.T) {
			mapping := elasticorm.MappingFromStruct(tt.Input)

			actualJSON, err := json.Marshal(mapping)
			ok(t, err)
			equals(t, tt.ExpectedJSON, string(actualJSON))
		})
	}
}
