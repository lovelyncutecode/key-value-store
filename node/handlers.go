package node

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func responseFormatter(handler func(w http.ResponseWriter, r *http.Request) (int, error)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, err := handler(w, r)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), status)
		}
		return
	})
}

func setKey(w http.ResponseWriter, r *http.Request) (int, error) {
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	var body map[string]string
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	keyValStorage.SetRecord(body)
	return 0, nil
}

func getKey(w http.ResponseWriter, r *http.Request) (int, error) {
	q := r.URL.Query()
	key := q.Get("key")
	value, err := keyValStorage.GetRecord(key)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	resp, err := json.Marshal(map[string]string{key: value})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	_, err = w.Write(resp)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func getNewData(w http.ResponseWriter, r *http.Request) (int, error) {
	newRecords, err := keyValStorage.GetNewRecords()
	if err != nil {
		return http.StatusBadRequest, err
	}

	_, err = w.Write(newRecords)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func saveNewData(w http.ResponseWriter, r *http.Request) (int, error) {
	bodyData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	keyValStorage.SetNewRecords(bodyData)
	return 0, nil
}
