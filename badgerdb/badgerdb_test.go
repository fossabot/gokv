package badgerdb_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/philippgille/gokv/badgerdb"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/test"
)

// TestStore tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestStore(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store, path := createStore(t, encoding.JSON)
		defer os.RemoveAll(path)
		defer store.Close()
		test.TestStore(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer os.RemoveAll(path)
		defer store.Close()
		test.TestStore(store, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store, path := createStore(t, encoding.JSON)
		defer os.RemoveAll(path)
		defer store.Close()
		test.TestTypes(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer os.RemoveAll(path)
		defer store.Close()
		test.TestTypes(store, t)
	})
}

// TestStoreConcurrent launches a bunch of goroutines that concurrently work with one store.
// The store works with a single file, so everything should be locked properly.
// The locking is implemented in the BadgerDB package, but test it nonetheless.
func TestStoreConcurrent(t *testing.T) {
	store, path := createStore(t, encoding.JSON)
	defer os.RemoveAll(path)
	defer store.Close()

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, store)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test empty key
	store, path := createStore(t, encoding.JSON)
	defer os.RemoveAll(path)
	defer store.Close()
	err := store.Set("", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = store.Get("", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = store.Delete("")
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		store, path := createStore(t, encoding.JSON)
		defer os.RemoveAll(path)
		defer store.Close()
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer os.RemoveAll(path)
		defer store.Close()
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			store, path := createStore(t, codec)
			defer os.RemoveAll(path)
			defer store.Close()

			// Prep
			err := store.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = store.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i interface{} // actually nil
			_, err = store.Get("foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = store.Get("foo", valPtr)
			if err == nil {
				t.Error("An error was expected")
			}
		}
	}
	t.Run("get with nil / nil value parameter", createTest(encoding.JSON))
	t.Run("get with nil / nil value parameter", createTest(encoding.Gob))
}

// TestClose tests if the close method returns any errors.
func TestClose(t *testing.T) {
	store, path := createStore(t, encoding.JSON)
	defer os.RemoveAll(path)
	err := store.Close()
	if err != nil {
		t.Error(err)
	}
}

// TestNonExistingDir tests whether the implementation can create the given directory on its own.
// When using BadgerDB directly, it requires the given path to exist and to be writeable.
func TestNonExistingDir(t *testing.T) {
	tmpDir := os.TempDir() + "/BadgerDB"
	err := os.RemoveAll(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	options := badgerdb.Options{
		Dir: tmpDir,
	}
	store, err := badgerdb.NewStore(options)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	// Clean up
	store.Close()
	err = os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Println("Couldn't delete DB directory")
	}
}

func createStore(t *testing.T, codec encoding.Codec) (badgerdb.Store, string) {
	randPath := generateRandomTempDBpath(t)
	options := badgerdb.Options{
		Dir:   randPath,
		Codec: codec,
	}
	store, err := badgerdb.NewStore(options)
	if err != nil {
		t.Fatal(err)
	}
	return store, randPath
}

func generateRandomTempDBpath(t *testing.T) string {
	path, err := ioutil.TempDir(os.TempDir(), "BadgerDB")
	if err != nil {
		t.Fatalf("Generating random DB path failed: %v", err)
	}
	return path
}
