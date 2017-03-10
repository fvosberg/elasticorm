package elasticorm_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/fvosberg/elasticorm"

	"gopkg.in/olivere/elastic.v5"
)

func TestDatastoreEnsureIndexExists(t *testing.T) {
	type User struct {
		Name        string `json:"name"`
		DateOfBirth string `json:"date" elasticorm:"type=date"`
	}
	client := elasticClient(t)
	deleteAllIndices(t, client)

	ds, err := elasticorm.NewDatastore(
		client,
		elasticorm.ForStruct(&User{}),
	)
	ok(t, err)
	err = ds.EnsureIndexExists()
	ok(t, err)

	indexExists(t, client, `users`)
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

func TestDatastoreCreateAUser(t *testing.T) {
	type User struct {
		ID        string `json:"id" elasticorm:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	elasticClient, ds := initDatastore(t, &User{})

	user := &User{
		// TODO test error on setted ID
		FirstName: `Foobar`,
		LastName:  `Barfoo`,
	}
	err := ds.Create(user)
	ok(t, err)

	if user.ID == `` {
		t.Error(`The ID of the user should be set after persisting`)
		t.FailNow()
	}
	_, err = elasticClient.Refresh().Do(context.Background())
	ok(t, err)
	gotUser := &User{}
	err = ds.Find(user.ID, gotUser)
	ok(t, err)
	equals(t, *user, *gotUser)
}

func TestDatastoreUpdateAUser(t *testing.T) {
	type User struct {
		ID        string `json:"id" elasticorm:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	_, ds := initDatastore(t, &User{})

	u := &User{
		FirstName: `Pre Firstname`,
		LastName:  `Lastname`,
	}
	err := ds.Create(u)
	ok(t, err)
	u.FirstName = `Post Firstname`
	err = ds.Update(u)
	ok(t, err)

	if u.ID == `` {
		t.Error(`The ID of the user should not be empty`)
		t.FailNow()
	}
	gotUser := User{}
	err = ds.Find(u.ID, &gotUser)
	ok(t, err)
	equals(t, `Post Firstname`, gotUser.FirstName)
}

func TestDatastoreFindOneBy(t *testing.T) {
	type User struct {
		ID        string `json:"id" elasticorm:"id"`
		FirstName string `json:"first_name" elasticorm:"sortable"`
		Email     string `json:"email" elasticorm:"type=keyword"`
	}
	tests := []struct {
		title         string
		users         []User
		searchField   string
		searchValue   string
		shouldFind    User
		expectedError error
		SortBy        string
		Ordering      string
	}{
		{
			title: `Find a user by email`,
			users: []User{
				User{FirstName: `Wrong user`, Email: `wrong1@bar.com`},
				User{FirstName: `The first name`, Email: `foo@bar.com`},
				User{FirstName: `Wrong user`, Email: `wrong2@bar.com`},
			},
			searchField: `Email`,
			searchValue: `foo@bar.com`,
			shouldFind:  User{FirstName: `The first name`, Email: `foo@bar.com`},
		},
		{
			title:         `Don't find a user by wrong email`,
			users:         []User{User{FirstName: `The first name`, Email: `foo@bar.com`}},
			searchField:   `Email`,
			searchValue:   `bar@foo.com`,
			shouldFind:    User{},
			expectedError: elasticorm.ErrNotFound,
		},
		{
			title:         `Search for a field which doesn't exist`,
			users:         []User{User{FirstName: `The first name`, Email: `foo@bar.com`}},
			searchField:   `email`,
			searchValue:   `foo@bar.com`,
			shouldFind:    User{},
			expectedError: errors.New(`Mapping configuration has no mapping for struct field`),
		},
		{
			title: `Search with sort - asc`,
			users: []User{
				User{FirstName: `ABC`, Email: `match@bar.com`},
				User{FirstName: `DEF`, Email: `not@bar.com`},
				User{FirstName: `GHI`, Email: `match@bar.com`},
			},
			searchField: `Email`,
			searchValue: `match@bar.com`,
			shouldFind:  User{FirstName: `ABC`, Email: `match@bar.com`},
			SortBy:      `FirstName`,
			Ordering:    `asc`,
		},
		{
			title: `Search with sort - desc`,
			users: []User{
				User{FirstName: `ABC`, Email: `match@bar.com`},
				User{FirstName: `DEF`, Email: `not@bar.com`},
				User{FirstName: `GHI`, Email: `match@bar.com`},
			},
			searchField: `Email`,
			searchValue: `match@bar.com`,
			shouldFind:  User{FirstName: `GHI`, Email: `match@bar.com`},
			SortBy:      `FirstName`,
			Ordering:    `desc`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			elasticClient, ds := initDatastore(t, &User{})
			for _, user := range tt.users {
				err := ds.Create(&user)
				ok(t, err)
				elasticClient.Refresh().Do(context.Background())
			}

			found := User{}
			opts := []elasticorm.QueryOptFunc{}
			if tt.SortBy != `` || tt.Ordering != `` {
				opts = append(opts, ds.SetSorting(tt.SortBy, tt.Ordering))
			}
			err := ds.FindOneBy(tt.searchField, tt.searchValue, &found, opts...)

			if tt.expectedError != nil {
				equals(t, tt.expectedError.Error(), err.Error())
			} else {
				ok(t, err)
			}
			tt.shouldFind.ID = found.ID
			equals(t, tt.shouldFind, found)
		})
	}
}

func TestFindByGeoBoundingBox(t *testing.T) {
	// TODO support deeper nested location structs like User.Home.Location
	// TODO check search on non geopoint
	type Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
	type User struct {
		ID       string    `json:"id" elasticorm:"id"`
		Name     string    `json:"name"`
		Location *Location `json:"loc" elasticorm:"type=geo_point"`
	}

	_, ds := initDatastore(t, &User{})
	err := ds.Create(&User{Name: "Juister", Location: &Location{Lat: 53.679598, Lon: 6.994391}})
	ok(t, err)
	err = ds.Create(&User{Name: "Swimmer", Location: &Location{Lat: 53.693986, Lon: 6.992063}})
	ok(t, err)
	ds.Refresh()
	bottomLeft := Location{
		Lat: 53.672103,
		Lon: 6.962326,
	}
	topRight := Location{
		Lat: 53.685006,
		Lon: 7.017360,
	}

	found := []User{}
	err = ds.FindByGeoBoundingBox(
		`Location`,
		elasticorm.NewBoundingBox(topRight.Lat, topRight.Lon, bottomLeft.Lat, bottomLeft.Lon),
		&found,
	)
	ok(t, err)

	equals(t, 1, len(found))
	equals(t, "Juister", found[0].Name)
}

func TestFindByGeoDistance(t *testing.T) {
	type Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
	type Isle struct {
		ID       string    `json:"id" elasticorm:"id"`
		Name     string    `json:"name"`
		Location *Location `json:"loc" elasticorm:"type=geo_point"`
	}

	_, ds := initDatastore(t, &Isle{})
	err := ds.Create(&Isle{Name: "Memmert", Location: &Location{Lat: 53.640652, Lon: 6.887995}})
	ok(t, err)
	err = ds.Create(&Isle{Name: "Juist", Location: &Location{Lat: 53.681747, Lon: 7.008158}})
	ok(t, err)
	err = ds.Create(&Isle{Name: "Borkum", Location: &Location{Lat: 53.600230, Lon: 6.711053}})
	ok(t, err)
	err = ds.Create(&Isle{Name: "Langeoog", Location: &Location{Lat: 53.743725, Lon: 7.481725}})
	ok(t, err)
	ds.Refresh()

	found := []Isle{}
	err = ds.FindByGeoDistance(
		`Location`,
		53.666499,
		7.050261,
		`11.1km`,
		&found,
	)
	ok(t, err)

	equals(t, 2, len(found))
	equals(t, "Juist", found[0].Name)
	equals(t, "Memmert", found[1].Name)
}

func TestFindAll(t *testing.T) {
	assertions := []struct {
		Offset        int
		Limit         int
		Order         string
		ExpectedNames []string
	}{
		{
			Offset: 0,
			Limit:  10,
			Order:  `asc`,
			ExpectedNames: []string{
				`Unknown No. 0`,
				`Unknown No. 1`,
				`Unknown No. 2`,
				`Unknown No. 3`,
				`Unknown No. 4`,
				`Unknown No. 5`,
				`Unknown No. 6`,
				`Unknown No. 7`,
				`Unknown No. 8`,
				`Unknown No. 9`,
			},
		},
		{
			Offset: 0,
			Limit:  3,
			Order:  `asc`,
			ExpectedNames: []string{
				`Unknown No. 0`,
				`Unknown No. 1`,
				`Unknown No. 2`,
			},
		},
		{
			Offset: 3,
			Limit:  3,
			Order:  `asc`,
			ExpectedNames: []string{
				`Unknown No. 3`,
				`Unknown No. 4`,
				`Unknown No. 5`,
			},
		},
		{
			Offset: 3,
			Limit:  3,
			Order:  `desc`,
			ExpectedNames: []string{
				`Unknown No. 6`,
				`Unknown No. 5`,
				`Unknown No. 4`,
			},
		},
	}

	type User struct {
		ID   string `json:"id" elasticorm:"id"`
		Name string `json:"name" elasticorm:"type=text,sortable"` // TODO error on not sorted | test with keyword
	}

	_, ds := initDatastore(t, &User{})
	err := ds.CleanUp()
	ok(t, err)
	for i := 0; i < 10; i++ {
		err := ds.Create(&User{
			Name: fmt.Sprintf("Unknown No. %d", i),
		})
		ok(t, err)
		// refresh after each creation to get the desired sorting
		ds.Refresh()
	}

	for _, a := range assertions {
		found := []User{}
		err = ds.FindAll(
			&found,
			ds.Offset(a.Offset),
			ds.Limit(a.Limit),
			ds.SetSorting(`Name`, a.Order),
		)
		ok(t, err)

		equals(t, len(found), len(a.ExpectedNames))
		for i, name := range a.ExpectedNames {
			equals(t, name, found[i].Name)
		}
	}
}

func initDatastore(t *testing.T, i interface{}) (*elastic.Client, *elasticorm.Datastore) {
	client := elasticClient(t)
	deleteAllIndices(t, client)
	ds, err := elasticorm.NewDatastore(
		client,
		elasticorm.ForStruct(i),
	)
	ok(t, err)
	err = ds.EnsureIndexExists()
	ok(t, err)
	return client, ds
}

func elasticClient(t *testing.T) *elastic.Client {
	client, err := elastic.NewClient(
		elastic.SetURL(elasticSearchURL),
		elastic.SetTraceLog(fileLogger(`elastic-trace.log`)),
	)
	if err != nil {
		t.Logf(
			"Using %s as elasticsearch URL - please provide a running elasticsearch instance under this URL, or configure it with the env variable EDS_ES_URL. Be carefull, all data will be erased.",
			elasticSearchURL,
		)
		t.Error(`Could not start elasticsearch` + err.Error())
		t.FailNow()
	}
	return client
}

func fileLogger(name string) *log.Logger {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	logger := log.New(f, ``, log.LstdFlags)
	return logger
}

func deleteAllIndices(t *testing.T, c *elastic.Client) {
	_, err := c.DeleteIndex(`_all`).Do(context.Background())
	ok(t, err)
	_, err = c.Refresh().Do(context.Background())
	ok(t, err)
}

func indexExists(t *testing.T, c *elastic.Client, indexName string) {
	_, err := c.Refresh().Do(context.Background())
	ok(t, err)
	exists, err := c.IndexExists(indexName).Do(context.Background())
	ok(t, err)
	assert(t, exists, `The index `+indexName+` should exist`)
}
