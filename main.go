package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/heartwilltell/scotty"
)

// Variables which are related to Version command.
// Should be specified by '-ldflags' during the build phase.
// Example:
// GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Branch=$BRANCH \
// -X main.Commit=$COMMIT -o api.
var (
	// Branch is the branch this binary built from.
	Branch = "local"
	// Commit is the commit this binary built from.
	Commit = "unknown"
	// BuildTime is the time this binary built.
	BuildTime = time.Now().Format(time.RFC822)
)

//go:embed .golangci.yml
var config embed.FS

const (
	// ErrNotEnoughArgs error indicates that some argument is missing.
	ErrNotEnoughArgs Error = "not enough arguments"

	// filePermission represents fs file permissions.
	filePermission = 0o600
)

func main() {
	rootCmd := scotty.Command{
		Name:  "lintcfg",
		Short: "Strict opinionated golangci-lint config",
	}

	rootCmd.AddSubcommands(
		generateCommand(),
		versionCommand(),
	)

	if err := rootCmd.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func generateCommand() *scotty.Command {
	cmd := scotty.Command{
		Name:  "generate",
		Short: "Creates the .golangci.yml in specified path.",
		Run: func(cmd *scotty.Command, args []string) error {
			if len(args) < 1 {
				cmd.Flags().Usage()

				return fmt.Errorf("%w: path to the directory should be specified", ErrNotEnoughArgs)
			}

			f, openErr := config.Open(".golangci.yml")
			if openErr != nil {
				return fmt.Errorf("failed to open embedded config file: %w", openErr)
			}

			data, readErr := io.ReadAll(f)
			if readErr != nil {
				return fmt.Errorf("failed to read embedded config file: %w", openErr)
			}

			p, pathErr := filepath.Abs(args[0])
			if pathErr != nil {
				return fmt.Errorf("failed to resolve absolute path: %w", pathErr)
			}

			if err := os.WriteFile(filepath.Join(p, "/.golangci.yml"), data, filePermission); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			return nil
		},
	}

	return &cmd
}

func versionCommand() *scotty.Command {
	cmd := scotty.Command{
		Name:  "version",
		Short: "Prints the version of the program.",
		Run: func(cmd *scotty.Command, args []string) error {
			fmt.Printf("Built from: %s [%s]\n", Branch, Commit)
			fmt.Printf("Built on: %s\n", BuildTime)
			fmt.Printf("Built time: %v\n", time.Now().UTC())

			info, ok := debug.ReadBuildInfo()
			if !ok {
				return nil
			}

			fmt.Printf("Go version: %s\n", info.GoVersion)

			return nil
		},
	}

	return &cmd
}

// Error represents package level errors.
type Error string

func (e Error) Error() string { return string(e) }
