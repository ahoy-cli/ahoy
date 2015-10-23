package main

import (
  "os"
  "github.com/codegangsta/cli"
  "fmt"
  "os/exec"
  "log"
  "path"
  "path/filepath"
)

func getComposeDir() (string, error) {
  var err error
  dir, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
  }
  for dir != "/" && err == nil {
    dir = path.Dir(dir)
    ymlpath := filepath.Join(dir, ".ahoy.yml")
    fmt.Println(ymlpath)
    if _, err := os.Stat(ymlpath); err == nil {
      fmt.Printf("found: %s", ymlpath )
      return dir, err
    }
  }
  return "", err
}

func main() {
  app := cli.NewApp()
  app.Name = "ahoy"
  app.Usage = "Send commands to docker-compose services"
  app.Action = func(c *cli.Context) {

    if ymlpath, err := getComposeDir(); err == nil {
      fmt.Println(ymlpath)
    }
    cmd := exec.Command(os.Args[1], os.Args[2:]...)
    cmd.Stdout = os.Stdout
    cmd.Stdin = os.Stdin
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
      fmt.Fprintln(os.Stderr)
      os.Exit(1)
    }
  }

  app.Run(os.Args)
}
