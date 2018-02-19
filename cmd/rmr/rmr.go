package main

import (
	"log"
	"os"

	"github.com/alfreddobradi/rmr/internal/cache"
	"github.com/huin/goupnp"

	"github.com/graphql-go/graphql"
)

func main() {

	x, _ := goupnp.DiscoverDevices("ssdp:all")

	log.Printf("%+v", x)

	os.Exit(1)

	db, err := cache.New()
	if err != nil {
		log.Fatalf("Error getting Badger instance: %+v", err)
	}

	cacheData := map[string]string{
		"Testing_1": "hi mom",
		"Testing_2": "hi dad",
	}

	_ = cache.Persist(db, cacheData)

	val, err := cache.Retrieve(db, []byte("Testing_1"))
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
					value, err := cache.Retrieve(db, []byte(key.(string)))
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
