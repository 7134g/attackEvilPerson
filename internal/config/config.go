package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BaiduURL       string   `yaml:"baidu_url"`
	TelNumber      string   `yaml:"tel_number"`
	TelName        string   `yaml:"tel_name"`
	Proxy          string   `yaml:"proxy"`
	BrowserPath    string   `yaml:"browser_path"`
	Titles         []string `yaml:"titles"`
	Relatives      []string `yaml:"relatives"`
	Situations     []string `yaml:"situations"`
	ContactMethods []string `yaml:"contact_methods"`
	Greetings      []string `yaml:"greetings"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.TelNumber == "" {
		return nil, fmt.Errorf("tel_number 不能为空")
	}
	if cfg.TelName == "" {
		return nil, fmt.Errorf("tel_name 不能为空")
	}
	return cfg, nil
}
