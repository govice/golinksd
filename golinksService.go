package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/spf13/viper"
)

type GolinksService struct {
	daemon *daemon
}

func NewGolinksService(daemon *daemon) (*GolinksService, error) {
	return &GolinksService{
		daemon: daemon,
	}, nil
}

func (gs *GolinksService) GetLength() (int, error) {

	req, err := http.NewRequest("GET", viper.GetString("chain_length_endpoint"), nil)
	if err != nil {
		return -1, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	payload := &struct {
		Length int `json:"length"`
	}{}

	if err := json.Unmarshal(bodyBytes, payload); err != nil {
		return -1, err
	}

	return payload.Length, nil
}
