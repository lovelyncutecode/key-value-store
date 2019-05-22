package main

import (
	"encoding/json"
	"flag"
	"github.com/lovelyncutecode/key-value-store/node"
	"log"
	"os"
)

func main() {
	configFile := flag.String("config", "./config.json", "")
	flag.Parse()
	f, err := os.Open(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	jsonParser := json.NewDecoder(f)
	config := &node.KeyValueStorageConfig{}
	err = jsonParser.Decode(config)
	if err != nil {
		log.Fatal(err)
	}

	kvs, err := node.NewKeyValueStorage(config)
	if err != nil {
		log.Fatal(err)
	}

	kvs.Run()
}
