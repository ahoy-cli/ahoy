package main

import (
  "os"
  "github.com/codegangsta/cli"
  "fmt"
  "os/exec"
  "log"
  "path"
  "path/filepath"
  "github.com/smallfish/simpleyaml"
  "io/ioutil"
  "strings"
)

var sourcedir string

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

func getConfig(sourcefile string) (*simpleyaml.Yaml, error) {

  source, err := ioutil.ReadFile(sourcefile)
  if err != nil {
    panic(err)
  }
  yaml, err := simpleyaml.NewYaml(source)
  if err != nil {
    panic(err)
  }
  return yaml, err
}

func getCommands(y *simpleyaml.Yaml) []cli.Command {
  yamlCmds := y.Get("commands")
  exportCmds := []cli.Command{}
  m, _ := yamlCmds.Map()
  for key, value := range m {
    newCmd := cli.Command{
      Name: key.(string),
      Action: func(c *cli.Context) {
       runCommand(value.(string));
      },
    }
    log.Println("found command: ", key, " > ", value )
    exportCmds = append(exportCmds, newCmd)
  }

  return exportCmds
}

func runCommand(c string) {
  //fmt.Printf("%+v\n", exportCmd)
  dir := sourcedir
  args := strings.Split(c, " ")
  //cmd := exec.Command(os.Args[1], os.Args[2:]...)
  log.Println("run command: ", args[0] )
  cmd := exec.Command(args[0], args[1:]...)
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
  app := cli.NewApp()
  app.Name = "ahoy"
  app.Usage = "Send commands to docker-compose services"
  app.EnableBashCompletion = true
  if sourcefile, err := getConfigPath(); err == nil {
    sourcedir = filepath.Dir(sourcefile)
    yml, _ := getConfig(sourcefile)
    app.Commands = getCommands(yml)
    version, _ := yml.Get("version").String()
    log.Println("version: ", version)
  }

  app.Run(os.Args)
}
