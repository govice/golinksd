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

type ConfigService struct {
	daemon *daemon
	token  *JWT
}

func NewConfigService(daemon *daemon) (*ConfigService, error) {
	cs := &ConfigService{
		daemon: daemon,
	}
	if err := cs.setupConfig(); err != nil {
		return nil, err
	}

	if err := cs.checkLogin(); err != nil {
		return nil, err
	}

	return cs, nil
}

func (cs *ConfigService) setupConfig() error {
	daemonHome := cs.HomeDir()

	os.Mkdir(daemonHome, os.ModePerm)

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.SetEnvPrefix("golinksd")
	viper.AutomaticEnv()
	viper.SetDefault("peer_port", 7777)
	viper.SetDefault("auth_server", "https://govice.org")
	viper.SetDefault("port", 8080)
	viper.SetDefault("genesis", false)
	viper.SetDefault("delay_startup", 0)
	viper.SetDefault("templates_home", "./templates")
	viper.SetDefault("tracking_period", 30000)

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

	viper.WatchConfig()
	return nil
}

var ErrInvalidLedger = errors.New("failed to load ledger from config")

func (cs *ConfigService) HomeDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".golinksd")
}

var ErrNotAuthorized = errors.New("Not Authorized.")

func (cs *ConfigService) checkLogin() error {
	tokenPath := filepath.Join(cs.HomeDir(), "credentials.json")
	if _, err := os.Stat(tokenPath); errors.Is(err, os.ErrNotExist) {
		email, eok := os.LookupEnv("GOLINKSD_USER")
		password, pok := os.LookupEnv("GOLINKSD_PASSWORD")
		var token *JWT
		//validate environment defined credentials
		if eok && pok {
			token, err = cs.authenticate(email, password)
			if err != nil {
				errln(err)
				return ErrNotAuthorized
			}

		} else {
			token, err = cs.promptLogin()
			if err != nil {
				errln("failed to prompt login:", err)
				return ErrNotAuthorized
			}
		}
		tokenBytes, err := json.Marshal(token)
		if err != nil {
			errln("failed marshal token:", err)
			return err
		}

		if err := ioutil.WriteFile(tokenPath, tokenBytes, os.ModePerm); err != nil {
			errln("failed to write credentials file:", err)
			return err
		}
		cs.token = token
	} else {
		tokenBytes, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			errln("failed to read token file", err)
			return err
		}

		token := &JWT{}
		if err := json.Unmarshal(tokenBytes, token); err != nil {
			errln("failed to unmarshal token", err)
			return err
		}
		cs.token = token
	}
	return nil
}

func (cs *ConfigService) promptLogin() (*JWT, error) {
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

	return cs.authenticate(email, password)
}

var ErrFailedAuthentication = errors.New(("failed to authenticate"))

func (cs *ConfigService) authenticate(email, password string) (*JWT, error) {
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
