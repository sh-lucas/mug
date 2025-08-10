package global

import (
	_ "embed"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var defaultConfigName = "mug.yml"

//go:embed defaults.yml
var defaultFile []byte

type Service string

const (
	Mongo  Service = "mongo"
	Rabbit Service = "rabbit"
)

type config struct {
	Debug    bool `yaml:"debug"`
	Features struct {
		Watch      bool `yaml:"watch"`
		Inj_envs   bool `yaml:"inj_envs"`
		Gen_router bool `yaml:"gen_router"`
		Gen_envs   bool `yaml:"gen_envs"`
	} `yaml:"features"`
	Services []Service `yaml:"services"`
}

var Config = config{}

func init() {
	cfgFile, err := os.ReadFile(defaultConfigName)
	// file not found -> use defaults
	if err != nil {
		setConfig(defaultFile)
		return
	}

	// empty file -> fill defaults
	if len(cfgFile) == 0 {
		_ = os.WriteFile(defaultConfigName, defaultFile, 0644)
		cfgFile = defaultFile
	}
	setConfig(cfgFile)
}

func setConfig(file []byte) {
	err := yaml.Unmarshal(file, &Config)
	if err != nil {
		log.Fatalf("Could not load config file %s: %v", defaultConfigName, err)
	}
}
