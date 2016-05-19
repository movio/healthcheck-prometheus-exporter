package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	Services  []ConfigService
	Tenants   []Tenant
	Templates map[string]string
}

type Tenant struct {
	Name string
	Host string
}

type ConfigService struct {
	Type string
	Name string
	Vkey string
	Help string
}

func (t Tenant) String() string {
	return fmt.Sprintf("%s: %s", t.Name, t.Host)
}

func readConfig() (Config, error) {
	configFile, _ := filepath.Abs("./config.yml")

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return Config{}, err
	}

	// TODO validate config
	// - Ensure service types matches the available templates

	return config, nil
}
