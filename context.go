package main

import (
	"fmt"
	"io/ioutil"
	"path"

	"gopkg.in/yaml.v2"
)

// Context holds all flags that a user has passed in
type Context struct {
	binaryPath       string
	isDebug          bool
	currentStageSlug string
	apiKey           string
}

type YAMLConfig struct {
	Debug bool `yaml:"debug"`
}

func (c Context) print() {
	fmt.Println("Binary Path =", c.binaryPath)
	fmt.Println("Debug =", c.isDebug)
	fmt.Println("Stage =", c.currentStageSlug)
}

// GetContext parses flags and returns a Context object
func GetContext(env map[string]string) (Context, error) {
	submissionDir, ok := env["CODECRAFTERS_SUBMISSION_DIR"]
	if !ok {
		return Context{}, fmt.Errorf("CODECRAFTERS_SUBMISSION_DIR env var not found")
	}

	currentStageSlug, ok := env["CODECRAFTERS_CURRENT_STAGE_SLUG"]
	if !ok {
		return Context{}, fmt.Errorf("CODECRAFTERS_CURRENT_STAGE_SLUG env var not found")
	}
	configPath := path.Join(submissionDir, "codecrafters.yml")
	binaryPath := path.Join(submissionDir, "spawn_redis_server.sh")

	yamlConfig, err := ReadFromYAML(configPath)
	if err != nil {
		return Context{}, err
	}

	return Context{
		binaryPath:       binaryPath,
		isDebug:          yamlConfig.Debug,
		currentStageSlug: currentStageSlug,
		apiKey:           "dummy",
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
