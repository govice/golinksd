package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/govice/golinks/block"
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

func (gs *GolinksService) BearerToken() string {
	return "Bearer " + gs.daemon.configService.token.Token
}

var ErrFailedChainLengthRequest = errors.New("failed to request chain length from remote")

func (gs *GolinksService) GetLength() (int, error) {
	req, err := http.NewRequest("GET", viper.GetString("chain_length_endpoint"), nil)
	if err != nil {
		return -1, err
	}

	req.Header.Add("Authorization", gs.BearerToken())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errln("failed to get length", err)
		return -1, err
	}

	if res.StatusCode != http.StatusOK {
		return -1, ErrFailedChainLengthRequest
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

func (gs *GolinksService) GetBlock(index int) (*block.Block, error) {
	req, err := http.NewRequest("GET", viper.GetString("chain_block_endpoint"), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", gs.BearerToken())

	query := req.URL.Query()
	query.Add("index", strconv.Itoa(index))
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errln("failed to get block")
		return nil, err
	}

	blockBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b := &block.Block{}
	if err := json.Unmarshal(blockBody, b); err != nil {
		return nil, err
	}

	return b, nil
}

var ErrFailedBlockUpload = errors.New("failed to upload block")

func (gs *GolinksService) UploadBlock(blk *block.Block) error {
	blockBytes, err := json.Marshal(blk)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", viper.GetString("chain_block_endpoint"), bytes.NewBuffer(blockBytes))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", gs.BearerToken())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return ErrFailedBlockUpload
	}

	return nil
}
