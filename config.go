package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func setupConfig() error {
	daemonHome := HomeDir()

	os.Mkdir(daemonHome, os.ModePerm)

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.SetEnvPrefix("golinksd")
	viper.SetDefault("peer_port", 7777)
	viper.SetDefault("auth_server", "https://govice.org")
	viper.SetDefault("port", 8080)
	viper.SetDefault("genesis", false)
	viper.SetDefault("delay_startup", 0)
	viper.SetDefault("templates_home", "./templates")

	viper.AddConfigPath(daemonHome)

	logln("reading config")
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		configFile := filepath.Join(daemonHome, "config.json")
		if _, err := os.Create(configFile); err != nil {
			return err
		}
		logln("creating new config file")
		if err := viper.WriteConfig(); err != nil {
			logln("failed to write new config file")
			return err
		}
	}
	return nil
}

var ErrInvalidLedger = errors.New("failed to load ledger from config")

func HomeDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".golinksd")
}

func checkLogin() error {
	tokenPath := filepath.Join(HomeDir(), "credentials.json")

	if _, err := os.Stat(tokenPath); errors.Is(err, os.ErrNotExist) {
		token, err := promptLogin()
		if err != nil {
			errln("failed to prompt login:", err)
			return err
		}

		tokenBytes, err := json.Marshal(token)
		if err != nil {
			errln("failed marshal token:", err)
			return err
		}

		if err := ioutil.WriteFile(tokenPath, tokenBytes, os.ModePerm); err != nil {
			errln("failed to write credentials file:", err)
		}
	}
	return nil
}

func promptLogin() (*JWT, error) {
	promptEmail := promptui.Prompt{
		Label: "Email",
	}
	promptPassword := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
	}

	email, err := promptEmail.Run()
	if err != nil {
		errln("failed to prompt email")
		return nil, err
	}

	password, err := promptPassword.Run()
	if err != nil {
		errln("failed to prompt password")
		return nil, err
	}

	return authenticate(email, password)
}

var ErrFailedAuthentication = errors.New(("failed to authenticate"))

func authenticate(email, password string) (*JWT, error) {
	loginPayload, err := json.Marshal(gin.H{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", viper.GetString("authorization_endpoint"), bytes.NewBuffer(loginPayload))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errln("failed to authorize:", string(respBody))
		return nil, ErrFailedAuthentication
	}

	token := &JWT{}
	if err := json.Unmarshal(respBody, token); err != nil {
		return nil, err
	}

	return token, nil

}

type JWT struct {
	Token string `json:"token"`
}
