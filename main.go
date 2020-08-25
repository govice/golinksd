package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/manifoldco/promptui"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"

	"github.com/kardianos/service"
)

var daemonLogger service.Logger

type Config struct {
	RendezvousString string
	BootstrapPeers   []maddr.Multiaddr
	ListenAddresses  []maddr.Multiaddr
	ProtocolID       string
}

func main() {
	if err := setupConfig(); err != nil {
		fatalln(err)
	}

	if err := checkLogin(); err != nil {
		fatalln(err)
	}

	l, err := loadLedger()
	if err != nil {
		fatalln(err)
	}
	ledger = *l
	logln(ledger)

	log.Println("PORT: " + viper.GetString("port"))
	log.Println("AUTH_SERVER: " + viper.GetString("auth_server"))

	serviceConfig := &service.Config{
		Name:        "golinksd",
		DisplayName: "golinksd",
		Description: "golinks daemon",
	}

	d := &daemon{}
	s, err := service.New(d, serviceConfig)
	if err != nil {
		fatalln(err)
	}

	d.service = s

	//TODO refactor to use system logger
	daemonLogger, err = s.Logger(nil)
	if err != nil {
		fatalln(err)
	}

	if err = s.Run(); err != nil {
		daemonLogger.Error(err)
	}
}

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
	viper.SetDefault("ledger", "https://govice.org/assets/ledger.json")

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

func loadLedger() (*Ledger, error) {
	ledger := &Ledger{}

	ledgerBytes, err := ioutil.ReadFile(filepath.Join(HomeDir(), "ledger.json"))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(ledgerBytes, ledger); err != nil {
		return nil, err
	}

	return ledger, nil
}

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
