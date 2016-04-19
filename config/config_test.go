package config

import (
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

var (
	SimpleF    = "../testdata/simple.yml"
	SubConfigF = "../testdata/subconfig.yml"
	MergeF     = "../testdata/merge.yml"
	NoExistF   = "../testdata/noexist.yml"
)

func TestFilePath(t *testing.T) {

	// Non-existing files should give an error
	filename, err := FilePath(NoExistF)
	if assert.NotNil(t, err) {

		//assert.Equal(t, NoExistF, filename)
	}

	// Non-existing files should give an error
	filename, err = FilePath(MergeF)
	if assert.Nil(t, err) {
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

func TestSubConfig(t *testing.T) {

	expectConfig := Config{
		AhoyAPI: "1.0",
		Usage:   "Example Usage",
		Cmd:     "",
		Hide:    false,
		Import:  "",
		Commands: map[string]Config{
			"test": Config{
				AhoyAPI: "",
				Usage:   "",
				Cmd:     "",
				Hide:    false,
				Import:  "",
				Commands: map[string]Config{
					"subtest": Config{
						AhoyAPI:  "",
						Usage:    "Create subcommand",
						Cmd:      `echo "Create subcommand"`,
						Hide:     false,
						Import:   "",
						Commands: nil,
					},
				},
			},
		},
	}

	filename, err := FilePath(SubConfigF)
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

func TestMergeConfig(t *testing.T) {

	expectConfig := Config{
		AhoyAPI: "1.0",
		Usage:   "This is an imported set of commands",
		Cmd:     "",
		Hide:    false,
		Import:  "commands.yml",
		Commands: map[string]Config{
			"test": Config{
				AhoyAPI:  "",
				Usage:    "Override the test command in command.yml",
				Cmd:      `echo "Override"`,
				Hide:     false,
				Import:   "",
				Commands: nil,
			},
			"new": Config{
				AhoyAPI:  "",
				Usage:    "Create a new subcommand",
				Cmd:      `echo "Override"`,
				Hide:     false,
				Import:   "",
				Commands: nil,
			},
		},
	}

	filename, err := FilePath(MergeF)
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

	baseDir := path.Dir(filename)
	config, err = MergeConfig(config, baseDir)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expectConfig, config, "they should be equal")
	//_ = expectConfig
}
