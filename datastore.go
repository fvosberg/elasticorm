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
	goType          reflect.Type
	idFieldName     string          // the name of the structs field to store the ID
	typeName        string          // in elasticsearch
	indexDefinition IndexDefinition // in elasticsearch
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
		ds.goType = reflect.TypeOf(i)
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
		// TODO configure ID
		ds.idFieldName = `ID`
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
	return nameOfType(reflect.TypeOf(myvar))
}

func nameOfType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func (ds *Datastore) Create(o interface{}) error {
	err := ds.isSaveableType(o)
	if err != nil {
		return err
	}
	put, err := ds.elasticClient.Index().
		Index(ds.indexName).
		Type(ds.typeName).
		BodyJson(o).
		Do(context.Background())

	if err != nil {
		return err
	}
	if !put.Created {
		return ErrCreationFailed
	}
	return ds.setID(o, put.Id)
}

func (ds *Datastore) isSaveableType(i interface{}) error {
	rt := reflect.TypeOf(i)
	if nameOfType(rt) != nameOfType(ds.goType) {
		return errors.Wrapf(
			ErrInvalidType,
			`create failed for %s, expected %s`, nameOfType(rt), nameOfType(ds.goType),
		)
	}
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Wrap(ErrInvalidType, `no pointer given`)
	}
	return nil
}

func (ds *Datastore) Find(ID string, result interface{}) error {
	err := ds.isSaveableType(result)
	if err != nil {
		return err
	}
	res, err := ds.elasticClient.Get().
		Index(ds.indexName).
		Type(ds.typeName).
		Id(ID).
		Do(context.Background())

	if !res.Found || elastic.IsNotFound(err) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	return ds.decodeElasticResponse(res.Source, res.Id, result)
}

func (ds *Datastore) Update(o interface{}) error {
	err := ds.isSaveableType(o)
	if err != nil {
		return err
	}
	ID, err := ds.getID(o)
	if err != nil {
		return err
	}
	if ID == `` {
		return errors.New(`can't save struct with empty ID`)
	}
	_, err = ds.elasticClient.Update().
		Index(ds.indexName).
		Type(ds.typeName).
		Id(ID).
		Doc(o).
		Do(context.Background())

	return err
}

func (ds *Datastore) FindOneBy(fieldName string, value interface{}, result interface{}) error {
	elasticFieldName, err := ds.indexDefinition.elasticFieldName(ds.typeName, fieldName)
	if err != nil {
		return err
	}
	res, err := ds.elasticClient.Search().
		Index(ds.indexName).
		Query(elastic.NewBoolQuery().Filter(elastic.NewTermQuery(elasticFieldName, value))).
		From(0).Size(1).
		Do(context.Background())
	if err != nil {
		return err
	}
	if res.TotalHits() < 1 {
		return ErrNotFound
	}
	return ds.decodeElasticResponse(res.Hits.Hits[0].Source, res.Hits.Hits[0].Id, result)
}

func (ds *Datastore) setID(o interface{}, ID string) error {
	eo := reflect.ValueOf(o).Elem()
	if eo.Kind() != reflect.Struct {
		return errors.Wrap(ErrInvalidType, `setID failed`)
	}
	idField := eo.FieldByName(ds.idFieldName)
	if !idField.IsValid() || !idField.CanSet() || idField.Kind() != reflect.String {
		return ErrInvalidIDField
	}
	idField.SetString(ID)
	return nil
}

func (ds *Datastore) getID(o interface{}) (string, error) {
	eo := reflect.ValueOf(o).Elem()
	if eo.Kind() != reflect.Struct {
		return ``, errors.Wrap(ErrInvalidType, `getID failed`)
	}
	idField := eo.FieldByName(ds.idFieldName)
	if !idField.IsValid() || idField.Kind() != reflect.String {
		return ``, ErrInvalidIDField
	}
	return idField.String(), nil
}

func (ds *Datastore) decodeElasticResponse(source *json.RawMessage, ID string, o interface{}) error {
	err := json.Unmarshal(*source, o)
	if err != nil {
		return err
	}
	ds.setID(o, ID)
	return nil
}
