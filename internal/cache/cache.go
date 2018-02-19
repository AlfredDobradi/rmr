package cache

import "github.com/dgraph-io/badger"

// New creates a Badger instance
func New() (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = "/tmp/badger"
	opts.ValueDir = "/tmp/badger"
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return db, err
}

// Retrieve gets an element from Badger
func Retrieve(conn *badger.DB, key []byte) ([]byte, error) {
	var value []byte

	err := conn.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(key))
		if err != nil {
			return err
		}

		val, err := item.Value()
		if err != nil {
			return err
		}

		value = val

		return nil
	})

	return value, err
}

// Persist writes data to Badger
func Persist(conn *badger.DB, data map[string]string) error {
	err := conn.Update(func(tx *badger.Txn) error {
		for k, v := range data {
			err := tx.Set([]byte(k), []byte(v))
			if err != nil {
				return err
			}
		}

		_ = tx.Commit(nil)
		return nil
	})
	return err
}
