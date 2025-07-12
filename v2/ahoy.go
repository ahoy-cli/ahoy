package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// Config handles the overall configuration in an ahoy.yml file
// with one Config per file.
type Config struct {
	Usage      string
	AhoyAPI    string
	Commands   map[string]Command
	Entrypoint []string
	Env        string
}

// Command is an ahoy command detailed in ahoy.yml files. Multiple
// commands can be defined per ahoy.yml file.
type Command struct {
	Description string
	Usage       string
	Cmd         string
	Env         string
	Hide        bool
	Optional    bool
	Imports     []string
	Aliases     []string
}

var (
	app            *cli.App
	sourcefile     string
	verbose        bool
	bashCompletion bool
)

// The build version can be set using the go linker flag `-ldflags "-X main.version=$VERSION"`
// Complete command: `go build -ldflags "-X main.version=$VERSION"`
var version string

// AhoyConf stores the global config.
var AhoyConf struct {
	srcDir  string
	srcFile string
}

func logger(errType string, text string) {
	errText := ""
	// Disable the flags which add date and time for instance.
	log.SetFlags(0)
	if errType != "debug" {
		errText = "[" + errType + "] " + text + "\n"
		log.Println(errText)
	}

	if errType == "fatal" {
		os.Exit(1)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getConfigPath(sourcefile string) (string, error) {
	var err error
	config := ""

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

	// Keep track of the previous directory to detect when we've reached the root
	prevDir := ""
	for dir != prevDir && err == nil {
		ymlpath := filepath.Join(dir, ".ahoy.yml")
		// log.Println(ymlpath)
		if _, err := os.Stat(ymlpath); err == nil {
			logger("debug", "Found .ahoy.yml at "+ymlpath)
			return ymlpath, err
		}
		// Chop off the last part of the path.
		prevDir = dir
		dir = filepath.Dir(dir)
	}
	logger("debug", "Can't find an .ahoy.yml file.")
	return "", err
}

func getConfig(file string) (Config, error) {
	config := Config{}
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		err = errors.New("an ahoy config file couldn't be found in your path. You can create an example one by using 'ahoy init'")
		return config, err
	}

	// Extract the yaml file into the config variable.
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	// All ahoy files (and imports) must specify the ahoy version.
	// This is so we can support backwards compatibility in the future.
	if config.AhoyAPI != "v2" {
		err = errors.New("Ahoy only supports API version 'v2', but '" + config.AhoyAPI + "' given in " + sourcefile)
		return config, err
	}

	if config.Entrypoint == nil {
		config.Entrypoint = []string{"bash", "-c", "{{cmd}}", "{{name}}"}
	}

	return config, err
}

func getSubCommands(includes []string) []cli.Command {
	subCommands := []cli.Command{}
	if len(includes) == 0 {
		return subCommands
	}
	commands := map[string]cli.Command{}
	for _, include := range includes {
		if len(include) == 0 {
			continue
		}
		// If the include path is not absolute or a home directory path,
		// prepend the source directory to make it relative to the config file.
		if !strings.HasPrefix(include, "/") && !strings.HasPrefix(include, "~") {
			include = filepath.Join(AhoyConf.srcDir, include)
		}
		if _, err := os.Stat(include); err != nil {
			// Skipping files that cannot be loaded allows us to separate
			// subcommands into public and private.
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

// Given a filepath, return a string array of environment variables.
func getEnvironmentVars(envFile string) []string {
	var envVars []string

	env, err := os.ReadFile(envFile)
	if err != nil {
		logger("fatal", "Invalid env file: "+envFile)
		return nil
	}

	lines := strings.Split(string(env), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignore empty lines and comments (lines starting with '#').
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		envVars = append(envVars, line)
	}
	return envVars
}

func getCommands(config Config) []cli.Command {
	exportCmds := []cli.Command{}
	envVars := []string{}

	// Get environment variables from the 'global' environment variable file, if it is defined.
	if config.Env != "" {
		globalEnvFile := filepath.Join(AhoyConf.srcDir, config.Env)
		envVars = append(envVars, getEnvironmentVars(globalEnvFile)...)
	}

	var keys []string
	for k := range config.Commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		cmd := config.Commands[name]

		// Check that a command has 'cmd' OR 'imports' set.
		if cmd.Cmd == "" && cmd.Imports == nil {
			logger("fatal", "Command ["+name+"] has neither 'cmd' or 'imports' set. Check your yaml file.")
		}

		// Check if a command has 'cmd' AND 'imports' set.
		if cmd.Cmd != "" && cmd.Imports != nil {
			logger("fatal", "Command ["+name+"] has both 'cmd' and 'imports' set, but only one is allowed. Check your yaml file.")
		}

		// Check that a command with 'imports' set has a least one entry.
		if cmd.Imports != nil && len(cmd.Imports) == 0 {
			logger("fatal", "Command ["+name+"] has 'imports' set, but it is empty. Check your yaml file.")
		}

		newCmd := cli.Command{
			Name:            name,
			Aliases:         cmd.Aliases,
			SkipFlagParsing: true,
			HideHelp:        cmd.Hide,
		}

		if cmd.Usage != "" {
			newCmd.Usage = cmd.Usage
		}

		if cmd.Cmd != "" {
			newCmd.Action = func(c *cli.Context) {
				// For some unclear reason, if we don't add an item at the end here,
				// the first argument is skipped... actually it's not!
				// 'bash -c' says that arguments will be passed starting with $0, which also means that
				// $@ skips the first item. See http://stackoverflow.com/questions/41043163/xargs-sh-c-skipping-the-first-argument
				var cmdItems []string
				var cmdArgs []string
				var cmdEntrypoint []string

				// c.Args()  is not a slice apparently.
				for _, arg := range c.Args() {
					if arg != "--" {
						cmdArgs = append(cmdArgs, arg)
					}
				}
				// fmt.Printf("%s : %+v\n", "Args", cmdArgs)

				// Replace the entry point placeholders.
				cmdEntrypoint = config.Entrypoint[:]
				for i := range cmdEntrypoint {
					if cmdEntrypoint[i] == "{{cmd}}" {
						cmdEntrypoint[i] = cmd.Cmd
					} else if cmdEntrypoint[i] == "{{name}}" {
						cmdEntrypoint[i] = c.Command.Name
					}
				}
				cmdItems = append(cmdEntrypoint, cmdArgs...)

				// If defined, included specified command-level environment variables.
				// Note that this will intentionally override any conflicting variables
				// defined in the 'global' env file.
				if cmd.Env != "" {
					cmdEnvFile := filepath.Join(AhoyConf.srcDir, cmd.Env)
					envVars = append(envVars, getEnvironmentVars(cmdEnvFile)...)
				}

				if verbose {
					log.Println("===> AHOY", name, "from", sourcefile, ":", cmdItems)
				}
				command := exec.Command(cmdItems[0], cmdItems[1:]...)
				command.Dir = AhoyConf.srcDir
				command.Stdout = os.Stdout
				command.Stdin = os.Stdin
				command.Stderr = os.Stderr
				command.Env = append(command.Environ(), envVars...)
				if err := command.Run(); err != nil {
					fmt.Fprintln(os.Stderr)
					os.Exit(1)
				}
			}
		}

		if cmd.Imports != nil {
			subCommands := getSubCommands(cmd.Imports)
			if len(subCommands) == 0 {
				if !cmd.Optional {
					logger("fatal", "Command ["+name+"] has 'imports' set, but no commands were found. Check your yaml file.")
				} else {
					continue
				}
			}
			newCmd.Subcommands = subCommands
		}

		// log.Println("Source file:", sourcefile, "- found command:", name, ">", cmd.Cmd)
		exportCmds = append(exportCmds, newCmd)
	}

	return exportCmds
}

func addDefaultCommands(commands []cli.Command) []cli.Command {
	defaultInitCmd := cli.Command{
		Name:  "init",
		Usage: "Initialize a new .ahoy.yml config file in the current directory.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force",
				Usage: "force overwriting the .ahoy.yml file in the current directory.",
			},
		},
		Action: func(c *cli.Context) {
			if fileExists(filepath.Join(".", ".ahoy.yml")) {
				if c.Bool("force") {
					fmt.Println("Warning: '--force' parameter passed, overwriting .ahoy.yml in current directory.")
				} else {
					fmt.Println("Warning: .ahoy.yml found in current directory.")
					fmt.Fprint(os.Stderr, "Are you sure you wish to overwrite it with an example file, y/N ? ")
					reader := bufio.NewReader(os.Stdin)
					char, _, err := reader.ReadRune()
					if err != nil {
						fmt.Println(err)
					}
					// If "y" or "Y", continue and overwrite.
					// Anything else, exit.
					if char != 'y' && char != 'Y' {
						fmt.Println("Abort: exiting without overwriting.")
						os.Exit(0)
					}
					if len(c.Args()) > 0 {
						fmt.Println("Ok, overwriting .ahoy.yml in current directory with specified file.")
					} else {
						fmt.Println("Ok, overwriting .ahoy.yml in current directory with example file.")
					}
				}
			}
			// Grab the URL or use a default for the initial ahoy file.
			// Allows users to define their own files to call to init.
			// TODO: Make file downloading OS-independent.
			wgetURL := "https://raw.githubusercontent.com/ahoy-cli/ahoy/master/examples/examples.ahoy.yml"
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
				if len(c.Args()) > 0 {
					fmt.Println("Your specified .ahoy.yml has been downloaded to the current directory.")
				} else {
					fmt.Println("Example .ahoy.yml downloaded to the current directory. You can customize it to suit your needs!")
				}
			}
		},
	}

	// Don't add default commands if they've already been set.
	if c := app.Command(defaultInitCmd.Name); c == nil {
		commands = append(commands, defaultInitCmd)
	}
	return commands
}

// TODO Move these to flag.go?
func init() {
	logger("debug", "init()")
	flag.StringVar(&sourcefile, "f", "", "specify the sourcefile")
	flag.BoolVar(&bashCompletion, "generate-bash-completion", false, "")
	flag.BoolVar(&verbose, "verbose", false, "")
}

// BashComplete prints the list of subcommands as the default app completion method
func BashComplete(c *cli.Context) {
	logger("debug", "BashComplete()")

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

	if AhoyConf.srcFile == "" {
		logger("error", "No .ahoy.yml found. You can use 'ahoy init' to download an example.")
	}

	if !c.Bool("help") || !c.Bool("version") {
		logger("warn", "Missing flag or argument.")
		os.Exit(1)
	}

	// Exit gracefully if we get to here.
	os.Exit(0)
}

// BeforeCommand runs before every command so arguments or flags must be passed
func BeforeCommand(c *cli.Context) error {
	args := c.Args()
	// fmt.Printf("%+v\n", args)
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
	// fmt.Printf("%+v\n", args)
	return nil
}

func setupApp(localArgs []string) *cli.App {
	var err error
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

	AhoyConf.srcFile, err = getConfigPath(sourcefile)
	if err != nil {
		logger("fatal", err.Error())
	} else {
		AhoyConf.srcDir = filepath.Dir(AhoyConf.srcFile)
		// If we don't have a sourcefile, then just supply the default commands.
		if AhoyConf.srcFile == "" {
			app.Commands = addDefaultCommands(app.Commands)
			app.Run(os.Args)
			os.Exit(0)
		}
		config, err := getConfig(AhoyConf.srcFile)
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
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ if len .Subcommands }}{{" \u25BC"}}{{end}}{{ "\t" }}{{.Usage}} {{if .Aliases}}[ Aliases: {{join .Aliases ", "}} ]{{end}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
ALIASES:
    Commands can have aliases for easier invocation. Aliases are displayed next to each command that has them.
    You can use any of a command's aliases interchangeably with its primary name.
`

	return app
}

func main() {
	logger("debug", "main()")
	app = setupApp(os.Args[1:])
	app.Run(os.Args)
}
