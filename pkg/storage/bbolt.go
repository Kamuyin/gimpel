// Package storage provides shared database utilities using bbolt.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

type DB struct {
	db   *bbolt.DB
	path string
}

type Options struct {
	Path        string
	FileMode    os.FileMode
	Timeout     time.Duration
	ReadOnly    bool
	NoSync      bool
	InitBuckets []string
}

func DefaultOptions(path string) *Options {
	return &Options{
		Path:     path,
		FileMode: 0600,
		Timeout:  5 * time.Second,
	}
}

func Open(opts *Options) (*DB, error) {
	dir := filepath.Dir(opts.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	boltOpts := &bbolt.Options{
		Timeout:  opts.Timeout,
		ReadOnly: opts.ReadOnly,
		NoSync:   opts.NoSync,
	}

	db, err := bbolt.Open(opts.Path, opts.FileMode, boltOpts)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	wrapper := &DB{
		db:   db,
		path: opts.Path,
	}

	if len(opts.InitBuckets) > 0 {
		if err := wrapper.InitBuckets(opts.InitBuckets...); err != nil {
			db.Close()
			return nil, fmt.Errorf("initializing buckets: %w", err)
		}
	}

	return wrapper, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Path() string {
	return d.path
}

func (d *DB) Bolt() *bbolt.DB {
	return d.db
}

func (d *DB) InitBuckets(names ...string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range names {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return fmt.Errorf("creating bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

func (d *DB) Put(bucket, key string, value []byte) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.Put([]byte(key), value)
	})
}

func (d *DB) Get(bucket, key string) ([]byte, error) {
	var value []byte
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		v := b.Get([]byte(key))
		if v != nil {
			value = make([]byte, len(v))
			copy(value, v)
		}
		return nil
	})
	return value, err
}

func (d *DB) Delete(bucket, key string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.Delete([]byte(key))
	})
}

func (d *DB) Exists(bucket, key string) (bool, error) {
	var exists bool
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		exists = b.Get([]byte(key)) != nil
		return nil
	})
	return exists, err
}

func (d *DB) ForEach(bucket string, fn func(key, value []byte) error) error {
	return d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.ForEach(fn)
	})
}

func (d *DB) Count(bucket string) (int, error) {
	var count int
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		count = b.Stats().KeyN
		return nil
	})
	return count, err
}

func (d *DB) PutJSON(bucket, key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	return d.Put(bucket, key, data)
}

func (d *DB) GetJSON(bucket, key string, v interface{}) error {
	data, err := d.Get(bucket, key)
	if err != nil {
		return err
	}
	if data == nil {
		return ErrNotFound
	}
	return json.Unmarshal(data, v)
}

func (d *DB) ListKeys(bucket string) ([]string, error) {
	var keys []string
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return keys, err
}

func (d *DB) Update(fn func(tx *bbolt.Tx) error) error {
	return d.db.Update(fn)
}

func (d *DB) View(fn func(tx *bbolt.Tx) error) error {
	return d.db.View(fn)
}

func (d *DB) Backup(path string) error {
	return d.db.View(func(tx *bbolt.Tx) error {
		return tx.CopyFile(path, 0600)
	})
}
