package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	SimpleF  = "../testdata/simple.yml"
	NoExistF = "../testdata/noexist.yml"
)

func TestFilePath(t *testing.T) {

	// Non-existing files should give an error
	filename, err := FilePath(NoExistF)
	if assert.NotNil(t, err) {

		//assert.Equal(t, NoExistF, filename)
	}

	// Existing relative files should not give an error
	filename, err = FilePath(SimpleF)
	if assert.Nil(t, err) {
		//assert.Equal(t, SimpleF, filename)
	}

	_ = filename
}

func TestFileLoad(t *testing.T) {
	expectYaml := []byte(
		"ahoyapi: 1.0\n" +
			"usage: Example Usage\n")

	// Non-existing files should give an error
	yamlFile, err := LoadFile(NoExistF)
	if assert.NotNil(t, err) {
		assert.Equal(t, []byte(nil), yamlFile)
	}

	// Existing relative files should not give an error
	yamlFile, err = LoadFile(SimpleF)
	if assert.Nil(t, err) {
		assert.Equal(t, expectYaml, yamlFile)
	}

}

func TestParseSimpleConfig(t *testing.T) {

	expectConfig := Config{
		AhoyAPI:  "1.0",
		Usage:    "Example Usage",
		Cmd:      "",
		Hide:     false,
		Import:   "",
		Commands: nil,
		//Commands: make(map[string]Config),
	}

	filename, err := FilePath(SimpleF)
	if !assert.Nil(t, err) {
		return
	}

	yamlFile, err := LoadFile(filename)
	if !assert.Nil(t, err) {
		return
	}

	config, err := ParseConfig(yamlFile)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expectConfig, config, "they should be equal")

}
