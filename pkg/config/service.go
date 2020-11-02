package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/govice/golinks-daemon/pkg/log"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type Service struct {
	token *JWT
}

type JWT struct {
	Token string `json:"token"`
}

func New() (*Service, error) {
	cs := &Service{}
	if err := cs.setupConfig(); err != nil {
		return nil, err
	}

	if err := cs.checkLogin(); err != nil {
		return nil, err
	}

	return cs, nil
}

func (cs *Service) setupConfig() error {
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
	viper.SetDefault("concurrent_task_limit", 3)

	viper.AddConfigPath(daemonHome)

	log.Logln("reading config")

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		configFile := filepath.Join(daemonHome, "config.json")
		if _, err := os.Create(configFile); err != nil {
			return err
		}
		log.Logln("creating new config file")
		if err := viper.WriteConfig(); err != nil {
			log.Logln("failed to write new config file")
			return err
		}
	}

	viper.WatchConfig()
	return nil
}

func (cs *Service) HomeDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".golinksd")
}

var ErrNotAuthorized = errors.New("Not Authorized.")

func (cs *Service) checkLogin() error {
	tokenPath := filepath.Join(cs.HomeDir(), "credentials.json")
	if _, err := os.Stat(tokenPath); errors.Is(err, os.ErrNotExist) {
		email, eok := os.LookupEnv("GOLINKSD_USER")
		password, pok := os.LookupEnv("GOLINKSD_PASSWORD")
		var token *JWT
		//validate environment defined credentials
		if eok && pok {
			token, err = cs.authenticate(email, password)
			if err != nil {
				log.Errln(err)
				return ErrNotAuthorized
			}

		} else {
			token, err = cs.promptLogin()
			if err != nil {
				log.Errln("failed to prompt login:", err)
				return ErrNotAuthorized
			}
		}
		tokenBytes, err := json.Marshal(token)
		if err != nil {
			log.Errln("failed marshal token:", err)
			return err
		}

		if err := ioutil.WriteFile(tokenPath, tokenBytes, os.ModePerm); err != nil {
			log.Errln("failed to write credentials file:", err)
			return err
		}
		cs.token = token
	} else {
		tokenBytes, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			log.Errln("failed to read token file", err)
			return err
		}

		token := &JWT{}
		if err := json.Unmarshal(tokenBytes, token); err != nil {
			log.Errln("failed to unmarshal token", err)
			return err
		}
		cs.token = token
	}
	return nil
}

func (cs *Service) promptLogin() (*JWT, error) {
	promptEmail := promptui.Prompt{
		Label: "Email",
	}
	promptPassword := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
	}

	email, err := promptEmail.Run()
	if err != nil {
		log.Errln("failed to prompt email")
		return nil, err
	}

	password, err := promptPassword.Run()
	if err != nil {
		log.Errln("failed to prompt password")
		return nil, err
	}

	return cs.authenticate(email, password)
}

var ErrFailedAuthentication = errors.New(("failed to authenticate"))

func (cs *Service) authenticate(email, password string) (*JWT, error) {
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
		log.Errln("failed to authorize:", string(respBody))
		return nil, ErrFailedAuthentication
	}

	token := &JWT{}
	if err := json.Unmarshal(respBody, token); err != nil {
		return nil, err
	}

	return token, nil
}

func (cs *Service) Token() string {
	return cs.token.Token
}
