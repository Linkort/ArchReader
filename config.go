package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	DefsCom  string `yaml:"defs_com"`
	DefsBaud string `yaml:"defs_baud"`
	DefsPlc  uint16 `yaml:"defs_plc"`
	Archives []struct {
		Type uint8  `yaml:"type"`
		Name string `yaml:"name"`
		Data []struct {
			Mode int    `yaml:"mode"`
			Text string `yaml:"text"`
		} `yaml:"data"`
	} `yaml:"Archives"`
}

// Чтение файла настроек и архивов
func getConfigYAML(filename string) (config, error) {
	conf := new(config)

	file, err := os.Open(filename)
	if err != nil {
		return *conf, err
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(conf)
	if err != nil {
		return *conf, err
	}

	return *conf, nil
}
