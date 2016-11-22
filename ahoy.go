package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Config handles the overall configuration in an ahoy.yml file
// with one Config per file.
type Config struct {
	Usage    string
	AhoyAPI  string
	Commands map[string]Command
}

// Command is an ahoy command detailed in ahoy.yml files. Multiple
// commands can be defined per ahoy.yml file.
type Command struct {
	Description string
	Usage       string
	Cmd         string
	Hide        bool
	Imports     []string
}

var app *cli.App
var sourcedir string
var sourcefile string
var args []string
var verbose bool
var bashCompletion bool

var version string

//The build version can be set using the go linker flag `-ldflags "-X main.version=$VERSION"`
//Complete command: `go build -ldflags "-X main.version=$VERSION"`
func logger(errType string, text string) {
	errText := ""
	// Disable the flags which add date and time for instance.
	log.SetFlags(0)

	if (errType == "error") || (errType == "fatal") || (verbose == true) {
		errText = "[" + errType + "] " + text + "\n"
		log.Println(errText)
	}
	if errType == "fatal" {
		os.Exit(1)
	}
}

func getConfigPath(sourcefile string) (string, error) {
	var err error
	var config = ""

	// If a specific source file was set, then try to load it directly.
	if sourcefile != "" {
		if _, err := os.Stat(sourcefile); err == nil {
			return sourcefile, err
		}
		err = errors.New("An ahoy config file was specified using -f to be at " + sourcefile + " but couldn't be found. Check your path.")
		return config, err
	}

	dir, err := os.Getwd()
	if err != nil {
		return config, err
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

func getConfig(sourcefile string) (Config, error) {
	var config = Config{}
	yamlFile, err := ioutil.ReadFile(sourcefile)
	if err != nil {
		err = errors.New("An ahoy config file couldn't be found in your path. You can create an example one by using 'ahoy init'.")
		return config, err
	}

	// Extract the yaml file into the config varaible.
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	// All ahoy files (and imports) must specify the ahoy version.
	// This is so we can support backwards compatability in the future.
	if config.AhoyAPI != "v2" {
		err = errors.New("Ahoy only supports API version 'v2', but '" + config.AhoyAPI + "' given in " + sourcefile)
		return config, err
	}

	return config, err
}

func getSubCommands(includes []string) []cli.Command {
	subCommands := []cli.Command{}
	if 0 == len(includes) {
		return subCommands
	}
	commands := map[string]cli.Command{}
	for _, include := range includes {
		if len(include) == 0 {
			continue
		}
		if include[0] != "/"[0] || include[0] != "~"[0] {
			include = filepath.Join(sourcedir, include)
		}
		if _, err := os.Stat(include); err != nil {
			//Skipping files that cannot be loaded allows us to separate
			//subcommands into public and private.
			continue
		}
		config, _ := getConfig(include)
		includeCommands := getCommands(config)
		for _, command := range includeCommands {
			commands[command.Name] = command
		}
	}

	var names []string
	for k := range commands {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, name := range names {
		subCommands = append(subCommands, commands[name])
	}
	return subCommands
}

func getCommands(config Config) []cli.Command {
	exportCmds := []cli.Command{}

	var keys []string
	for k := range config.Commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		cmd := config.Commands[name]
		cmdName := name

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

		subCommands := getSubCommands(cmd.Imports)
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

	defaultInitCmd := cli.Command{
		Name:  "init",
		Usage: "Initialize a new .ahoy.yml config file in the current directory.",
		Action: func(c *cli.Context) {
			// Grab the URL or use a default for the initial ahoy file.
			// Allows users to define their own files to call to init.
			var wgetURL = "https://raw.githubusercontent.com/ahoy-cli/ahoy/master/examples/examples.ahoy.yml"
			if len(c.Args()) > 0 {
				wgetURL = c.Args()[0]
			}
			grabYaml := "wget " + wgetURL + " -O .ahoy.yml"
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

	// Don't add default commands if they've already been set.
	if c := app.Command(defaultInitCmd.Name); c == nil {
		commands = append(commands, defaultInitCmd)
	}
	return commands
}

//TODO Move these to flag.go?
func init() {
	flag.StringVar(&sourcefile, "f", "", "specify the sourcefile")
	flag.BoolVar(&bashCompletion, "generate-bash-completion", false, "")
	flag.BoolVar(&verbose, "verbose", false, "")
}

// BashComplete prints the list of subcommands as the default app completion method
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

// NoArgsAction is the application wide default action, for when no flags or arguments
// are passed or when a command doesn't exist.
// Looks like -f flag still works through here though.
func NoArgsAction(c *cli.Context) {
	args := c.Args()
	if len(args) > 0 {
		msg := "Command not found for '" + strings.Join(args, " ") + "'"
		logger("fatal", msg)
	}

	cli.ShowAppHelp(c)

	if sourcefile == "" {
		logger("fatal", "No .ahoy.yml found. You can use 'ahoy init' to download an example.")
	}

	if !c.Bool("help") || !c.Bool("version") {
		logger("fatal", "Missing flag or argument.")
	}

	// Looks like we never reach here.
	fmt.Println("ERROR: NoArg Action ")
}

// BeforeCommand runs before every command so arguments or flags must be passed
func BeforeCommand(c *cli.Context) error {
	args := c.Args()
	if c.Bool("version") {
		fmt.Println(version)
		return errors.New("don't continue with commands")
	}
	if c.Bool("help") {
		if len(args) > 0 {
			cli.ShowCommandHelp(c, args.First())
		} else {
			cli.ShowAppHelp(c)
		}
		return errors.New("don't continue with commands")
	}
	//fmt.Printf("%+v\n", args)
	return nil
}

func setupApp(localArgs []string) *cli.App {
	initFlags(localArgs)
	// cli stuff
	app = cli.NewApp()
	app.Action = NoArgsAction
	app.Before = BeforeCommand
	app.Name = "ahoy"
	app.Version = version
	app.Usage = "Creates a configurable cli app for running commands."
	app.EnableBashCompletion = true
	app.BashComplete = BashComplete
	overrideFlags(app)

	if sourcefile, err := getConfigPath(sourcefile); err != nil {
		logger("fatal", err.Error())
	} else {
		sourcedir = filepath.Dir(sourcefile)
		// If we don't have a sourcefile, then just supply the default commands.
		if sourcefile == "" && true {
			app.Commands = addDefaultCommands(app.Commands)
			app.Run(os.Args)
			os.Exit(0)
		}
		config, err := getConfig(sourcefile)
		if err != nil {
			logger("fatal", err.Error())
		}
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

	return app
}

func main() {
	app = setupApp(os.Args[1:])
	app.Run(os.Args)
}
