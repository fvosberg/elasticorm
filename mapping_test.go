package elasticorm_test

import (
	"testing"

	"encoding/json"

	"github.com/fvosberg/elasticorm"
	"github.com/pkg/errors"
)

type mappingTestCase struct {
	Title         string
	Input         interface{}
	ExpectedJSON  string
	ExpectedError error
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
		ExpectedJSON:  `{"properties":{"FirstName":{"type":"text"},"LastName":{"type":"text"}}}`,
		ExpectedError: nil,
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
		ExpectedJSON:  `{"properties":{"first_name":{"type":"text"},"last_name":{"type":"text"}}}`,
		ExpectedError: nil,
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
		ExpectedJSON:  `{"properties":{"date":{"type":"date"},"first_name":{"type":"text"},"last_name":{"type":"text"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with invalid elasticorm tag options`,
		Input: func() interface{} {
			type User struct {
				FirstName   string `json:"first_name"`
				LastName    string `json:"last_name"`
				DateOfBirth string `json:"date" elasticorm:"foo=date"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"date":{"type":"text"},"first_name":{"type":"text"},"last_name":{"type":"text"}}}`,
		ExpectedError: errors.Wrap(elasticorm.InvalidOptionErr, `parsing option foo=date failed`),
	},
	mappingTestCase{
		Title: `For a struct with elasticorm tag option for analyzer`,
		Input: func() interface{} {
			type User struct {
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name" elasticorm:"analyzer=simple"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"first_name":{"type":"text"},"last_name":{"type":"text","analyzer":"simple"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with json tag with option`,
		Input: func() interface{} {
			type User struct {
				FirstName string `json:"first_name,omitempty"`
				LastName  string `json:"last_name"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"first_name":{"type":"text"},"last_name":{"type":"text"}}}`,
		ExpectedError: nil,
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
			mapping, actErr := elasticorm.MappingFromStruct(tt.Input)

			actualJSON, err := json.Marshal(mapping)
			ok(t, err)

			if tt.ExpectedError != nil && actErr != nil {
				equals(t, tt.ExpectedError.Error(), actErr.Error())
			} else {
				equals(t, tt.ExpectedError, actErr)
			}
			equals(t, tt.ExpectedJSON, string(actualJSON))
		})
	}
}
