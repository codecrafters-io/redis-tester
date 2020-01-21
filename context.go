package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Context holds all flags that a user has passed in
type Context struct {
	binaryPath        string
	isDebug           bool
	currentStageIndex int
	reportOnSuccess   bool
	apiKey            string
}

type YAMLConfig struct {
	CurrentStage int  `yaml:"current_stage"`
	Debug        bool `yaml:"debug"`
}

func (c Context) print() {
	fmt.Println("Binary Path =", c.binaryPath)
	fmt.Println("Debug =", c.isDebug)
	fmt.Println("Report On Success =", c.reportOnSuccess)
	fmt.Println("Stage =", c.currentStageIndex)
}

// GetContext parses flags and returns a Context object
func GetContext(args []string) (Context, error) {
	flagSet := flag.NewFlagSet("redis-tester", flag.ExitOnError)
	binaryPathPtr := flagSet.String(
		"binary-path",
		"",
		"path to the redis executable to test. Ex: ./run_redis.sh")

	configPathPtr := flagSet.String(
		"config-path",
		"",
		"path to the codecrafters config file. Ex: ./.codecrafters.yml")

	flagSet.Parse(args)

	if *binaryPathPtr == "" {
		return Context{}, fmt.Errorf("" +
			"The --binary-path flag must be specified")
	}

	if *configPathPtr == "" {
		return Context{}, fmt.Errorf("" +
			"The --config-path flag must be specified")
	}

	yamlConfig, err := ReadFromYAML(*configPathPtr)
	if err != nil {
		return Context{}, err
	}

	return Context{
		binaryPath:        *binaryPathPtr,
		isDebug:           yamlConfig.Debug,
		currentStageIndex: yamlConfig.CurrentStage,
		reportOnSuccess:   true,
		apiKey:            "dummy",
	}, nil
}

func ReadFromYAML(configPath string) (YAMLConfig, error) {
	c := &YAMLConfig{}

	fileContents, err := ioutil.ReadFile(configPath)
	if err != nil {
		return YAMLConfig{}, err
	}

	if err := yaml.UnmarshalStrict(fileContents, c); err != nil {
		return YAMLConfig{}, err
	}

	return *c, nil
}
