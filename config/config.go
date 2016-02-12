package config
import (
  "github.com/devinci-code/ahoy/logger"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "log"
  "os"
  "path"
  "path/filepath"
)

type Config struct {
  Usage    string
  AhoyAPI  string
  Version  string
  Commands map[string]Command
}

type Command struct {
  Description string
  Usage       string
  Cmd         string
  Hide        bool
  Import      string
}

func GetConfigPath(sourcefile string) (string, error) {
  var err error

  // If a specific source file was set, then try to load it directly.
  if sourcefile != "" {
    if _, err := os.Stat(sourcefile); err == nil {
      return sourcefile, err
    } else {
      logger.Log("fatal", "An ahoy config file was specified using -f to be at "+sourcefile+" but couldn't be found. Check your path.")
    }
  }

  dir, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
  }
  for dir != "/" && err == nil {
    ymlpath := filepath.Join(dir, ".ahoy.yml")
    //log.Println(ymlpath)
    if _, err := os.Stat(ymlpath); err == nil {
      //log.Println("found: ", ymlpath )
      return ymlpath, err
    }
    // Chop off the last part of the path.
    dir = path.Dir(dir)
  }
  return "", err
}

func GetConfig(sourcefile string) (Config, error) {

  yamlFile, err := ioutil.ReadFile(sourcefile)
  if err != nil {
    logger.Log("fatal", "An ahoy config file couldn't be found in your path. You can create an example one by using 'ahoy init'.")
  }

  var config Config
  // Extract the yaml file into the config varaible.
  err = yaml.Unmarshal(yamlFile, &config)
  if err != nil {
    panic(err)
  }

  // All ahoy files (and imports) must specify the ahoy version.
  // This is so we can support backwards compatability in the future.
  if config.AhoyAPI != "v1" {
    logger.Log("fatal", "Ahoy only supports API version 'v1', but '"+config.AhoyAPI+"' given in "+sourcefile)
  }

  return config, err
}
