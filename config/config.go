package config

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type Config struct {
	AhoyAPI     string
	Commands    map[string]Config
	Description string
	Usage       string
	Cmd         string
	Hide        bool
	Import      string
}

var DefaultFilename = ".ahoy.yml"
var RequiredAPIVersion = "0.2"
var Cwd = ""

func init() {
	// Grab the current working directory that ahoy was called from.
	Cwd, _ = os.Getwd()
}

// GetFile returns a valid config path if it exists.
// If sourcefile is set, it checks directly that the file exists.
// Else it searches up from the working directory until it finds it or reaches the root and throws an error.
func FilePath(sourcefile string) (string, error) {

	// If a specific source file was set, then try to load it directly.
	if sourcefile != "" {
		// Use relative paths if an absolute path wasn't specified.
		// If the first character isn't "/" or "~" we assume a relative path.
		if sourcefile[0] != "/"[0] || sourcefile[0] != "~"[0] {
			sourcefile = filepath.Join(Cwd, sourcefile)
		}
		if _, err := os.Stat(sourcefile); err == nil {
			return sourcefile, err
		} else {
			return sourcefile, err
			//logger.Log("fatal", "An ahoy config file was specified using -f or 'import' to be at '"+sourcefile+"' but couldn't be found. Check your path.")
		}
	}

	// Otherwise, start in the current directory that ahoy was called from and work
	// our way up the tree until we either find a .ahoy.yml file or we reach the root.
	return FindFile(Cwd, DefaultFilename)
}

func LoadFile(filename string) ([]byte, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return file, err
}

// FindPath will search for 'filename' starting in 'dir' until it either finds it or reaches the root '/'.
func FindFile(dir string, filename string) (string, error) {
	var err error

	for dir != "/" && err == nil {
		checkFile := filepath.Join(dir, filename)
		// If the file exists at that path, return it
		if _, err := os.Stat(checkFile); err == nil {
			return checkFile, err
		}
		// Othersie, chop off the last part of dir to check one level higher, and repeat.
		dir = path.Dir(dir)
	}
	if dir == "/" {
		return "", errors.New("findpath: No file found.")
	}
	return "", err
}

// Parse a config file and return a simple Config Item.
// Imports are not processed yet.
func ParseConfig(yamlFile []byte) (Config, error) {

	var config Config
	// Extract the yaml file into the config varaible.
	err := yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return Config{}, err
	}

	return config, err
}

func MergeConfig(config Config, baseDir string) (Config, error) {
	// Handle imports.
	if config.Import != "" {
		filename := path.Join(baseDir, config.Import)
		// filename, err := FilePath(filename)
		//if err != nil {
		//return Config{}, err
		//}
		yamlFile, err := LoadFile(filename)
		if err != nil {
			return Config{}, err
		}
		importConfig, err := ParseConfig(yamlFile)
		if err != nil {
			return Config{}, err
		}
		spew.Dump(config)
		spew.Dump(importConfig)
		mergo.Merge(&config, importConfig)
	}
	return config, nil
}

// Checks that version matches reqVersion and outputs an error if it does.
func checkVersion(version string, reqVersion string) error {
	// All ahoy files (and imports) must specify the ahoy version.
	// This is so we can support backwards compatability in the future.
	// TODO: Make this handle ranges of versions
	if version != reqVersion {
		return errors.New("Version is " + version + ", but only version " + reqVersion + " is supported")
	}
	return nil
}
