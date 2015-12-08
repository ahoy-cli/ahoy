package main

import (
	"flag"
	"github.com/codegangsta/cli"
	"os"
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

func initFlags() {
	// Grab the global flags first ourselves so we can customize the yaml file loaded.
	tempFlags := flagSet("tempFlags", globalFlags)
	tempFlags.Parse(os.Args[1:])
}

func overrideFlags(app *cli.App) {
	app.Flags = globalFlags
	app.HideVersion = true
	app.HideHelp = true
}
