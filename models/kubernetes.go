package models

type Kubernetes struct {
	Namespaces []Namespace `json:"namespaces"`
	ConfigFile ConfigPath  `json:"config_file"`
}
