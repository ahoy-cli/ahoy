package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli"
)

// InitArgs contains arguments for the init command
type InitArgs struct {
	Force bool
	URL   string
}

// RunConfigInit performs the init command functionality
func RunConfigInit(args InitArgs) error {
	if fileExists(filepath.Join(".", ".ahoy.yml")) {
		if args.Force {
			fmt.Println("Warning: '--force' parameter passed, overwriting .ahoy.yml in current directory.")
		} else {
			fmt.Println("Warning: .ahoy.yml found in current directory.")
			fmt.Fprint(os.Stderr, "Are you sure you wish to overwrite it with an example file, y/N ? ")
			reader := bufio.NewReader(os.Stdin)
			char, _, err := reader.ReadRune()
			if err != nil {
				return fmt.Errorf("failed to read input: %v", err)
			}
			// If "y" or "Y", continue and overwrite.
			// Anything else, exit.
			if char != 'y' && char != 'Y' {
				fmt.Println("Abort: exiting without overwriting.")
				return nil
			}
			if args.URL != "" {
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
	if args.URL != "" {
		wgetURL = args.URL
	}

	grabYaml := "wget " + wgetURL + " -O .ahoy.yml"
	cmd := exec.Command("bash", "-c", grabYaml)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download config file: %v", err)
	}

	if args.URL != "" {
		fmt.Println("Your specified .ahoy.yml has been downloaded to the current directory.")
	} else {
		fmt.Println("Example .ahoy.yml downloaded to the current directory. You can customize it to suit your needs!")
	}

	return nil
}

// initCommandAction performs the init command functionality for CLI
func initCommandAction(c *cli.Context) {
	args := InitArgs{
		Force: c.Bool("force"),
	}

	if len(c.Args()) > 0 {
		args.URL = c.Args()[0]
	}

	if err := RunConfigInit(args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
