package daemon

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/govice/golinksd/pkg/log"
	"github.com/govice/golinksd/pkg/worker"
)

type WorkerConfigManager struct {
	Path string
}

func (w *WorkerConfigManager) ReadConfig() (*worker.Config, error) {
	_, err := os.Stat(w.Path)
	if os.IsNotExist(err) {
		log.Logln("no worker configuration defined, initializing with empty config")
		return &worker.Config{}, nil
	}
	configBytes, err := ioutil.ReadFile(w.Path)
	if err != nil {
		return nil, err
	}

	workerConfig := &worker.Config{}
	if err := json.Unmarshal(configBytes, workerConfig); err != nil {
		return nil, err
	}

	return workerConfig, nil
}

func (w *WorkerConfigManager) WriteConfig(cfg *worker.Config) error {
	log.Logln("saving worker config...")

	configBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(w.Path, configBytes, 0666)
}
