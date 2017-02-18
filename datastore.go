package elasticorm

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	elastic "gopkg.in/olivere/elastic.v5"
)

// NewDatastore returns a fresh instance of an elasticorm datastore
func NewDatastore(esc *elastic.Client, options ...DatastoreOptFunc) (*Datastore, error) {
	ds := &Datastore{
		elasticClient: esc,
	}
	var err error
	for _, opt := range options {
		err = opt(ds)
	}
	return ds, err
}

// DatastoreOptFunc is used as a parameter to NewDatastore and provides a way of configuration
type DatastoreOptFunc func(*Datastore) error

// Datstore is an instance to communicate easy with elasticsearch
// It is meant to be used for one struct and helps storing and retrieving it in/from elasticsearch
// It leverages the great elastic package from olivere
type Datastore struct {
	elasticClient   *elastic.Client
	indexName       string
	typeName        string
	indexDefinition IndexDefinition
}

// EnsureIndexExists checks wether the needed index for this datastore exists. It it doesn't it gets created
// the name of the datastore is determined by the structs name (+ plural s)
func (ds *Datastore) EnsureIndexExists() error {
	if ds.indexName == `` {
		return errors.New(`EnsureIndexExists failed, because no index name is defined`)
	}
	exists, err := ds.elasticClient.
		IndexExists(ds.indexName).
		Do(context.Background())

	if exists || err != nil {
		return err
	}

	err = ds.createIndex()
	if err != nil {
		return err
	}
	ds.elasticClient.Refresh().Do(context.Background())
	return nil
}

func (ds *Datastore) createIndex() error {
	JSON, err := json.Marshal(ds.indexDefinition)
	if err != nil {
		return err
	}
	ack, err := ds.elasticClient.
		CreateIndex(ds.indexName).
		BodyString(string(JSON)).
		Do(context.Background())
	if err != nil || !ack.Acknowledged {
		return errors.Wrap(
			err, fmt.Sprintf("creating elasticsearch index %s failed - %s", ds.indexName, string(JSON)),
		)
	}
	return nil
}

// ForStruct generates a DatastoreOptFunc
// It is used to generate a default mapping, type name and index name by analysing a provided struct
func ForStruct(i interface{}) DatastoreOptFunc {
	return func(ds *Datastore) error {
		typeName, err := typeNameFromStruct(i)
		if err != nil {
			return err
		}
		ds.typeName = typeName
		ds.indexName = typeName + `s`
		indexDefinition, err := NewIndexDefinition(
			AddMappingFromStruct(ds.typeName, i),
		)
		if err != nil {
			return err
		}
		ds.indexDefinition = indexDefinition
		return nil
	}
}

func typeNameFromStruct(i interface{}) (string, error) {
	typeName := getType(i)
	if typeName == `` {
		return ``, errors.New(`Could not determine type name from struct`)
	}

	typeName = strings.ToLower(typeName)
	typeName = strings.TrimLeft(typeName, `*`)
	return typeName, nil
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
