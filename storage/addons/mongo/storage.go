package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lansfy/gonkex/storage/addons/fixtures"
	"go.mongodb.org/mongo-driver/mongo"
)

type Storage struct {
	client *mongo.Client
}

type StorageOpts struct {
}

func NewStorage(client *mongo.Client, opts *StorageOpts) *Storage {
	return &Storage{
		client: client,
	}
}

func (f *Storage) GetType() string {
	return "mongo"
}

func (f *Storage) LoadFixtures(location string, names []string) error {
	opts := &fixtures.LoadDataOpts{
		AllowedTypes: []string{"collections"},
	}

	coll, err := fixtures.LoadData(fixtures.CreateFileLoader(location), names, opts)
	if err != nil {
		return fmt.Errorf("load fixtures: %w", err)
	}

	// truncate first
	err = f.truncateCollections(coll)
	if err != nil {
		return fmt.Errorf("trancate collections: %w", err)
	}

	// then load data
	for _, cl := range coll {
		if len(cl.Items) == 0 {
			continue
		}

		err = insertCollection(f.client, cl)
		if err != nil {
			return fmt.Errorf("load collection '%s': %w", cl.Name, err)
		}
	}

	return nil
}

func (f *Storage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	return nil, errors.New("not implemented")
}

func (f *Storage) truncateCollections(collections []*fixtures.Collection) error {
	truncated := map[string]bool{}
	for _, cl := range collections {
		name := newCollectionName(cl.Name)
		if _, ok := truncated[name.getFullName()]; ok {
			continue
		}

		if err := truncate(f.client, name.database, name.name); err != nil {
			return err
		}

		truncated[name.getFullName()] = true
	}

	return nil
}

func truncate(client *mongo.Client, database string, collection string) error {
	cl := client.Database(database).Collection(collection)
	return cl.Drop(context.Background())
}

func insertCollection(client *mongo.Client, cl *fixtures.Collection) error {
	documents := make([]interface{}, len(cl.Items))
	for idx, doc := range cl.Items {
		documents[idx] = doc
	}

	name := newCollectionName(cl.Name)
	conn := client.Database(name.database).Collection(name.name)
	_, err := conn.InsertMany(context.Background(), documents)
	return err
}

type collectionName struct {
	name     string
	database string
}

func newCollectionName(source string) *collectionName {
	parts := strings.SplitN(source, ".", 2)
	if len(parts) == 1 {
		parts = append(parts, parts[0])
		parts[0] = "public"
	} else if parts[0] == "" {
		parts[0] = "public"
	}

	return &collectionName{
		database: parts[0],
		name:     parts[1],
	}
}

func (t *collectionName) getFullName() string {
	return fmt.Sprintf("%q.%q", t.database, t.name)
}
