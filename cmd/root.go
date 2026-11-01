package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	dryRun  bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "flatty [directory]",
	Short: "Prefix files with their directory path",
	Args:  cobra.ExactArgs(1),
	RunE:  run,
}

func init() {
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print renames without applying them")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Log each rename")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(command *cobra.Command, args []string) error {
	rootPath := args[0]

	rootInfo, err := os.Stat(rootPath)
	if err != nil {
		return fmt.Errorf("cannot access directory: %w", err)
	}
	if !rootInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", rootPath)
	}

	rootPath, err = filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	return filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		dir := filepath.Dir(relativePath)
		base := filepath.Base(relativePath)

		prefix := ""
		if dir != "." {
			prefix = strings.ReplaceAll(dir, string(filepath.Separator), "_")
		}

		newBase := base
		if prefix != "" {
			newBase = prefix + "_" + base
		}

		if newBase == base {
			return nil
		}

		dirPath := filepath.Dir(path)
		newPath := filepath.Join(dirPath, newBase)

		if dryRun || verbose {
			fmt.Printf("%s -> %s\n", path, newPath)
		}

		if dryRun {
			return nil
		}

		return os.Rename(path, newPath)
	})
}
