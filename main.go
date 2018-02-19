package main

import (
	"log"

	"github.com/dgraph-io/badger"
	"github.com/graphql-go/graphql"
)

func main() {

	opts := badger.DefaultOptions
	opts.Dir = "/tmp/badger"
	opts.ValueDir = "/tmp/badger"
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatalf("Error opening DB: %+v", err)
	}

	cache := map[string]string{
		"Testing_1": "hi mom",
		"Testing_2": "hi dad",
	}

	_ = Persist(db, cache)

	val, err := Retrieve(db, []byte("Testing_1"))
	if err != nil {
		log.Fatalf("Error reading stuff: %+v", err)
	}

	log.Println(val)

	hashConfig := graphql.ArgumentConfig{
		Type:         graphql.String,
		DefaultValue: "",
		Description:  "Hash of the user you want to look up",
	}

	fields := graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"hash": &hashConfig,
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if p.Args["hash"] != nil {
					key := p.Args["hash"]
					value, err := Retrieve(db, []byte(key.(string)))
					if err != nil {
						return nil, err
					}

					return string(value), nil
				}

				return nil, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("Failed to create schema: %+v", err)
	}

	query := `{
        user(hash: "Testing_1")
    }`

	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("Error querying: %+v", r.Errors)
	}

	log.Printf("%+v", r)

	// log.Println(db)

	db.Close()
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
