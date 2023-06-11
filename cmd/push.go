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
	"path/filepath"

	"github.com/arzkar/adb-sync/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var pushCmd *cobra.Command

func init() {
	pushCmd = &cobra.Command{
		Use:   "push [source_path] [destination_path]",
		Short: "Push files from the source_path to the destination_path",
		Args:  cobra.ExactArgs(2),
		Run:   runPushCommand,
	}

	var checksum bool
	var dryRun bool
	var debug bool
	pushCmd.Flags().BoolVar(&checksum, "checksum", false, "Compare files using MD5 checksums")
	pushCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Perform a trial run without making any changes")
	pushCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")

	rootCmd.AddCommand(pushCmd)
}

func runPushCommand(cmd *cobra.Command, args []string) {
	sourcePath := args[0]
	destinationPath := args[1]
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	checksum, _ := cmd.Flags().GetBool("checksum")
	debug, _ := cmd.Flags().GetBool("debug")

	push(sourcePath, destinationPath, dryRun, checksum, debug)
}

func push(sourcePath string, destinationPath string, dryRun bool, checksum bool, debug bool) {
	sourceFiles, err := utils.GetFilesRecursive(sourcePath, "push")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, file := range sourceFiles {
		relativePath, err := filepath.Rel(sourcePath, file)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		destFile := filepath.Join(destinationPath, relativePath)
		sanitizedDestFile := utils.SanitizeAndroidPath(destFile)

		if utils.NeedsCopy(file, sanitizedDestFile, "push", checksum, debug) {
			fmt.Printf("Copying: %s -> %s\n", color.BlueString(file), color.BlueString(sanitizedDestFile))
			utils.SyncFile(file, sanitizedDestFile, "push", dryRun, checksum, debug)
		} else {
			fmt.Printf("Skipped: %s -> %s (File already exists and is up to date)\n\n", color.RedString(file), color.RedString(sanitizedDestFile))
		}
	}
}
