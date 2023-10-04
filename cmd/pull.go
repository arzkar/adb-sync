/*
Copyright 2023 Arbaaz Laskar

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arzkar/adb-sync/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var pullCmd *cobra.Command

func init() {
	pullCmd = &cobra.Command{
		Use:   "pull [source_path] [destination_path]",
		Short: "Pull files from the source_path to the destination_path",
		Args:  cobra.ExactArgs(2),
		Run:   runPullCommand,
	}

	var checksum bool
	var dryRun bool
	var debug bool
	pullCmd.Flags().BoolVar(&checksum, "checksum", false, "Compare files using MD5 checksums")
	pullCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Perform a trial run without making any changes")
	pullCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")

	rootCmd.AddCommand(pullCmd)
}

func runPullCommand(cmd *cobra.Command, args []string) {
	sourcePath := args[0]
	destinationPath := args[1]
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	checksum, _ := cmd.Flags().GetBool("checksum")
	debug, _ := cmd.Flags().GetBool("debug")

	pull(sourcePath, destinationPath, dryRun, checksum, debug)
}

func pull(sourcePath string, destinationPath string, dryRun bool, checksum bool, debug bool) {
	sourceFiles, err := utils.GetFilesRecursive(sourcePath, "pull")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	destinationFiles, err := utils.GetFilesRecursive(destinationPath, "push")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Create a map of destination files for faster lookups
	destinationFileMap := make(map[string]bool)
	for _, destFile := range destinationFiles {
		destinationFileMap[destFile] = true
	}

	for _, file := range sourceFiles {
		relativePath, err := filepath.Rel(sourcePath, file)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		destFile := filepath.Join(destinationPath, relativePath)
		sanitizedSourceFile := utils.SanitizeAndroidPath(file)

		// Create the destination directory if it doesn't exist
		destDir := filepath.Dir(destFile)
		err = os.MkdirAll(destDir, 0755)

		if err != nil {
			fmt.Printf("Failed to create destination directory: %v\n", err)
			return
		}

		if utils.NeedsCopy(sanitizedSourceFile, destFile, "pull", checksum, debug) {
			fmt.Printf("Copying: %s -> %s\n", color.BlueString(sanitizedSourceFile), color.BlueString(destFile))
			utils.SyncFile(sanitizedSourceFile, destFile, "pull", dryRun, checksum, debug)
		} else {
			fmt.Printf("Skipped: %s -> %s (File already exists and is up to date)\n\n", color.RedString(sanitizedSourceFile), color.RedString(destFile))
		}

		// Remove the destination file & its parent from the map if it exists
		delete(destinationFileMap, destFile)
		delete(destinationFileMap, filepath.Dir(destFile))
	}

	// Remove any remaining files in the destination directory that were not in the source
	for destFile := range destinationFileMap {
		fmt.Printf("Removing: %s\n", color.YellowString(destFile))
		if !dryRun {
			err := os.Remove(destFile)
			if err != nil {
				fmt.Printf("Failed to remove file: %v\n", err)
			}
		}
	}
}
