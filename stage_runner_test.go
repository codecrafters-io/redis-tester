package main

import (
	"io/ioutil"

	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v2"
)

type StageYAML struct {
	Slug string `yaml:"slug"`
}

type CourseYAML struct {
	Stages []StageYAML `yaml:"stages"`
}

func TestStagesMatchYAML(t *testing.T) {
	bytes, err := ioutil.ReadFile("test_helpers/course_definition.yml")
	if err != nil {
		t.Fatal(err)
	}
	c := CourseYAML{}
	if err := yaml.Unmarshal(bytes, &c); err != nil {
		t.Fatal(err)
	}

	definitionStages := []string{}
	for _, stage := range c.Stages {
		definitionStages = append(definitionStages, stage.Slug)
	}

	runnerStages := []string{}
	for _, stage := range newStageRunner(true).stages {
		runnerStages = append(runnerStages, stage.slug)
	}

	assert.Equal(t, definitionStages, runnerStages)
}
