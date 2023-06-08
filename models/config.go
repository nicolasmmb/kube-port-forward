package models

import (
	"encoding/json"
	"os"

	"github.com/nicolasmmb/kube-port-forward/utils/check"
)

type BaseConfig struct {
	Kubernetes Kubernetes `json:"kubernetes"`
}

func (b *BaseConfig) LoadConfig() {
	actualDir := os.Getenv("PWD")
	configFile := actualDir + "/config.json"

	_, err := os.Stat(configFile)
	check.Error(err, true, true)

	file, err := os.Open(configFile)
	check.Error(err, true, true)
	defer file.Close()

	err = json.NewDecoder(file).Decode(&b)
	check.Error(err, true, true)

}

func (b *BaseConfig) PrintConfig() {
	val, err := json.Marshal(b)
	check.Error(err, true, true)
	val = append(val, '\n')
	os.Stdout.Write(val)
}
