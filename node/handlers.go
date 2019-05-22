package node

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func setKey(w http.ResponseWriter, r *http.Request) {
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body map[string]string
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	keyValStorage.SaveRecord(body)
}

func getKey(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key := q.Get("key")
	value, err := keyValStorage.GetRecord(key)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]string{key: value})
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getNewData(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lastUpdateTime, err := strconv.Atoi(q.Get("last_update_time"))
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newRecords, err := keyValStorage.GetNewRecords(int64(lastUpdateTime))
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(newRecords)
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func saveNewData(w http.ResponseWriter, r *http.Request) {
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body map[string]StorageRecord
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		log.Println(errors.WithStack(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	keyValStorage.SetNewRecords(body)
}
