package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {
	url := flag.String("url", "127.0.0.1:8080", "node url in format IP:port")
	mode := flag.String("mode", "get-key", "specify mode: set-key or get-key")
	data := flag.String("data", "key", "specify key to get from node or specify key and its value in format key:value to send to node")
	flag.Parse()
	if url == nil {
		log.Fatal("please, set url flag")
	}
	switch *mode {
	case "set-key":
		err := setKey(*url, *data)
		if err != nil {
			log.Fatal(err)
		}

	case "get-key":
		err := getKey(*url, *data)
		if err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatal("unknown mode: choose either get-key or set-key")
	}
}

func setKey(url, keyVal string) error {
	keyValSl := strings.Split(keyVal, ":")
	if len(keyValSl) < 2 {
		return errors.New("invalid set-key flag format")
	}

	bodyBytes, err := json.Marshal(map[string]string{keyValSl[0]: keyValSl[1]})
	if err != nil {
		log.Fatalln(err)
	}

	req, err := http.NewRequest("POST", "http://"+url+"/set", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("failed to send data to node with status code %s", resp.Status)
	}
	return nil
}

func getKey(url, key string) error {
	req, err := http.NewRequest("GET", "http://"+url+"/get", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Set("key", key)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("failed to get data from node with status code %s", resp.Status)
	}

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	var body map[string]string
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println(body[key])
	return nil
}
