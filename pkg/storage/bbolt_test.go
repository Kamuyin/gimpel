package storage

import (
	"path/filepath"
	"testing"
)

func TestOpenClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	opts := DefaultOptions(dbPath)
	opts.InitBuckets = []string{"test"}

	db, err := Open(opts)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if db.Path() != dbPath {
		t.Errorf("Path() = %s, want %s", db.Path(), dbPath)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestPutGet(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	if err := db.Put("bucket1", "key1", []byte("value1")); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	value, err := db.Get("bucket1", "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(value) != "value1" {
		t.Errorf("Get = %s, want value1", value)
	}

	value, err = db.Get("bucket1", "nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != nil {
		t.Errorf("Get nonexistent = %v, want nil", value)
	}
}

func TestDelete(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	db.Put("bucket1", "key1", []byte("value1"))

	if err := db.Delete("bucket1", "key1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	value, _ := db.Get("bucket1", "key1")
	if value != nil {
		t.Errorf("After delete, value = %v, want nil", value)
	}
}

func TestExists(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	exists, _ := db.Exists("bucket1", "key1")
	if exists {
		t.Error("Exists returned true for non-existent key")
	}

	db.Put("bucket1", "key1", []byte("value1"))

	exists, _ = db.Exists("bucket1", "key1")
	if !exists {
		t.Error("Exists returned false for existing key")
	}
}

func TestPutGetJSON(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := TestData{Name: "test", Value: 42}

	if err := db.PutJSON("bucket1", "key1", &data); err != nil {
		t.Fatalf("PutJSON failed: %v", err)
	}

	var result TestData
	if err := db.GetJSON("bucket1", "key1", &result); err != nil {
		t.Fatalf("GetJSON failed: %v", err)
	}

	if result.Name != data.Name || result.Value != data.Value {
		t.Errorf("GetJSON = %+v, want %+v", result, data)
	}
}

func TestGetJSONNotFound(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	var result struct{}
	err := db.GetJSON("bucket1", "nonexistent", &result)
	if err != ErrNotFound {
		t.Errorf("GetJSON nonexistent error = %v, want ErrNotFound", err)
	}
}

func TestForEach(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	db.Put("bucket1", "key1", []byte("value1"))
	db.Put("bucket1", "key2", []byte("value2"))
	db.Put("bucket1", "key3", []byte("value3"))

	var keys []string
	err := db.ForEach("bucket1", func(k, v []byte) error {
		keys = append(keys, string(k))
		return nil
	})
	if err != nil {
		t.Fatalf("ForEach failed: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("ForEach found %d keys, want 3", len(keys))
	}
}

func TestListKeys(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	db.Put("bucket1", "a", []byte("1"))
	db.Put("bucket1", "b", []byte("2"))
	db.Put("bucket1", "c", []byte("3"))

	keys, err := db.ListKeys("bucket1")
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("ListKeys returned %d keys, want 3", len(keys))
	}
}

func TestCount(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	count, _ := db.Count("bucket1")
	if count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	db.Put("bucket1", "key1", []byte("1"))
	db.Put("bucket1", "key2", []byte("2"))

	count, _ = db.Count("bucket1")
	if count != 2 {
		t.Errorf("After puts, count = %d, want 2", count)
	}
}

func TestBackup(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	db.Put("bucket1", "key1", []byte("value1"))

	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "backup.db")

	if err := db.Backup(backupPath); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	opts := DefaultOptions(backupPath)
	opts.InitBuckets = []string{"bucket1"}
	backupDB, err := Open(opts)
	if err != nil {
		t.Fatalf("Opening backup failed: %v", err)
	}
	defer backupDB.Close()

	value, _ := backupDB.Get("bucket1", "key1")
	if string(value) != "value1" {
		t.Errorf("Backup value = %s, want value1", value)
	}
}

func TestBucketNotFound(t *testing.T) {
	db := testDB(t, "bucket1")
	defer db.Close()

	_, err := db.Get("nonexistent", "key")
	if err == nil {
		t.Error("Get on nonexistent bucket should fail")
	}

	err = db.Put("nonexistent", "key", []byte("value"))
	if err == nil {
		t.Error("Put on nonexistent bucket should fail")
	}
}

func testDB(t *testing.T, buckets ...string) *DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	opts := DefaultOptions(dbPath)
	opts.InitBuckets = buckets

	db, err := Open(opts)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	return db
}
