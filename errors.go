package elasticorm

import "errors"

var (
	// ErrInvalidOption is returned when a not valid option is used in a elasticorm tag to configure the mapping of a struct
	ErrInvalidOption = errors.New(`Invalid elasticorm option is used`)
	// errIdField is returned when a Mapping for a field is tried to retrived, which should hold the elasticsearch id
	errIdField = errors.New(`No mapping for ID field`)
)
