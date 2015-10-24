package main

import (
  "os"
  "github.com/codegangsta/cli"
  "fmt"
  "os/exec"
  "log"
  "path"
  "path/filepath"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "strings"
)

type Config struct {
  Version string
  Commands map[string]Command
}

type Command struct {
  Description string
  Usage string
  Cmd string
}

var sourcedir string
var args []string

func getConfigPath() (string, error) {
  var err error
  dir, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
  }
  for dir != "/" && err == nil {
    ymlpath := filepath.Join(dir, ".ahoy.yml")
    log.Println(ymlpath)
    if _, err := os.Stat(ymlpath); err == nil {
      log.Println("found: ", ymlpath )
      return ymlpath, err
    }
    // Chop off the last part of the path.
    dir = path.Dir(dir)
  }
  return "", err
}

func getConfig(sourcefile string) (Config, error) {

  yamlFile, err := ioutil.ReadFile(sourcefile)
  if err != nil {
    panic(err)
  }

  var config Config

  err = yaml.Unmarshal(yamlFile, &config)
  if err != nil {
    panic(err)
  }
  return config, err
}

func getCommands(config Config) []cli.Command {
  exportCmds := []cli.Command{}
  for name, cmd := range config.Commands {
    cmdCopy := cmd.Cmd
    newCmd := cli.Command{
      Name: name,
      Usage: cmd.Usage,
      Action: func(c *cli.Context) {
       args = c.Args()
       runCommand(cmdCopy);
      },
    }
    log.Println("found command: ", name, " > ", cmd.Cmd )
    exportCmds = append(exportCmds, newCmd)
  }

  return exportCmds
}

func runCommand(c string) {
  //fmt.Printf("%+v\n", exportCmd)

  cReplace := strings.Replace(c, "{{args}}", strings.Join(args, " "), 1)

  dir := sourcedir
  log.Println("args: ", args)
  log.Println("run command: ", cReplace)
  cmd := exec.Command("bash", "-c", cReplace)
  cmd.Dir = dir
  cmd.Stdout = os.Stdout
  cmd.Stdin = os.Stdin
  cmd.Stderr = os.Stderr
  if err := cmd.Run(); err != nil {
    fmt.Fprintln(os.Stderr)
    os.Exit(1)
  }
}

func main() {
  // cli stuff
  app := cli.NewApp()
  app.Name = "ahoy"
  app.Usage = "Send commands to docker-compose services"
  app.EnableBashCompletion = true
  if sourcefile, err := getConfigPath(); err == nil {
    sourcedir = filepath.Dir(sourcefile)
    config, _ := getConfig(sourcefile)
    app.Commands = getCommands(config)
    log.Println("version: ", config.Version)
  }

  app.Run(os.Args)
}
