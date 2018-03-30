package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const CONFIG_VERSION = "0.1"

type Config struct {
	BaseURI    string
	MetaAlias  string
	SPEntityID string
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux struct {
		BaseURI    string `yaml:"BaseURI"`
		MetaAlias  string `yaml:"MetaAlias"`
		SPEntityID string `yaml:"SPEntityID"`
	}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	c.BaseURI = aux.BaseURI
	c.MetaAlias = aux.MetaAlias
	c.SPEntityID = aux.SPEntityID

	return nil
}

func (c *Config) MarshalYAML() (interface{}, error) {
	return map[string]interface{}{
		"BaseURI":       c.BaseURI,
		"MetaAlias":     c.MetaAlias,
		"SPEntityID":    c.SPEntityID,
		"ConfigVersion": CONFIG_VERSION,
	}, nil
}

func Load() (*Config, error) {
	path, err := configFile()
	if err != nil {
		return nil, err
	}

	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(d, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func Save(c *Config) error {
	path, err := configFile()
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, b, 0644); err != nil {
		return err
	}

	return nil
}
