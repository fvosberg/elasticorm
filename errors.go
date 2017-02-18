package elasticorm

import "errors"

var (
	// ErrInvalidOption is returned when a not valid option is used in a elasticorm tag to configure the mapping of a struct
	ErrInvalidOption = errors.New(`Invalid elasticorm option is used`)
)
