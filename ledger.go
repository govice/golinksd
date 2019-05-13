package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Info struct {
	Message string `json:"message"`
}

type Ledger struct {
	Nodes []Node `json:"nodes"`
}

type Node struct {
	Address   string `json:"address"`
	Available bool
}

func (ledger Ledger) PingNodes() {
	for _, node := range ledger.Nodes {
		if !node.Available {
			log.Println(node.Address + ": " + "Node was unavailable. Skipping")
			continue
		}

		if strings.Contains(node.Address, os.Getenv("PORT")) {
			continue
		}

		resp, err := http.Get(node.Address + "/ping")
		if err != nil {
			log.Printf("Failed to ping %v", node.Address)
			node.Available = false
			return
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read response body")
			return
		}

		var message Info
		if err := json.Unmarshal(bodyBytes, &message); err != nil {
			log.Println("Failed to marshal info message")
			return
		}

		log.Println(node.Address + ": " + message.Message)
	}
}
