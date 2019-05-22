package node

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

type (
	StorageRecord struct {
		Value string
		// updated field is used for conflict resolution
		Updated int64
	}

	KeyValueStorage struct {
		*sync.Mutex
		storage        map[string]StorageRecord
		runningSourceNode    *http.Client
		runningSourceNodeURL string
		server         *http.Server
		config         *KeyValueStorageConfig
	}

	KeyValueStorageConfig struct {
		Host                      string  `json:"host"`
		Port                      int     `json:"port"`
		SourceNodeHost            *string `json:"node_host,omitempty"`
		SourceNodePort            *int    `json:"node_port,omitempty"`
		RunningNodeRequestTimeout int     `json:"request_timeout"`
	}
)

var keyValStorage *KeyValueStorage

func NewKeyValueStorage(config *KeyValueStorageConfig) (*KeyValueStorage, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/set", responseFormatter(setKey))
	mux.HandleFunc("/get", responseFormatter(getKey))
	mux.HandleFunc("/internal/set", responseFormatter(saveNewData))
	mux.HandleFunc("/internal/get", responseFormatter(getNewData))
	srv := &http.Server{
		Addr:    config.Host + ":" + strconv.Itoa(config.Port),
		Handler: mux,
	}

	keyValStorage = &KeyValueStorage{
		Mutex:          new(sync.Mutex),
		storage:        make(map[string]StorageRecord),
		server:         srv,
		config:         config,
	}
	if config.SourceNodeHost == nil && config.SourceNodePort == nil {
		return keyValStorage, nil
	}

	nodeClient := new(http.Client)
	keyValStorage.runningSourceNode = nodeClient
	keyValStorage.runningSourceNodeURL = "http://" + *config.SourceNodeHost + ":" + strconv.Itoa(*config.SourceNodePort)
	return keyValStorage, nil
}

func (kvs *KeyValueStorage) Run() {
	var goroutinesNum int
	if kvs.runningSourceNode == nil {
		goroutinesNum = 1
	} else {
		goroutinesNum = 2
	}

	stop := make(chan bool, goroutinesNum)
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go kvs.runNode(stop, wg)

	if kvs.runningSourceNode != nil {
		wg.Add(1)
		go kvs.runClient(stop, wg)
	}

	<-sig
	for i := 0; i < goroutinesNum; i++ {
		stop <- true
	}
	wg.Wait()
}

func (kvs *KeyValueStorage) runNode(stop chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	go func() {
		<-stop
		err := kvs.server.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("started server on " + kvs.config.Host + ":" + strconv.Itoa(kvs.config.Port))
	if err := kvs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func (kvs *KeyValueStorage) SetRecord(records map[string]string) {
	keyValStorage.Lock()
	defer keyValStorage.Unlock()

	for k, v := range records {
		keyValStorage.storage[k] = StorageRecord{
			Value:   v,
			Updated: time.Now().Unix(),
		}
	}
}

func (kvs *KeyValueStorage) GetRecord(key string) (string, error) {
	keyValStorage.Lock()
	defer keyValStorage.Unlock()

	value, ok := keyValStorage.storage[key]
	if !ok {
		return "", errors.Errorf("key '%s' not found", key)
	}

	return value.Value, nil
}

func (kvs *KeyValueStorage) runClient(stop chan bool, wg *sync.WaitGroup) {
	tickerC := time.NewTicker(time.Duration(kvs.config.RunningNodeRequestTimeout) * time.Second).C
	for {
		select {
		case <-stop:
			wg.Done()
			return

		case <-tickerC:
			err := kvs.exchangeNewData()
			if err != nil {
				wg.Done()
				log.Fatal(err)
			}
		}
	}
}

func (kvs *KeyValueStorage) exchangeNewData() error {
	err := kvs.retrieveNewData()
	if err != nil {
		return err
	}

	err = kvs.sendNewData()
	return err
}

func (kvs *KeyValueStorage) retrieveNewData() error {
	resp, err := kvs.runningSourceNode.Get(kvs.runningSourceNodeURL+"/internal/get")
	if err != nil {
		return errors.WithStack(err)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("failed to get data from node (%s) with status code %s", kvs.runningSourceNodeURL, resp.Status)
	}

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	var body map[string]StorageRecord
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		return errors.WithStack(err)
	}

	kvs.SetNewRecords(body)
	return nil
}

func (kvs *KeyValueStorage) sendNewData() error {
	newRecords, err := kvs.GetNewRecords()
	if err != nil {
		return err
	}

	resp, err := kvs.runningSourceNode.Post(kvs.runningSourceNodeURL+"/internal/set", "application/json", bytes.NewReader(newRecords))
	if err != nil {
		return errors.WithStack(err)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("new data was not accepted by node (%s) with status code %s", kvs.runningSourceNodeURL, resp.Status)
	}

	return nil
}

func (kvs *KeyValueStorage) GetNewRecords() ([]byte, error) {
	kvs.Lock()
	defer kvs.Unlock()

	records, err := json.Marshal(kvs.storage)
	return records, errors.WithStack(err)
}

func (kvs *KeyValueStorage) SetNewRecords(records map[string]StorageRecord) {
	kvs.Lock()
	defer kvs.Unlock()

	for k, newVal := range records {
		if oldVal, ok := kvs.storage[k]; ok && oldVal.Updated > newVal.Updated {
			continue
		}
		kvs.storage[k] = newVal
	}
}
