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
func NewDatastore(esc *elastic.Client, opts ...DatastoreOptFunc) (*Datastore, error) {
	ds := &Datastore{
		elasticClient: esc,
		Ctx:           context.Background(),
	}
	var err error
	for _, opt := range opts {
		err = opt(ds)
	}
	return ds, err
}

func NewDatastoreForURL(URL string, opts ...DatastoreOptFunc) (*Datastore, error) {
	esc, err := elasticClient(URL)
	if err != nil {
		return nil, err
	}
	return NewDatastore(esc, opts...)
}

func elasticClient(URL string) (*elastic.Client, error) {
	c, err := elastic.NewClient(
		elastic.SetURL(URL),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed connecting to elasticsearch via \"%s\"", URL)
	}
	return c, nil
}

// DatastoreOptFunc is used as a parameter to NewDatastore and provides a way of configuration
type DatastoreOptFunc func(*Datastore) error

// QueryOptFunc is used as a parameter
type QueryOptFunc func(*elastic.SearchService) error

// Datstore is an instance to communicate easy with elasticsearch
// It is meant to be used for one struct and helps storing and retrieving it in/from elasticsearch
// It leverages the great elastic package from olivere
type Datastore struct {
	elasticClient   *elastic.Client
	Ctx             context.Context
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
		Do(ds.Ctx)

	if exists || err != nil {
		return err
	}

	err = ds.createIndex()
	if err != nil {
		return err
	}
	ds.Refresh()
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
		Do(ds.Ctx)
	if err != nil || !ack.Acknowledged {
		return errors.Wrap(
			err, fmt.Sprintf("creating elasticsearch index %s failed - %s", ds.indexName, string(JSON)),
		)
	}
	return nil
}

func (ds *Datastore) EnsureIndexDoesntExist() error {
	if ds.indexName == `` {
		return errors.New(`EnsureIndexDoesntExists failed, because no index name is defined`)
	}
	exists, err := ds.elasticClient.
		IndexExists(ds.indexName).
		Do(ds.Ctx)

	if !exists || err != nil {
		return err
	}
	res, err := ds.elasticClient.
		DeleteIndex(ds.indexName).
		Do(context.Background())
	if err != nil || !res.Acknowledged {
		return errors.Wrap(
			err, fmt.Sprintf("deleting elasticsearch index %s failed", ds.indexName),
		)
	}
	return nil
}

func (ds *Datastore) CleanUp() error {
	err := ds.EnsureIndexDoesntExist()
	if err != nil {
		return err
	}
	err = ds.EnsureIndexExists()
	if err != nil {
		return err
	}
	return ds.Refresh()
}

func (ds *Datastore) Refresh() error {
	_, err := ds.elasticClient.Refresh().Do(ds.Ctx)
	return err
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
		if ds.idFieldName == "" {
			ds.idFieldName = "ID"
		}
		return nil
	}
}

func WithIDField(name string) DatastoreOptFunc {
	return func(ds *Datastore) error {
		ds.idFieldName = name
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
		Do(ds.Ctx)

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
		Do(ds.Ctx)

	if elastic.IsNotFound(err) || (res != nil && !res.Found) {
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
		Do(ds.Ctx)

	return err
}

func (ds *Datastore) FindOneBy(fieldName string, value interface{}, result interface{}, opts ...QueryOptFunc) error {
	elasticFieldName, err := ds.indexDefinition.elasticFieldName(ds.typeName, fieldName)
	if err != nil {
		return err
	}
	q := ds.elasticClient.Search().
		Index(ds.indexName).
		Query(elastic.NewBoolQuery().Filter(elastic.NewTermQuery(elasticFieldName, value))).
		From(0).Size(1)

	for _, opt := range opts {
		err := opt(q)
		if err != nil {
			return err
		}
	}

	res, err := q.Do(ds.Ctx)
	if err != nil {
		return err
	}

	if res.TotalHits() < 1 {
		return ErrNotFound
	}
	return ds.decodeElasticResponse(res.Hits.Hits[0].Source, res.Hits.Hits[0].Id, result)
}

func (ds *Datastore) FindAll(results interface{}, opts ...QueryOptFunc) error {
	q := ds.elasticClient.Search().
		Index(ds.indexName).
		Query(elastic.NewMatchAllQuery())

	for _, opt := range opts {
		err := opt(q)
		if err != nil {
			return err
		}
	}

	res, err := q.Do(ds.Ctx)
	if err != nil {
		return err
	}
	return ds.decodeElasticResponses(res.Hits.Hits, results)
}

func (ds *Datastore) SetSorting(fieldName string, order string) QueryOptFunc {
	return func(srv *elastic.SearchService) error {
		if order != `asc` && order != `desc` {
			return errors.New(`sorting order must be asc or desc`)
		}
		elasticFieldName, err := ds.indexDefinition.elasticFieldName(ds.typeName, fieldName)
		// TODO loosen coupling with indexDefinition
		if err != nil {
			return err
		}
		srv.Sort(elasticFieldName+`.raw`, order == `asc`)
		return nil
	}
}

func (ds *Datastore) FilterByField(fieldName string, value interface{}) QueryOptFunc {
	return func(srv *elastic.SearchService) error {
		// TODO loosen coupling with indexDefinition
		elasticFieldName, err := ds.indexDefinition.elasticFieldName(ds.typeName, fieldName)
		if err != nil {
			return err
		}
		q := elastic.NewBoolQuery()
		q.Filter(elastic.NewTermQuery(elasticFieldName, value))
		srv.Query(q)
		return nil
	}
}

func (ds *Datastore) Offset(offset int) QueryOptFunc {
	return func(srv *elastic.SearchService) error {
		srv.From(offset)
		return nil
	}
}

func (ds *Datastore) Limit(limit int) QueryOptFunc {
	return func(srv *elastic.SearchService) error {
		srv.Size(limit)
		return nil
	}
}

type BoundingBox struct {
	Top    float64
	Left   float64
	Bottom float64
	Right  float64
}

func NewBoundingBox(top, right, bottom, left float64) BoundingBox {
	return BoundingBox{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

func (ds *Datastore) FindByGeoBoundingBox(fieldName string, box BoundingBox, results interface{}) error {
	elasticFieldName, err := ds.indexDefinition.elasticFieldName(ds.typeName, fieldName)
	if err != nil {
		return err
	}
	query := elastic.NewGeoBoundingBoxQuery(elasticFieldName).
		TopLeft(box.Top, box.Left).
		BottomRight(box.Bottom, box.Right)

	res, err := ds.elasticClient.Search().
		Index(ds.indexName).
		// TODO query type
		Query(elastic.NewBoolQuery().Filter(query)).
		Do(ds.Ctx)

	if err != nil {
		return err
	}

	return ds.decodeElasticResponses(res.Hits.Hits, results)
}

func (ds *Datastore) FindByGeoDistance(fieldName string, lat float64, lon float64, distance string, results interface{}) error {
	elasticFieldName, err := ds.indexDefinition.elasticFieldName(ds.typeName, fieldName)
	if err != nil {
		return err
	}
	query := elastic.NewGeoDistanceQuery(elasticFieldName).
		Lat(lat).
		Lon(lon).
		Distance(distance)

	res, err := ds.elasticClient.Search().
		Index(ds.indexName).
		// TODO query type
		Query(query).
		SortBy(elastic.NewGeoDistanceSort(elasticFieldName).Point(lat, lon)).
		Do(ds.Ctx)

	if err != nil {
		return err
	}

	return ds.decodeElasticResponses(res.Hits.Hits, results)
}

func (ds *Datastore) decodeElasticResponses(hits []*elastic.SearchHit, results interface{}) error {
	resultsv := reflect.ValueOf(results)
	if resultsv.Kind() != reflect.Ptr || resultsv.Elem().Kind() != reflect.Slice {
		return errors.New("result argument must be a slice address")
	}

	slicev := resultsv.Elem()
	elemt := slicev.Type().Elem()

	for _, hit := range hits {
		elemp := reflect.New(elemt)
		err := ds.decodeElasticResponse(hit.Source, hit.Id, elemp.Interface())
		if err != nil {
			return err
		}
		slicev = reflect.Append(slicev, elemp.Elem())
	}

	resultsv.Elem().Set(slicev.Slice(0, len(hits)))
	return nil
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
