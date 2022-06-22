package main

import (
	"flag"

	"github.com/urfave/cli"
)

var globalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:        "verbose, v",
		Usage:       "Output extra details like the commands to be run.",
		EnvVar:      "AHOY_VERBOSE",
		Destination: &verbose,
	},
	cli.StringFlag{
		Name:        "file, f",
		Usage:       "Use a specific ahoy file.",
		Destination: &sourcefile,
	},
	cli.BoolFlag{
		Name:  "help, h",
		Usage: "show help",
	},
	cli.BoolFlag{
		Name:  "version",
		Usage: "print the version",
	},
	cli.BoolFlag{
		Name: "generate-bash-completion",
	},
}

func flagSet(name string, flags []cli.Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

func initFlags(incomingFlags []string) {

	// Reset the sourcedir for when we're testing. Otherwise the global state
	// is preserved between the tests.
	AhoyConf.srcDir = ""

	// Grab the global flags first ourselves so we can customize the yaml file loaded.
	// Flags are only parsed once, so we need to do this before cli has the chance to?
	tempFlags := flagSet("tempFlags", globalFlags)
	tempFlags.Parse(incomingFlags)
}

func overrideFlags(app *cli.App) {
	app.Flags = globalFlags
	app.HideVersion = true
	app.HideHelp = true
}
