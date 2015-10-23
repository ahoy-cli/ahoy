package main

import (
  "os"
  "github.com/codegangsta/cli"
  "fmt"
  "os/exec"
  "log"
  "path"
)

func getComposeDir() string {
  dir, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
  }
  for dir != "" && err == nil {
    dir, file := path.Split(dir)
    cut_off_last_char_len := len(dir) - 1
    dir = dir[:cut_off_last_char_len]
    fmt.Println(dir)
    _ = file
  }
  // Go complains with an error if you don't use a variable.
  if err != nil {
    log.Fatal(err)
  }
  return dir
}

func main() {
  app := cli.NewApp()
  app.Name = "ahoy"
  app.Usage = "Send commands to docker-compose services"
  app.Action = func(c *cli.Context) {

    fmt.Println(getComposeDir())
    cmd := exec.Command(os.Args[1], os.Args[2:]...)
    cmd.Stdout = os.Stdout
    cmd.Stdin = os.Stdin
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
      fmt.Fprintln(os.Stderr, err)
      os.Exit(1)
    }
  }

  app.Run(os.Args)
}
