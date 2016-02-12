package flag

import (
	"flag"
	"github.com/codegangsta/cli"
	"os"
)

var Verbose bool
var SourceFile string
var BashCompletion bool

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:        "verbose, v",
		Usage:       "Output extra details like the commands to be run.",
		EnvVar:      "AHOY_VERBOSE",
		Destination: &Verbose,
	},
	cli.StringFlag{
		Name:        "file, f",
		Usage:       "Use a specific ahoy file.",
		Destination: &SourceFile,
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

func FlagSet(name string, flags []cli.Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

func InitFlags() {
	// Grab the global flags first ourselves so we can customize the yaml file loaded.
	tempFlags := FlagSet("tempFlags", GlobalFlags)
	tempFlags.Parse(os.Args[1:])
}

func OverrideFlags(app *cli.App) {
	app.Flags = GlobalFlags
	app.HideVersion = true
	app.HideHelp = true
}

func init() {
	flag.StringVar(&SourceFile, "f", "", "specify the sourcefile")
	flag.BoolVar(&BashCompletion, "generate-bash-completion", false, "")
	flag.BoolVar(&Verbose, "verbose", false, "")
}
