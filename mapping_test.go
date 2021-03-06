package elasticorm_test

import (
	"testing"
	"time"

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
		ExpectedError: errors.Wrap(elasticorm.ErrInvalidOption, `parsing option foo=date failed`),
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
	mappingTestCase{
		Title: `For a struct with multiple elasticorm tag options - for analyzer and type`,
		Input: func() interface{} {
			type User struct {
				FirstName   string     `json:"first_name"`
				DateOfBirth *time.Time `json:"date" elasticorm:"analyzer=simple,type=date"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"date":{"type":"date","analyzer":"simple"},"first_name":{"type":"text"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with an elastic ID - which should not be mapped`,
		Input: func() interface{} {
			type User struct {
				ID        string `elasticorm:"id"`
				FirstName string `json:"first_name"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"first_name":{"type":"text"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a nested struct with an anonymous`,
		Input: func() interface{} {
			type User struct {
				Gender string `json:"gender" elasticorm:"type=keyword"`
				Name   struct {
					Title    string `json:"title" elasticorm:"type=keyword"`
					LastName string `json:"last_name"`
				} `json:"name"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"gender":{"type":"keyword"},"name":{"type":"object","properties":{"last_name":{"type":"text"},"title":{"type":"keyword"}}}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a nested struct with an own definition`,
		Input: func() interface{} {
			type Name struct {
				Title    string `json:"title" elasticorm:"type=keyword"`
				LastName string `json:"last_name"`
			}
			type User struct {
				Gender string `json:"gender" elasticorm:"type=keyword"`
				Name   Name   `json:"name"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"gender":{"type":"keyword"},"name":{"type":"object","properties":{"last_name":{"type":"text"},"title":{"type":"keyword"}}}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a nested struct with a pointer to the sub struct`,
		Input: func() interface{} {
			type Name struct {
				Title    string `json:"title" elasticorm:"type=keyword"`
				LastName string `json:"last_name"`
			}
			type User struct {
				Gender string `json:"gender" elasticorm:"type=keyword"`
				Name   *Name  `json:"name"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"gender":{"type":"keyword"},"name":{"type":"object","properties":{"last_name":{"type":"text"},"title":{"type":"keyword"}}}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with an field which is omitted in JSON marshalling`,
		Input: func() interface{} {
			type User struct {
				FirstName string `json:"-"`
				LastName  string `json:"last_name"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"last_name":{"type":"text"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with a not configured *time.Time`,
		Input: func() interface{} {
			type User struct {
				FirstName   string     `json:"first_name"`
				DateOfBirth *time.Time `json:"date" elasticorm:"analyzer=simple"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"date":{"type":"date","analyzer":"simple"},"first_name":{"type":"text"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with a slice of elements`,
		Input: func() interface{} {
			type Coffee struct {
				Brand string `json:"brand" elasticorm:"type=keyword"`
			}
			type User struct {
				FirstName string   `json:"first_name"`
				Coffees   []Coffee `json:"coffees"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"coffees":{"type":"nested","properties":{"brand":{"type":"keyword"}}},"first_name":{"type":"text"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with a sortable text field`,
		Input: func() interface{} {
			type User struct {
				FirstName string `json:"first_name" elasticorm:"sortable"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"first_name":{"type":"text","fields":{"raw":{"type":"keyword"}}}}}`,
		ExpectedError: nil,
		// TODO sortable as tag for a keyword field
	},
	mappingTestCase{
		Title: `For a struct with a slice of strings`,
		Input: func() interface{} {
			type User struct {
				FirstName string   `json:"first_name"`
				Interests []string `json:"interests"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"first_name":{"type":"text"},"interests":{"type":"text"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with a boolean`,
		Input: func() interface{} {
			type User struct {
				FirstName string `json:"first_name"`
				IsActive  bool   `json:"is_active"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"first_name":{"type":"text"},"is_active":{"type":"boolean"}}}`,
		ExpectedError: nil,
	},
	mappingTestCase{
		Title: `For a struct with numeric types`,
		Input: func() interface{} {
			type User struct {
				Float32   float32 `json:"float_32"`
				Float64   float64 `json:"float_64"`
				Integer   int     `json:"integer"`
				Integer16 int16   `json:"integer_16"`
				Integer32 int32   `json:"integer_32"`
				Integer64 int64   `json:"integer_64"`
				Integer8  int8    `json:"integer_8"`
			}
			return &User{}
		}(),
		ExpectedJSON:  `{"properties":{"float_32":{"type":"float"},"float_64":{"type":"double"},"integer":{"type":"integer"},"integer_16":{"type":"short"},"integer_32":{"type":"integer"},"integer_64":{"type":"long"},"integer_8":{"type":"byte"}}}`,
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
