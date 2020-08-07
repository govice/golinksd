package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/spf13/viper"
)

type AuthenticationService struct {
}

var authService *AuthenticationService

func (service *AuthenticationService) valid(userAuth *externalUserAuth) (bool, error) {
	authServerURI := viper.GetString("auth_server")
	authJSON, err := json.Marshal(userAuth)
	if err != nil {
		return false, err
	}
	var buffer bytes.Buffer
	buffer.Write(authJSON)
	res, err := http.Post(authServerURI, "application/json", &buffer)
	if err != nil {
		return false, err
	}

	if res.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil
}

type externalUserAuth struct {
	Token string `json:"token"`
	Email string `json:"email"`
}
