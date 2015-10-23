package main

import (
  "os"
  "github.com/codegangsta/cli"
  "fmt"
  "os/exec"
)

func main() {
  app := cli.NewApp()
  app.Name = "greet"
  app.Usage = "fight the loneliness!"
  app.Action = func(c *cli.Context) {
    cmd := exec.Command(os.Args[1], os.Args[2:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
      fmt.Fprintln(os.Stderr, err)
      os.Exit(1)
    }
  }

  app.Run(os.Args)
}
