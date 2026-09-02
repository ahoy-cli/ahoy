package main

import (
	"errors"
	"fmt"
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

// errNoConfig is returned by getConfigPath when no .ahoy.yml file can be
// located in the current directory tree and no -f flag was given.
var errNoConfig = errors.New("no .ahoy.yml config file found")

// Config handles the overall configuration in an ahoy.yml file
// with one Config per file.
type Config struct {
	Usage      string             `yaml:"usage"`
	AhoyAPI    string             `yaml:"ahoyapi"`
	Commands   map[string]Command `yaml:"commands"`
	Entrypoint []string           `yaml:"entrypoint"`
	Env        EnvFiles           `yaml:"env"`
}

// Command is an ahoy command detailed in ahoy.yml files. Multiple
// commands can be defined per ahoy.yml file.
type Command struct {
	Description string   `yaml:"description"`
	Usage       string   `yaml:"usage"`
	Cmd         string   `yaml:"cmd"`
	Env         EnvFiles `yaml:"env"`
	Hide        bool     `yaml:"hide"`
	Optional    bool     `yaml:"optional"`
	Imports     []string `yaml:"imports"`
	Aliases     []string `yaml:"aliases"`
}

// Build metadata variables injected at link time via -ldflags "-X main.version=...".
var (
	version   string
	GitCommit string
	GitBranch string
	BuildTime string
)

// simulateVersion is a test-only package-level var set by the hidden
// --simulate-version flag. It overrides the reported Ahoy version for
// exercising the validation system without rebuilding the binary.
// Never set in production use.
var simulateVersion string

// appState holds all mutable runtime state for an ahoy invocation,
// replacing the package-level globals that made concurrent testing unsafe.
type appState struct {
	sourcefile     string
	verbose        bool
	ahoyExecutable string
	commandArgs    []string
	importVisited  map[string]bool
	srcDir         string
	srcFile        string
	// flag pre-parse results written by initFlags, read by setupApp/main.
	invalidFlagError      string
	versionFlagSet        bool
	helpFlagSet           bool
	bashCompletionFlagSet bool
}

func newAppState() *appState {
	return &appState{}
}

// logLineBreaks escapes the characters that would otherwise let a logged value
// start a new log line. CRLF is collapsed to a single escape so Windows-authored
// input does not read as two breaks.
var logLineBreaks = strings.NewReplacer("\r\n", `\n`, "\r", `\r`, "\n", `\n`)

// sanitiseLogValue neutralises carriage returns and line feeds in a single
// config-derived value, so it cannot forge additional log lines. The characters
// are escaped rather than dropped so the original content stays visible.
//
// This is applied to interpolated values rather than inside logger(), because
// several messages deliberately span multiple lines to format help text, and
// sanitising the assembled message would mangle that formatting.
func sanitiseLogValue(value string) string {
	return logLineBreaks.Replace(value)
}

func (s *appState) logger(errType string, text string) {
	log.SetFlags(0)
	if errType == logLevelDebug {
		if s.verbose {
			log.Println("[debug] " + text)
		}
		return
	}
	log.Println("[" + errType + "] " + text)
	if errType == logLevelFatal {
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

// normalizePath normalizes a file path to its absolute and clean form.
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	cleaned := filepath.Clean(path)
	if abs, err := filepath.Abs(cleaned); err == nil {
		return abs
	}
	return cleaned
}

func (s *appState) getConfigPath() (string, error) {
	if s.sourcefile != "" {
		if _, statErr := os.Stat(s.sourcefile); statErr == nil {
			return s.sourcefile, nil
		}
		return "", errors.New("An ahoy config file was specified using -f to be at " + s.sourcefile + " but couldn't be found. Check your path.")
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	prevDir := ""
	for dir != prevDir {
		ymlpath := filepath.Join(dir, ".ahoy.yml")
		if _, err := os.Stat(ymlpath); err == nil {
			s.logger(logLevelDebug, "Found .ahoy.yml at "+sanitiseLogValue(ymlpath))
			return ymlpath, nil
		}
		prevDir = dir
		dir = filepath.Dir(dir)
	}
	s.logger(logLevelDebug, "Can't find an .ahoy.yml file.")
	return "", errNoConfig
}

func getConfig(file string) (Config, error) {
	config := Config{}
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		err = errors.New("an ahoy config file couldn't be found in your path. You can create an example one by using 'ahoy init'")
		return config, err
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	if config.AhoyAPI != "v2" {
		err = errors.New("Ahoy only supports API version 'v2', but '" + config.AhoyAPI + "' given in " + file)
		return config, err
	}

	if config.Entrypoint == nil {
		config.Entrypoint = []string{"bash", "-c", "{{cmd}}", "{{name}}"}
	}

	return config, err
}

func (s *appState) processImport(include string, commands map[string]*cobra.Command) {
	include = expandPath(include, s.srcDir)
	normalizedInclude := normalizePath(include)

	// Guard against circular imports. Lazily initialise so direct callers
	// in tests don't need to prime the map themselves.
	if s.importVisited == nil {
		s.importVisited = map[string]bool{}
	}
	if s.importVisited[normalizedInclude] {
		s.logger(logLevelWarn, "Circular import detected for '"+sanitiseLogValue(include)+"', skipping.")
		return
	}
	s.importVisited[normalizedInclude] = true
	defer func() {
		delete(s.importVisited, normalizedInclude)
	}()

	if _, err := os.Stat(include); err != nil {
		if !os.IsNotExist(err) {
			// File exists but is unreadable (e.g. EACCES) - log so the
			// user knows why commands are missing.
			s.logger(logLevelError, "Cannot access import file '"+sanitiseLogValue(include)+"': "+sanitiseLogValue(err.Error()))
		}
		// Skipping missing or unreadable files allows subcommands to be
		// separated into public and private sets.
		return
	}
	config, err := getConfig(include)
	if err != nil {
		s.logger(logLevelError, "Could not load imported config '"+sanitiseLogValue(include)+"': "+sanitiseLogValue(err.Error()))
		return
	}
	includeCommands := s.getCommands(config)
	for _, command := range includeCommands {
		commands[command.Name()] = command
	}
}

func (s *appState) getSubCommands(includes []string) []*cobra.Command {
	subCommands := []*cobra.Command{}
	if len(includes) == 0 {
		return subCommands
	}
	commands := map[string]*cobra.Command{}
	for _, include := range includes {
		if len(include) == 0 {
			continue
		}
		s.processImport(include, commands)
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

// getEnvironmentVars returns the KEY=VALUE lines of an env file, resolving the
// path relative to baseDir.
//
// A missing file is not fatal, which is long-standing behaviour that plenty of
// configurations rely on for layered overrides. It is however reported unless
// the entry was marked optional: a command that silently runs with empty
// credentials because its env file was never created is a worse outcome than
// a line on stderr (issue #188).
func (s *appState) getEnvironmentVars(envFile EnvFile, baseDir string) []string {
	var envVars []string

	resolvedPath := expandPath(envFile.Path, baseDir)

	if !fileExists(resolvedPath) {
		if !envFile.Optional {
			s.logger(logLevelWarn, "environment file '"+sanitiseLogValue(envFile.Path)+"' not found, continuing without it. Add 'optional: true' to that entry if it is meant to be absent.")
		}
		return nil
	}

	env, err := os.ReadFile(resolvedPath)
	if err != nil {
		// The file was confirmed to exist above, so this is a real read
		// failure (e.g. EACCES, EIO) - not a routine missing-file case.
		s.logger(logLevelError, "Failed to read environment file '"+sanitiseLogValue(resolvedPath)+"': "+sanitiseLogValue(err.Error()))
		return nil
	}

	lines := strings.Split(string(env), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignore empty lines and comments (lines starting with '#').
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Warn on lines that don't contain '=' - common culprit is shell
		// `export KEY=VALUE` syntax, which is not supported here.
		if !strings.Contains(line, "=") {
			s.logger(logLevelWarn, "ignoring malformed line in env file '"+sanitiseLogValue(resolvedPath)+"' (expected KEY=VALUE, got: "+sanitiseLogValue(line)+")")
			continue
		}
		envVars = append(envVars, line)
	}
	return envVars
}

func (s *appState) getCommands(config Config) []*cobra.Command {
	exportCmds := []*cobra.Command{}
	envVars := []string{}

	// Get environment variables from all 'global' environment variable files, if any are defined.
	if len(config.Env) > 0 {
		for _, envFile := range config.Env {
			vars := s.getEnvironmentVars(envFile, s.srcDir)
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

	var configErrors []string
	for _, name := range keys {
		cmd := config.Commands[name]

		// Check that a command has 'cmd' OR 'imports' set.
		if cmd.Cmd == "" && cmd.Imports == nil {
			configErrors = append(configErrors, "Command ["+name+"] has neither 'cmd' or 'imports' set. Check your yaml file.")
			continue
		}

		// Check if a command has 'cmd' AND 'imports' set.
		if cmd.Cmd != "" && cmd.Imports != nil {
			configErrors = append(configErrors, "Command ["+name+"] has both 'cmd' and 'imports' set, but only one is allowed. Check your yaml file.")
			continue
		}

		// Check that a command with 'imports' set has at least one entry.
		if cmd.Imports != nil && len(cmd.Imports) == 0 {
			configErrors = append(configErrors, "Command ["+name+"] has 'imports' set, but it is empty. Check your yaml file.")
			continue
		}

		newCmd := &cobra.Command{
			Use:     name,
			Aliases: cmd.Aliases,
			// A command that runs something owns every argument after its
			// own name, exactly as it was typed - ahoy is a wrapper, so
			// `ahoy chown -R .` must reach chown with its -R intact.
			// FParseErrWhitelist is not enough here: it suppresses the
			// unknown-flag error but pflag still drops the flag AND the
			// token after it, silently mangling the wrapped command
			// (issue #182). Ahoy's own flags are read by the pre-parser in
			// flag.go, which only looks before the command name.
			//
			// Commands that only group imports keep normal parsing so that
			// subcommand routing and help still work.
			DisableFlagParsing: cmd.Cmd != "",
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
			// Capture variables for the closure.
			cmdString := cmd.Cmd
			cmdEnv := cmd.Env
			cmdName := name

			newCmd.RunE = func(cobraCmd *cobra.Command, args []string) error {
				// 'bash -c' passes arguments starting with $0, so $@ skips the first item.
				// See http://stackoverflow.com/questions/41043163/xargs-sh-c-skipping-the-first-argument
				var cmdItems []string
				var cmdArgs []string
				var cmdEntrypoint []string

				// A leading "--" is Ahoy's, and is dropped. It dates from
				// urfave/cli, where it was the only way to stop the parser
				// claiming the flags that followed it (issue #100). Ahoy no
				// longer parses a command's arguments at all, so it is now
				// only a readability habit, but scripts and aliases still
				// use it and it costs nothing to honour.
				//
				// Every later "--" is the command's own data. Plenty of
				// tools require one - kubectl exec, npm run, cargo run,
				// git checkout - and stripping those silently changed what
				// the user asked for (issue #186).
				cmdArgs = args
				if len(cmdArgs) > 0 && cmdArgs[0] == "--" {
					cmdArgs = cmdArgs[1:]
				}

				// Replace the entry point placeholders.
				cmdEntrypoint = config.Entrypoint[:]
				for i := range cmdEntrypoint {
					switch cmdEntrypoint[i] {
					case "{{cmd}}":
						cmdEntrypoint[i] = cmdString
					case "{{name}}":
						cmdEntrypoint[i] = cmdName
					}
				}
				cmdItems = append(cmdEntrypoint, cmdArgs...)

				// Collect environment variables.
				cmdEnvVars := append([]string{}, envVars...)

				// If defined, include any command-level environment variables.
				// Note that this will intentionally override any conflicting variables
				// defined in the 'global' env file.
				if len(cmdEnv) > 0 {
					for _, envFile := range cmdEnv {
						vars := s.getEnvironmentVars(envFile, s.srcDir)
						if vars != nil {
							cmdEnvVars = append(cmdEnvVars, vars...)
						}
					}
				}

				// Inject ahoy-specific environment variables so subprocesses can
				// identify the running binary and the invoked command name.
				ahoyEnvVars := []string{"AHOY_COMMAND_NAME=" + cmdName}
				if s.ahoyExecutable != "" {
					ahoyEnvVars = append(ahoyEnvVars, "AHOY_CMD="+s.ahoyExecutable)
				}
				cmdEnvVars = append(cmdEnvVars, ahoyEnvVars...)

				if s.verbose {
					log.Println("===> Ahoy", cmdName, "from", s.sourcefile, ":", cmdItems)
				}
				// gosec: executing operator-supplied commands from .ahoy.yml is
				// the entire purpose of ahoy, so the "tainted input" here is the
				// intended input and G204 does not apply.
				// noctx: ahoy runs each command to completion and has no
				// cancellation or timeout semantics, so CommandContext with a
				// background context would add ceremony without behaviour.
				command := exec.Command(cmdItems[0], cmdItems[1:]...) //nolint:gosec,noctx // Running user-defined commands to completion is ahoy's core function.
				command.Dir = s.srcDir
				command.Stdout = os.Stdout
				command.Stdin = os.Stdin
				command.Stderr = os.Stderr
				// Build the environment so cmdEnvVars always take precedence.
				// macOS getenv(3) returns the first match, so we put cmdEnvVars
				// first and append inherited entries only when their key is not
				// already covered.
				overridden := make(map[string]bool, len(cmdEnvVars))
				for _, kv := range cmdEnvVars {
					if i := strings.Index(kv, "="); i > 0 {
						overridden[kv[:i]] = true
					}
				}
				mergedEnv := make([]string, len(cmdEnvVars), len(cmdEnvVars)+len(command.Environ()))
				copy(mergedEnv, cmdEnvVars)
				for _, kv := range command.Environ() {
					if i := strings.Index(kv, "="); i <= 0 || !overridden[kv[:i]] {
						mergedEnv = append(mergedEnv, kv)
					}
				}
				command.Env = mergedEnv
				if err := command.Run(); err != nil {
					fmt.Fprintln(os.Stderr)
					return err
				}
				return nil
			}
		}

		if cmd.Imports != nil {
			subCommands := s.getSubCommands(cmd.Imports)
			if len(subCommands) == 0 {
				if !cmd.Optional {
					errorMsg := fmt.Sprintf("Command [%s] has 'imports' set, but no commands were found.", sanitiseLogValue(name))

					// List any import files that are missing to help diagnose the issue.
					var missingFiles []string
					for _, importPath := range cmd.Imports {
						fullPath := expandPath(importPath, s.srcDir)
						if !fileExists(fullPath) {
							missingFiles = append(missingFiles, importPath)
						}
					}

					if len(missingFiles) > 0 {
						errorMsg += fmt.Sprintf("\n\nMissing import files: %s", sanitiseLogValue(strings.Join(missingFiles, ", ")))
						errorMsg += "\n\nSolutions:"
						errorMsg += "\n1. Create the missing files"
						errorMsg += "\n2. Mark imports as optional with 'optional: true'"
						if !VersionSupports(GetAhoyVersion(), featureOptionalImports) {
							errorMsg += fmt.Sprintf("\n3. Upgrade Ahoy to v%s+ for optional import support", FeatureSupport[featureOptionalImports])
						}
						errorMsg += "\n\nFor more help, run: ahoy config validate"
					}

					s.logger(logLevelFatal, errorMsg)
				} else {
					if !VersionSupports(GetAhoyVersion(), featureOptionalImports) {
						errorMsg := fmt.Sprintf("Command [%s] uses 'optional: true' but this Ahoy version (%s) doesn't support optional imports.", sanitiseLogValue(name), GetAhoyVersion())
						errorMsg += fmt.Sprintf("\n\nThis feature requires Ahoy %s or later.", FeatureSupport[featureOptionalImports])
						errorMsg += "\n\nSolutions:"
						errorMsg += "\n1. Upgrade Ahoy to the latest version"
						errorMsg += "\n2. Remove 'optional: true' and create the missing import files"
						errorMsg += "\n\nFor more help, run: ahoy config validate"
						s.logger(logLevelFatal, errorMsg)
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

	for _, e := range configErrors {
		s.logger(logLevelError, e)
	}
	if len(configErrors) > 0 {
		s.logger(logLevelFatal, "Fix the above configuration errors and try again.")
	}

	return exportCmds
}

func (s *appState) addDefaultCommands(commands []*cobra.Command) []*cobra.Command {
	// 'ahoy config' command group with 'validate' and 'init' subcommands.
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Ahoy configuration.",
	}
	configCmd.SetHelpFunc(commandHelpFunc)

	configValidateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate and diagnose an Ahoy configuration file.",
		Run:   s.validateCommandAction,
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

// bashComplete prints the list of subcommands as the default app completion method.
func (s *appState) bashComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	s.logger(logLevelDebug, "bashComplete()")

	completions := []string{}
	for _, command := range cmd.Root().Commands() {
		completions = append(completions, command.Name())
		completions = append(completions, command.Aliases...)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

// noArgsAction is the application-wide default action when no flags or arguments
// are passed or when a command doesn't exist.
func (s *appState) noArgsAction(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		msg := "Command not found for '" + strings.Join(args, " ") + "'"
		s.logger(logLevelFatal, msg)
	}

	if err := cmd.Help(); err != nil {
		s.logger(logLevelError, err.Error())
	}

	if s.srcFile == "" {
		s.logger(logLevelError, "No .ahoy.yml found. You can use 'ahoy init' to download an example.")
	}

	helpRequested, _ := cmd.Flags().GetBool("help")
	versionRequested, _ := cmd.Flags().GetBool("version")
	if !helpRequested && !versionRequested {
		s.logger(logLevelWarn, "Missing flag or argument.")
		os.Exit(1)
	}

	// Exit gracefully if we get to here.
	os.Exit(0)
}

// beforeCommand is a PersistentPreRunE hook that handles --version and --help
// flag processing before cobra executes each command.
func (s *appState) beforeCommand(cmd *cobra.Command, args []string) error {
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
					if err := subcmd.Help(); err != nil {
						s.logger(logLevelError, err.Error())
					}
					os.Exit(0)
				}
			}
		}
		if err := cmd.Help(); err != nil {
			s.logger(logLevelError, err.Error())
		}
		os.Exit(0)
	}
	return nil
}

func (s *appState) setupApp(localArgs []string) *cobra.Command {
	s.initFlags(localArgs)

	// initFlags() pre-parsed sourcefile and verbose from the legacy
	// single-dash forms (-f, -verbose, etc.) - see flag.go for the full
	// rationale. The cobra flag definitions below would re-bind those
	// same variables and reset them to their zero values, so we capture
	// the parsed values now and pass them as the cobra flag defaults.
	parsedSourcefile := s.sourcefile
	parsedVerbose := s.verbose
	parsedSimulateVersion := simulateVersion

	// Create root command.
	rootCmd := &cobra.Command{
		Use:     "ahoy",
		Version: version,
		Short:   "Creates a configurable cli app for running commands.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s.noArgsAction(cmd, args)
			return nil
		},
		PersistentPreRunE: s.beforeCommand,
		ValidArgsFunction: s.bashComplete,
	}

	// Set up global flags with the parsed values as defaults.
	rootCmd.PersistentFlags().BoolVarP(&s.verbose, "verbose", "v", parsedVerbose, "Output extra details like the commands to be run.")
	rootCmd.PersistentFlags().StringVarP(&s.sourcefile, "file", "f", parsedSourcefile, "Use a specific ahoy file.")
	rootCmd.PersistentFlags().Bool("help", false, "show help")
	rootCmd.PersistentFlags().Bool("version", false, "print the version")
	rootCmd.PersistentFlags().Bool("generate-bash-completion", false, "")

	// Add hidden --simulate-version flag for testing the validation system
	// against older Ahoy versions without needing to rebuild the binary.
	rootCmd.PersistentFlags().StringVar(&simulateVersion, "simulate-version", parsedSimulateVersion, "simulate a specific Ahoy version for testing")

	// Mark help, version, and internal flags as hidden since we handle them manually.
	for _, name := range []string{"help", "version", "generate-bash-completion", "simulate-version"} {
		_ = rootCmd.PersistentFlags().MarkHidden(name)
	}

	// `ahoy help <command>` is the way to reach a command's own help,
	// because `--help` after a command name belongs to that command and is
	// passed through to it (see getCommands). Hidden so it doesn't clutter
	// the list of the user's own commands.
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help [command]",
		Short:  "Show help for a command.",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			target, _, err := cmd.Root().Find(args)
			if target == nil || err != nil {
				s.logger(logLevelError, "Command not found for '"+strings.Join(args, " ")+"'")
				os.Exit(1)
			}
			if err := target.Help(); err != nil {
				s.logger(logLevelError, err.Error())
			}
		},
	})

	s.importVisited = map[string]bool{}

	var err error
	s.srcFile, err = s.getConfigPath()
	if errors.Is(err, errNoConfig) {
		// No config found - supply default commands only and return early.
		commands := s.addDefaultCommands([]*cobra.Command{})
		rootCmd.AddCommand(commands...)
		return rootCmd
	} else if err != nil {
		s.logger(logLevelFatal, err.Error())
	} else {
		s.srcDir = filepath.Dir(s.srcFile)
		s.importVisited[normalizePath(s.srcFile)] = true
		config, err := getConfig(s.srcFile)
		if err != nil {
			s.logger(logLevelFatal, err.Error())
		}
		commands := s.getCommands(config)
		commands = s.addDefaultCommands(commands)
		rootCmd.AddCommand(commands...)
		if config.Usage != "" {
			rootCmd.Short = config.Usage
		}
	}

	// Set up custom help template.
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
Use 'ahoy help <command>' for detailed information about a command.
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
	state := newAppState()
	if exe, err := os.Executable(); err == nil {
		state.ahoyExecutable = exe
	}
	rootCmd := state.setupApp(os.Args[1:])

	// Ahoy's own flags were consumed by the pre-parser in flag.go, which
	// stops at the command name. Give cobra only what follows so it never
	// gets the chance to claim a flag meant for the wrapped command.
	rootCmd.SetArgs(state.commandArgs)

	// Check for invalid flag error from initFlags - show help and exit 1.
	if state.invalidFlagError != "" {
		fmt.Print(state.invalidFlagError)
		if err := rootCmd.Help(); err != nil {
			log.Printf("help error: %v", err)
		}
		os.Exit(1)
	}

	// Check for -version and -help flags set during initFlags (single-dash versions).
	// This handles single-dash versions that cobra doesn't support.
	if state.versionFlagSet {
		if version != "" {
			fmt.Println(version)
		}
		os.Exit(0)
	}

	if state.helpFlagSet {
		if err := rootCmd.Help(); err != nil {
			log.Printf("help error: %v", err)
		}
		os.Exit(0)
	}

	// Handle bash completion flag - print completions and exit.
	if state.bashCompletionFlagSet {
		for _, command := range rootCmd.Commands() {
			if !command.Hidden {
				fmt.Println(command.Name())
			}
		}
		os.Exit(0)
	}

	// Ahoy and its subprocesses share the stderr ahoy was started with.
	// Nothing is interposed on it: an intervening pipe would strip the
	// terminal properties that interactive children rely on (issue #180),
	// and cobra's own error output is already suppressed via SilenceErrors.
	err := rootCmd.Execute()

	if err != nil {
		// Cobra has SilenceErrors=true so the error has not been printed.
		// Translate "unknown command" into ahoy's friendly equivalent.
		if strings.Contains(err.Error(), "unknown command") {
			// Format: "unknown command \"something\" for \"ahoy\""
			parts := strings.Split(err.Error(), "\"")
			if len(parts) >= 2 {
				cmdName := parts[1]
				msg := "Command not found for '" + cmdName + "'"
				state.logger("fatal", msg)
			}
		}
		os.Exit(1)
	}
}
