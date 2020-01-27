package main

import (
	"fmt"
	"io/ioutil"
	"path"

	"gopkg.in/yaml.v2"
)

// Context holds all flags that a user has passed in
type Context struct {
	binaryPath        string
	isDebug           bool
	currentStageIndex int
	apiKey            string
}

type YAMLConfig struct {
	CurrentStage int  `yaml:"current_stage"`
	Debug        bool `yaml:"debug"`
}

func (c Context) print() {
	fmt.Println("Binary Path =", c.binaryPath)
	fmt.Println("Debug =", c.isDebug)
	fmt.Println("Stage =", c.currentStageIndex+1)
}

// GetContext parses flags and returns a Context object
func GetContext(env map[string]string) (Context, error) {
	appDir, ok := env["APP_DIR"]
	if !ok {
		return Context{}, fmt.Errorf("APP_DIR env var not found")
	}
	configPath := path.Join(appDir, "codecrafters.yml")
	binaryPath := path.Join(appDir, "spawn_redis_server.sh")

	yamlConfig, err := ReadFromYAML(configPath)
	if err != nil {
		return Context{}, err
	}

	return Context{
		binaryPath:        binaryPath,
		isDebug:           yamlConfig.Debug,
		currentStageIndex: yamlConfig.CurrentStage - 1,
		apiKey:            "dummy",
	}, nil
}

func ReadFromYAML(configPath string) (YAMLConfig, error) {
	c := &YAMLConfig{}

	fileContents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return YAMLConfig{}, err
	}

	if err := yaml.Unmarshal(fileContents, c); err != nil {
		return YAMLConfig{}, err
	}

	return *c, nil
}
