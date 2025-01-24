package redis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lansfy/gonkex/storage/addons/redis/parser"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
}

type StorageOpts struct {
}

func NewStorage(client *redis.Client, opts *StorageOpts) *Storage {
	return &Storage{
		client: client,
	}
}

func (l *Storage) GetType() string {
	return "redis"
}

func (l *Storage) LoadFixtures(location string, names []string) error {
	ctx := parser.NewContext()
	fileParser := parser.New([]string{location})
	fixtureList, err := fileParser.ParseFiles(ctx, names)
	if err != nil {
		return err
	}

	return l.loadData(fixtureList)
}

func (f *Storage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	return nil, errors.New("not implemented")
}

func (l *Storage) loadKeys(ctx context.Context, pipe redis.Pipeliner, db parser.Database) error {
	if db.Keys == nil {
		return nil
	}
	for k, v := range db.Keys.Values {
		if err := pipe.Set(ctx, k, v.Value.Value, v.Expiration).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (l *Storage) loadSets(ctx context.Context, pipe redis.Pipeliner, db parser.Database) error {
	if db.Sets == nil {
		return nil
	}
	for setKey, setRecord := range db.Sets.Values {
		values := make([]interface{}, 0, len(setRecord.Values))
		for _, v := range setRecord.Values {
			values = append(values, v.Value.Value)
		}
		if err := pipe.SAdd(ctx, setKey, values).Err(); err != nil {
			return err
		}
		if setRecord.Expiration > 0 {
			if err := pipe.Expire(ctx, setKey, setRecord.Expiration).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Storage) loadHashes(ctx context.Context, pipe redis.Pipeliner, db parser.Database) error {
	if db.Hashes == nil {
		return nil
	}
	for key, record := range db.Hashes.Values {
		values := make([]interface{}, 0, len(record.Values)*2)
		for _, v := range record.Values {
			values = append(values, v.Key.Value, v.Value.Value)
		}
		if err := pipe.HSet(ctx, key, values...).Err(); err != nil {
			return err
		}
		if record.Expiration > 0 {
			if err := pipe.Expire(ctx, key, record.Expiration).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Storage) loadLists(ctx context.Context, pipe redis.Pipeliner, db parser.Database) error {
	if db.Lists == nil {
		return nil
	}
	for key, record := range db.Lists.Values {
		values := make([]interface{}, 0, len(record.Values))
		for _, v := range record.Values {
			values = append(values, v.Value.Value)
		}
		if err := pipe.RPush(ctx, key, values...).Err(); err != nil {
			return err
		}
		if record.Expiration > 0 {
			if err := pipe.Expire(ctx, key, record.Expiration).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Storage) loadSortedSets(ctx context.Context, pipe redis.Pipeliner, db parser.Database) error {
	if db.ZSets == nil {
		return nil
	}
	for key, record := range db.ZSets.Values {
		values := make([]redis.Z, 0, len(record.Values))
		for _, v := range record.Values {
			values = append(values, redis.Z{
				Score:  v.Score,
				Member: v.Value.Value,
			})
		}
		if err := pipe.ZAdd(ctx, key, values...).Err(); err != nil {
			return err
		}
		if record.Expiration > 0 {
			if err := pipe.Expire(ctx, key, record.Expiration).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Storage) loadRedisDatabase(ctx context.Context, dbID int, db parser.Database, needTruncate bool) error {
	pipe := l.client.Pipeline()
	err := pipe.Select(ctx, dbID).Err()
	if err != nil {
		return err
	}

	if needTruncate {
		if err := pipe.FlushDB(ctx).Err(); err != nil {
			return err
		}
	}

	if err := l.loadKeys(ctx, pipe, db); err != nil {
		return err
	}

	if err := l.loadSets(ctx, pipe, db); err != nil {
		return err
	}

	if err := l.loadHashes(ctx, pipe, db); err != nil {
		return err
	}

	if err := l.loadLists(ctx, pipe, db); err != nil {
		return err
	}

	if err := l.loadSortedSets(ctx, pipe, db); err != nil {
		return err
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (l *Storage) loadData(fixtures []*parser.Fixture) error {
	truncatedDatabases := make(map[int]struct{})

	for _, redisFixture := range fixtures {
		for dbID, db := range redisFixture.Databases {
			var needTruncate bool
			if _, ok := truncatedDatabases[dbID]; !ok {
				truncatedDatabases[dbID] = struct{}{}
				needTruncate = true
			}
			err := l.loadRedisDatabase(context.Background(), dbID, db, needTruncate)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
