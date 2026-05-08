package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// InitArgs contains arguments for the init command.
type InitArgs struct {
	Force bool
	URL   string
}

// downloadFile downloads a file from the given URL and saves it to the specified path.
// Only http:// and https:// URLs are accepted; all other schemes are rejected.
func downloadFile(rawURL, destPath string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid URL %q: only http and https schemes are supported", rawURL)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL %s: %v", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: server returned %s", resp.Status)
	}

	// Write to a temporary file first, then atomically rename to the destination
	// on success. This prevents a failed or partial download from leaving a
	// corrupt file at the destination path.
	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file %s: %v", tmpPath, err)
	}
	// Always clean up the temp file; harmless no-op after a successful rename.
	defer os.Remove(tmpPath)

	if _, err = io.Copy(out, resp.Body); err != nil {
		out.Close()
		return fmt.Errorf("failed to write file %s: %v", destPath, err)
	}

	// Close explicitly (not via defer) so buffered write errors reported at
	// close time - common on NFS/SMB/Docker bind mounts - are not swallowed.
	if err = out.Close(); err != nil {
		return fmt.Errorf("failed to finalise file %s: %v", destPath, err)
	}

	if err = os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to finalise file %s: %v", destPath, err)
	}

	return nil
}

// RunConfigInit performs the init command functionality.
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

	downloadURL := "https://raw.githubusercontent.com/ahoy-cli/ahoy/master/examples/examples.ahoy.yml"
	if args.URL != "" {
		downloadURL = args.URL
	}

	if err := downloadFile(downloadURL, ".ahoy.yml"); err != nil {
		return fmt.Errorf("failed to download config file: %v", err)
	}

	if args.URL != "" {
		fmt.Println("Your specified .ahoy.yml has been downloaded to the current directory.")
	} else {
		fmt.Println("Example .ahoy.yml downloaded to the current directory. You can customize it to suit your needs!")
	}

	return nil
}

// initCommandAction is the Cobra handler for the init command.
func initCommandAction(cmd *cobra.Command, args []string) {
	initArgs := InitArgs{
		Force: func() bool { f, _ := cmd.Flags().GetBool("force"); return f }(),
	}

	if len(args) > 0 {
		initArgs.URL = args[0]
	}

	if err := RunConfigInit(initArgs); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
