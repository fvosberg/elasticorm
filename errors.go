package elasticorm

import "errors"

var (
	// ErrInvalidOption is returned when a not valid option is used in a elasticorm tag to configure the mapping of a struct
	ErrInvalidOption = errors.New(`Invalid elasticorm option is used`)

	// ErrInvalidType is returned when you try to save a struct with a datastore which has been initialized for another struct
	ErrInvalidType = errors.New(`Invalid type for this datastore`)

	// ErrInvalidResultType is returned when no pointer is passed to a find method
	ErrInvalidResultType = errors.New(`invalid result type`)

	// ErrInvalidIDField is returned when the defined ID field can't be set
	ErrInvalidIDField = errors.New(`invalid ID field`)

	// ErrNotFound is returned when no record could be found
	ErrNotFound = errors.New(`not found`)

	// ErrCreationFailed is returned by the Create method, when there was no error by the elastic client, but the record could not have been created - TODO when does this happen?
	ErrCreationFailed = errors.New(`creation of new elasticsearch record failed`)

	// errIdField is returned when a Mapping for a field is tried to retrived, which should hold the elasticsearch id
	errIdField = errors.New(`No mapping for ID field`)
)
