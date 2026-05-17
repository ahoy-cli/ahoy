package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Config handles the overall configuration in an ahoy.yml file
// with one Config per file.
type Config struct {
	Usage      string
	AhoyAPI    string
	Commands   map[string]Command
	Entrypoint []string
	Env        StringArray
}

// Command is an ahoy command detailed in ahoy.yml files. Multiple
// commands can be defined per ahoy.yml file.
type Command struct {
	Description string
	Usage       string
	Cmd         string
	Env         StringArray
	Hide        bool
	Optional    bool
	Imports     []string
	Aliases     []string
}

var (
	rootCmd         *cobra.Command
	sourcefile      string
	verbose         bool
	simulateVersion string
	ahoyExecutable  string
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
	// Disable the flags which add date and time for instance.
	log.SetFlags(0)
	if errType == "debug" {
		if verbose {
			log.Println("[debug] " + text)
		}
		return
	}
	log.Println("[" + errType + "] " + text)
	if errType == "fatal" {
		os.Exit(1)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// expandPath expands a file path, handling tilde expansion and relative paths.
// For absolute paths, returns the path as-is.
// For tilde paths (starting with ~), expands to the user home directory.
// For relative paths, joins with the provided base directory.
func expandPath(path, baseDir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	// On Windows, filepath.IsAbs returns false for Unix-style paths like "/foo"
	// (which require a drive letter to be considered absolute). Treat them as
	// absolute here so cross-platform config files behave consistently.
	if strings.HasPrefix(path, "/") {
		return path
	}
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			remainder := path[1:]
			if len(remainder) > 0 && remainder[0] == '/' {
				remainder = remainder[1:]
			}
			return filepath.Join(home, remainder)
		}
		return path
	}
	return filepath.Join(baseDir, path)
}

func getConfigPath(sourcefile string) (string, error) {
	var err error
	config := ""

	// If a specific source file was set, then try to load it directly.
	if sourcefile != "" {
		if _, statErr := os.Stat(sourcefile); statErr == nil {
			return sourcefile, nil
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
	for dir != prevDir {
		ymlpath := filepath.Join(dir, ".ahoy.yml")
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

func getSubCommands(includes []string) []*cobra.Command {
	subCommands := []*cobra.Command{}
	if len(includes) == 0 {
		return subCommands
	}
	commands := map[string]*cobra.Command{}
	for _, include := range includes {
		if len(include) == 0 {
			continue
		}
		include = expandPath(include, AhoyConf.srcDir)
		if _, err := os.Stat(include); err != nil {
			if !os.IsNotExist(err) {
				// File exists but is unreadable (e.g. EACCES) - log so the
				// user knows why commands are missing.
				logger("error", "Cannot access import file '"+include+"': "+err.Error())
			}
			// Skipping missing or unreadable files allows subcommands to be
			// separated into public and private sets.
			continue
		}
		config, err := getConfig(include)
		if err != nil {
			logger("error", "Could not load imported config '"+include+"': "+err.Error())
			continue
		}
		includeCommands := getCommands(config)
		for _, command := range includeCommands {
			commands[command.Name()] = command
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

	// We allow non-existent "env" files, so skip if file doesn't exist.
	if !fileExists(envFile) {
		return nil
	}

	env, err := os.ReadFile(envFile)
	if err != nil {
		// The file was confirmed to exist above, so this is a real read
		// failure (e.g. EACCES, EIO) - not a routine missing-file case.
		logger("error", "Failed to read environment file '"+envFile+"': "+err.Error())
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

func getCommands(config Config) []*cobra.Command {
	exportCmds := []*cobra.Command{}
	envVars := []string{}

	// Get environment variables from all 'global' environment variable files, if any are defined.
	if len(config.Env) > 0 {
		for _, envPath := range config.Env {
			globalEnvFile := expandPath(envPath, AhoyConf.srcDir)
			vars := getEnvironmentVars(globalEnvFile)
			if vars != nil {
				envVars = append(envVars, vars...)
			}
		}
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

		// Check that a command with 'imports' set has at least one entry.
		if cmd.Imports != nil && len(cmd.Imports) == 0 {
			logger("fatal", "Command ["+name+"] has 'imports' set, but it is empty. Check your yaml file.")
		}

		newCmd := &cobra.Command{
			Use:     name,
			Aliases: cmd.Aliases,
			// Don't use DisableFlagParsing - it prevents persistent flags from being parsed
			// Instead, we'll use FParseErrWhitelist to allow unknown flags to pass through
			FParseErrWhitelist: cobra.FParseErrWhitelist{
				UnknownFlags: true,
			},
			Hidden: cmd.Hide,
		}

		if cmd.Usage != "" {
			newCmd.Short = cmd.Usage
		}

		if cmd.Description != "" {
			newCmd.Long = cmd.Description
		}

		if cmd.Cmd != "" {
			// Capture variables for the closure
			cmdString := cmd.Cmd
			cmdEnv := cmd.Env
			cmdName := name

			newCmd.Run = func(cobraCmd *cobra.Command, args []string) {
				// 'bash -c' passes arguments starting with $0, so $@ skips the first item.
				// See http://stackoverflow.com/questions/41043163/xargs-sh-c-skipping-the-first-argument
				var cmdItems []string
				var cmdArgs []string
				var cmdEntrypoint []string

				// Filter out "--" separator
				for _, arg := range args {
					if arg != "--" {
						cmdArgs = append(cmdArgs, arg)
					}
				}

				// Replace the entry point placeholders.
				cmdEntrypoint = config.Entrypoint[:]
				for i := range cmdEntrypoint {
					if cmdEntrypoint[i] == "{{cmd}}" {
						cmdEntrypoint[i] = cmdString
					} else if cmdEntrypoint[i] == "{{name}}" {
						cmdEntrypoint[i] = cmdName
					}
				}
				cmdItems = append(cmdEntrypoint, cmdArgs...)

				// Collect environment variables
				cmdEnvVars := append([]string{}, envVars...)

				// If defined, include any command-level environment variables.
				// Note that this will intentionally override any conflicting variables
				// defined in the 'global' env file.
				if len(cmdEnv) > 0 {
					for _, envPath := range cmdEnv {
						cmdEnvFile := expandPath(envPath, AhoyConf.srcDir)
						vars := getEnvironmentVars(cmdEnvFile)
						if vars != nil {
							cmdEnvVars = append(cmdEnvVars, vars...)
						}
					}
				}

				// Inject ahoy-specific environment variables so subprocesses can
				// identify the running binary and the invoked command name.
				ahoyEnvVars := []string{"AHOY_COMMAND_NAME=" + cmdName}
				if ahoyExecutable != "" {
					ahoyEnvVars = append(ahoyEnvVars, "AHOY_CMD="+ahoyExecutable)
				}
				cmdEnvVars = append(ahoyEnvVars, cmdEnvVars...)

				if verbose {
					log.Println("===> Ahoy", cmdName, "from", sourcefile, ":", cmdItems)
				}
				command := exec.Command(cmdItems[0], cmdItems[1:]...)
				command.Dir = AhoyConf.srcDir
				command.Stdout = os.Stdout
				command.Stdin = os.Stdin
				command.Stderr = os.Stderr
				command.Env = append(command.Environ(), cmdEnvVars...)
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
					errorMsg := fmt.Sprintf("Command [%s] has 'imports' set, but no commands were found.", name)

					// List any import files that are missing to help diagnose the issue.
					var missingFiles []string
					for _, importPath := range cmd.Imports {
						fullPath := expandPath(importPath, AhoyConf.srcDir)
						if !fileExists(fullPath) {
							missingFiles = append(missingFiles, importPath)
						}
					}

					if len(missingFiles) > 0 {
						errorMsg += fmt.Sprintf("\n\nMissing import files: %s", strings.Join(missingFiles, ", "))
						errorMsg += "\n\nSolutions:"
						errorMsg += "\n1. Create the missing files"
						errorMsg += "\n2. Mark imports as optional with 'optional: true'"
						if !VersionSupports(GetAhoyVersion(), "optional_imports") {
							errorMsg += fmt.Sprintf("\n3. Upgrade Ahoy to v%s+ for optional import support", FeatureSupport["optional_imports"])
						}
						errorMsg += "\n\nFor more help, run: ahoy config validate"
					}

					logger("fatal", errorMsg)
				} else {
					if !VersionSupports(GetAhoyVersion(), "optional_imports") {
						errorMsg := fmt.Sprintf("Command [%s] uses 'optional: true' but this Ahoy version (%s) doesn't support optional imports.", name, GetAhoyVersion())
						errorMsg += fmt.Sprintf("\n\nThis feature requires Ahoy %s or later.", FeatureSupport["optional_imports"])
						errorMsg += "\n\nSolutions:"
						errorMsg += "\n1. Upgrade Ahoy to the latest version"
						errorMsg += "\n2. Remove 'optional: true' and create the missing import files"
						errorMsg += "\n\nFor more help, run: ahoy config validate"
						logger("fatal", errorMsg)
					}
					continue
				}
			}
			newCmd.AddCommand(subCommands...)
		}

		// Set per-command help template to show the full description.
		newCmd.SetHelpFunc(commandHelpFunc)

		exportCmds = append(exportCmds, newCmd)
	}

	return exportCmds
}

func addDefaultCommands(commands []*cobra.Command) []*cobra.Command {
	// 'ahoy config' command group with 'validate' and 'init' subcommands.
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Ahoy configuration.",
	}
	configCmd.SetHelpFunc(commandHelpFunc)

	configValidateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate and diagnose an Ahoy configuration file.",
		Run:   validateCommandAction,
	}

	configInitCmd := &cobra.Command{
		Use:   "init [url]",
		Short: "Initialise a new .ahoy.yml config file in the current directory.",
		Run:   initCommandAction,
	}
	configInitCmd.Flags().Bool("force", false, "force overwriting the .ahoy.yml file in the current directory.")

	configCmd.AddCommand(configValidateCmd, configInitCmd)

	// 'ahoy init' kept for backwards compatibility with a deprecation notice.
	deprecatedInitCmd := &cobra.Command{
		Use:   "init [url]",
		Short: "Initialise a new .ahoy.yml config file in the current directory.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(os.Stderr, "Note: 'ahoy init' is deprecated. Please use 'ahoy config init' instead.")
			initCommandAction(cmd, args)
		},
	}
	deprecatedInitCmd.Flags().Bool("force", false, "force overwriting the .ahoy.yml file in the current directory.")

	// Don't add default commands if they've already been set.
	hasConfig := false
	hasInit := false
	for _, cmd := range commands {
		switch cmd.Name() {
		case "config":
			hasConfig = true
		case "init":
			hasInit = true
		}
	}
	if !hasConfig {
		commands = append(commands, configCmd)
	}
	if !hasInit {
		commands = append(commands, deprecatedInitCmd)
	}
	return commands
}

// BashComplete prints the list of subcommands as the default app completion method
func BashComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	logger("debug", "BashComplete()")

	if sourcefile != "" {
		log.Println(sourcefile)
		os.Exit(0)
	}

	completions := []string{}
	for _, command := range cmd.Root().Commands() {
		completions = append(completions, command.Name())
		completions = append(completions, command.Aliases...)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// NoArgsAction is the application wide default action, for when no flags or arguments
// are passed or when a command doesn't exist.
func NoArgsAction(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		msg := "Command not found for '" + strings.Join(args, " ") + "'"
		logger("fatal", msg)
	}

	cmd.Help()

	if AhoyConf.srcFile == "" {
		logger("error", "No .ahoy.yml found. You can use 'ahoy init' to download an example.")
	}

	helpRequested, _ := cmd.Flags().GetBool("help")
	versionRequested, _ := cmd.Flags().GetBool("version")
	if !helpRequested && !versionRequested {
		logger("warn", "Missing flag or argument.")
		os.Exit(1)
	}

	// Exit gracefully if we get to here.
	os.Exit(0)
}

// BeforeCommand is a PersistentPreRunE hook that handles --version and --help
// flag processing before cobra executes each command.
func BeforeCommand(cmd *cobra.Command, args []string) error {
	// Check if version was set via --version (double dash) by cobra.
	versionRequested, _ := cmd.Flags().GetBool("version")
	if versionRequested {
		if version != "" {
			fmt.Println(version)
		}
		os.Exit(0)
	}

	// Check if help was set via --help (double dash) by cobra.
	helpRequested, _ := cmd.Flags().GetBool("help")
	if helpRequested {
		if len(args) > 0 {
			// Find the subcommand and show its help.
			for _, subcmd := range cmd.Commands() {
				if subcmd.Name() == args[0] {
					subcmd.Help()
					os.Exit(0)
				}
			}
		}
		cmd.Help()
		os.Exit(0)
	}
	return nil
}

func setupApp(localArgs []string) *cobra.Command {
	var err error

	initFlags(localArgs)

	// initFlags() pre-parsed sourcefile and verbose from the legacy
	// single-dash forms (-f, -verbose, etc.) - see flag.go for the full
	// rationale. The cobra flag definitions below would re-bind those
	// same variables and reset them to their zero values, so we capture
	// the parsed values now and pass them as the cobra flag defaults.
	parsedSourcefile := sourcefile
	parsedVerbose := verbose

	// Create root command
	rootCmd = &cobra.Command{
		Use:     "ahoy",
		Version: version,
		Short:   "Creates a configurable cli app for running commands.",
		RunE: func(cmd *cobra.Command, args []string) error {
			NoArgsAction(cmd, args)
			return nil
		},
		PersistentPreRunE: BeforeCommand,
		ValidArgsFunction: BashComplete,
	}

	// Set up global flags with the parsed values as defaults
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", parsedVerbose, "Output extra details like the commands to be run.")
	rootCmd.PersistentFlags().StringVarP(&sourcefile, "file", "f", parsedSourcefile, "Use a specific ahoy file.")
	rootCmd.PersistentFlags().Bool("help", false, "show help")
	rootCmd.PersistentFlags().Bool("version", false, "print the version")
	rootCmd.PersistentFlags().Bool("generate-bash-completion", false, "")

	// Add hidden --simulate-version flag for testing the validation system
	// against older Ahoy versions without needing to rebuild the binary.
	rootCmd.PersistentFlags().StringVar(&simulateVersion, "simulate-version", "", "simulate a specific Ahoy version for testing")

	// Mark help, version, and internal flags as hidden since we handle them manually.
	rootCmd.PersistentFlags().MarkHidden("help")
	rootCmd.PersistentFlags().MarkHidden("version")
	rootCmd.PersistentFlags().MarkHidden("generate-bash-completion")
	rootCmd.PersistentFlags().MarkHidden("simulate-version")

	// Disable default help command
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	AhoyConf.srcFile, err = getConfigPath(sourcefile)
	if err != nil {
		logger("fatal", err.Error())
	} else {
		AhoyConf.srcDir = filepath.Dir(AhoyConf.srcFile)
		// If we don't have a sourcefile, then just supply the default commands.
		if AhoyConf.srcFile == "" {
			commands := addDefaultCommands([]*cobra.Command{})
			rootCmd.AddCommand(commands...)
			rootCmd.Execute()
			os.Exit(0)
		}
		config, err := getConfig(AhoyConf.srcFile)
		if err != nil {
			logger("fatal", err.Error())
		}
		commands := getCommands(config)
		commands = addDefaultCommands(commands)
		rootCmd.AddCommand(commands...)
		if config.Usage != "" {
			rootCmd.Short = config.Usage
		}
	}

	// Set up custom help template
	rootCmd.SetHelpFunc(customHelpFunc)

	// Suppress cobra's built-in error/usage prints. main() inspects the
	// error returned by Execute() and prints ahoy's own friendlier
	// equivalents (e.g. "Command not found for ...").
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	return rootCmd
}

// commandHelpFunc provides per-command help output with a DESCRIPTION section.
func commandHelpFunc(cmd *cobra.Command, args []string) {
	funcMap := template.FuncMap{
		"join":      strings.Join,
		"trimSpace": strings.TrimSpace,
	}

	helpTemplate := `NAME:
   {{.Name}} - {{.Short}}{{if .Long}}

DESCRIPTION:

{{trimSpace .Long}}
{{end}}
USAGE:
   {{.UseLine}} [arguments...]
{{if .HasAvailableSubCommands}}
COMMANDS:{{range .Commands}}{{if not .Hidden}}
   {{.Name}}{{if .Aliases}}, {{join .Aliases ", "}}{{end}}	{{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
FLAGS:
{{.LocalFlags.FlagUsages}}{{end}}{{if .Aliases}}
ALIASES:
   {{join .Aliases ", "}}
{{end}}
`

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 8, 2, ' ', 0)
	t := template.Must(template.New("commandHelp").Funcs(funcMap).Parse(helpTemplate))
	err := t.Execute(w, cmd)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error rendering help: %s\n", err)
		if os.Getenv("CLI_TEMPLATE_ERROR_DEBUG") != "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "CLI TEMPLATE ERROR: %#v\n", err)
		}
		return
	}
	w.Flush()
}

// customHelpFunc provides custom help output with aliases support.
func customHelpFunc(cmd *cobra.Command, args []string) {
	funcMap := template.FuncMap{
		"join":      strings.Join,
		"replace":   strings.ReplaceAll,
		"trimSpace": strings.TrimSpace,
	}

	helpTemplate := `NAME:
   {{.Use}} - {{.Short}}

USAGE:
   {{.UseLine}}{{if .HasAvailableSubCommands}} command [command options]{{end}} [arguments...]
{{if .HasAvailableSubCommands}}
COMMANDS:{{range .Commands}}{{if not .Hidden}}
   {{.Name}}{{if .Aliases}}, {{join .Aliases ", "}}{{end}}{{if .HasSubCommands}} ▼{{end}}	{{.Short}}
{{end}}{{end}}
Use 'ahoy <command> --help' for detailed information about a command.
Run 'ahoy config validate' to check your configuration for issues.
{{end}}{{if .HasAvailableLocalFlags}}
GLOBAL OPTIONS:
{{.LocalFlags.FlagUsages}}{{end}}{{if .Version}}
VERSION:
   {{.Version}}{{end}}
`

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 8, 2, ' ', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Parse(helpTemplate))
	err := t.Execute(w, cmd)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error rendering help: %s\n", err)
		if os.Getenv("CLI_TEMPLATE_ERROR_DEBUG") != "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "CLI TEMPLATE ERROR: %#v\n", err)
		}
		return
	}
	w.Flush()
}

func main() {
	logger("debug", "main()")
	if exe, err := os.Executable(); err == nil {
		ahoyExecutable = exe
	}
	rootCmd = setupApp(os.Args[1:])

	// Check for invalid flag error from initFlags - show help and exit 1.
	if invalidFlagError != "" {
		fmt.Print(invalidFlagError)
		rootCmd.Help()
		os.Exit(1)
	}

	// Check for -version and -help flags set during initFlags (single-dash versions)
	// This handles single-dash versions that cobra doesn't support
	if versionFlagSet {
		if version != "" {
			fmt.Println(version)
		}
		os.Exit(0)
	}

	if helpFlagSet {
		rootCmd.Help()
		os.Exit(0)
	}

	// Handle bash completion flag - print completions and exit
	if bashCompletionFlagSet {
		for _, command := range rootCmd.Commands() {
			if !command.Hidden {
				fmt.Println(command.Name())
			}
		}
		os.Exit(0)
	}

	// Route stderr through a pipe drained by a goroutine so subprocesses
	// writing more than the pipe buffer (~64 KB) to stderr don't deadlock.
	// Output is teed to the real stderr in real time, preserving live
	// pass-through for child processes (the primary use case for ahoy).
	// If pipe creation fails, fall back to running with stderr untouched.
	oldStderr := os.Stderr
	r, w, pipeErr := os.Pipe()

	var err error

	if pipeErr != nil {
		err = rootCmd.Execute()
	} else {
		os.Stderr = w

		drained := make(chan struct{})
		go func() {
			defer close(drained)
			io.Copy(oldStderr, r)
		}()

		err = rootCmd.Execute()

		// Closing the writer signals EOF to the drain goroutine. Wait for
		// it to finish so any in-flight stderr is flushed before we exit.
		w.Close()
		<-drained
		os.Stderr = oldStderr
	}

	if err != nil {
		// Cobra has SilenceErrors=true so the error has not been printed.
		// Translate "unknown command" into ahoy's friendly equivalent.
		if strings.Contains(err.Error(), "unknown command") {
			// Format: "unknown command \"something\" for \"ahoy\""
			parts := strings.Split(err.Error(), "\"")
			if len(parts) >= 2 {
				cmdName := parts[1]
				msg := "Command not found for '" + cmdName + "'"
				logger("fatal", msg)
			}
		}
		os.Exit(1)
	}
}
