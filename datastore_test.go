package elasticorm_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fvosberg/elasticorm"

	"gopkg.in/olivere/elastic.v5"
)

func TestEnsureIndexExists(t *testing.T) {
	type User struct {
		Name        string `json:"name"`
		DateOfBirth string `json:"date" elasticorm:"type=date"`
	}
	client, err := elastic.NewClient(
		elastic.SetURL(elasticSearchURL),
	)
	t.Logf(
		"Using %s as elasticsearch URL - please provide a running elasticsearch instance under this URL, or configure it with the env variable EDS_ES_URL. Be carefull, all data will be erased.",
		elasticSearchURL,
	)
	ok(t, err)
	_, err = client.DeleteIndex(`_all`).Do(context.Background())
	ok(t, err)
	_, err = client.Refresh().Do(context.Background())
	ok(t, err)

	ds, err := elasticorm.NewDatastore(
		client,
		elasticorm.ForStruct(&User{}),
	)
	ok(t, err)
	err = ds.EnsureIndexExists()
	ok(t, err)

	_, err = client.Refresh().Do(context.Background())
	ok(t, err)
	exists, err := client.IndexExists(`users`).Do(context.Background())
	ok(t, err)
	assert(t, exists, `The index users should exist`)
	actMapping, err := client.GetMapping().Do(context.Background())
	ok(t, err)
	actMappingJSON, err := json.Marshal(actMapping)
	ok(t, err)
	equals(
		t,
		`{"users":{"mappings":{"user":{"properties":{"date":{"type":"date"},"name":{"type":"text"}}}}}}`,
		string(actMappingJSON),
	)
}
