package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/devinci-code/ahoy/config"
	"github.com/devinci-code/ahoy/flag"
	"github.com/devinci-code/ahoy/logger"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var app *cli.App
var args []string

func getCommands(config config.Config) []cli.Command {
	exportCmds := []cli.Command{}

	var keys []string
	for k := range config.Commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		cmd := config.Commands[name]
		cmdName := name
		subCommands := []cli.Command{}

		// Handle the import of subcommands.
		if cmd.Import != "" {
			// If the first character isn't "/" or "~" we assume a relative path.
			subSource := cmd.Import
			if cmd.Import[0] != "/"[0] || cmd.Import[0] != "~"[0] {
				subSource = filepath.Join(sourcedir, cmd.Import)
			}
			logger.Log("info", "Importing commands into '"+name+"' command from "+subSource)
			subConfig, _ := getConfig(subSource)
			subCommands = getCommands(subConfig)
		}

		newCmd := cli.Command{
			Name:            name,
			SkipFlagParsing: true,
			HideHelp:        cmd.Hide,
		}

		if cmd.Usage != "" {
			newCmd.Usage = cmd.Usage
		}

		if cmd.Cmd != "" {
			newCmd.Action = func(c *cli.Context) {
				args = c.Args()
				runCommand(cmdName, cmd.Cmd)
			}
		}

		if subCommands != nil {
			newCmd.Subcommands = subCommands
		}

		//log.Println("found command: ", name, " > ", cmd.Cmd )
		exportCmds = append(exportCmds, newCmd)
	}

	return exportCmds
}

func runCommand(name string, c string) {

	cReplace := strings.Replace(c, "{{args}}", strings.Join(args, " "), -1)

	dir := sourcedir

	if verbose {
		log.Println("===> AHOY", name, "from", sourcefile, ":", cReplace)
	}
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

func addDefaultCommands(commands []cli.Command) []cli.Command {
	newCmd := cli.Command{
		Name:  "init",
		Usage: "Initialize a new .ahoy.yml config file in the current directory.",
		Action: func(c *cli.Context) {
			// Grab the URL or use a default for the initial ahoy file.
			// Allows users to define their own files to call to init.
			var wgetUrl = "https://raw.githubusercontent.com/devinci-code/ahoy/master/examples/examples.ahoy.yml"
			if len(c.Args()) > 0 {
				wgetUrl = c.Args()[0]
			}
			grabYaml := "wget " + wgetUrl + " -O .ahoy.yml"
			cmd := exec.Command("bash", "-c", grabYaml)
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr)
				os.Exit(1)
			} else {
				fmt.Println("example.ahoy.yml downloaded to the current directory. You can customize it to suit your needs!")
			}
		},
	}

	// TODO: Check if a command has already been set. Don't add defaults if it has.
	commands = append(commands, newCmd)
	return commands
}

// Prints the list of subcommands as the default app completion method
func BashComplete(c *cli.Context) {

	if sourcefile != "" {
		log.Println(sourcefile)
		os.Exit(0)
	}
	for _, command := range c.App.Commands {
		for _, name := range command.Names() {
			fmt.Fprintln(c.App.Writer, name)
		}
	}
}

func main() {
	flag.InitFlags()
	//log.Println(sourcefile)
	// cli stuff
	app = cli.NewApp()
	app.Name = "ahoy"
	app.Usage = "Creates a configurable cli app for running commands."
	app.EnableBashCompletion = true
	app.BashComplete = BashComplete
	flag.OverrideFlags(app)
	if sourcefile, err := config.GetConfigPath(sourcefile); err == nil {
		sourcedir = filepath.Dir(sourcefile)
		config, _ := config.GetConfig(sourcefile)
		app.Commands = getCommands(config)
		app.Commands = addDefaultCommands(app.Commands)
		if config.Usage != "" {
			app.Usage = config.Usage
		}
	}

	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .Flags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
   {{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t" }}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`

	app.Run(os.Args)
}
