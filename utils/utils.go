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
package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type RemoteFileMetadata struct {
	Size     uint64
	Modified *time.Time
}

func GetRemoteFileMetadata(filePath string) (*RemoteFileMetadata, error) {
	output, err := exec.Command("adb", "shell", "stat", "-c", "%s,%Y", fmt.Sprintf(`"%s"`, filePath)).Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to execute adb command: %v", err)
	}

	if len(output) > 0 {
		outputStr := string(output)
		split := strings.Split(strings.TrimSpace(outputStr), ",")

		fileSizeStr := split[0]
		modifiedStr := split[1]

		fileSize, _ := strconv.ParseUint(fileSizeStr, 10, 64)
		modifiedTimestamp, _ := strconv.ParseInt(modifiedStr, 10, 64)
		modified := time.Unix(modifiedTimestamp, 0)

		metadata := &RemoteFileMetadata{
			Size:     fileSize,
			Modified: &modified,
		}

		return metadata, nil
	}

	return nil, fmt.Errorf("Failed to get remote file metadata")
}

func SyncFile(sourceFile string, destFile string, command string, dryRun bool, checksum bool, debug bool) {
	if !dryRun {
		// Create the directories for the destination file
		destDir := filepath.Dir(destFile)
		err := os.MkdirAll(destDir, 0755)
		if err != nil {
			fmt.Printf("Failed to create destination directories: %v\n", err)
			return
		}

		// Perform the synchronization
		output, err := exec.Command("adb", command, sourceFile, destFile).CombinedOutput()

		if err != nil {
			errorMessage := strings.TrimSpace(string(output))
			fmt.Printf("Failed to sync: %s -> %s\nError: %s\n", color.RedString(sourceFile), color.RedString(destFile), color.RedString(errorMessage))
		} else {
			fmt.Printf("Copied: %s -> %s\n\n", color.GreenString(sourceFile), color.GreenString(destFile))
		}

	}
}

func NeedsCopy(sourceFile, destFile string, command string, checksum, debug bool) bool {
	var sourceSize uint64
	var destSize uint64

	var sourceModified time.Time
	var destModified time.Time

	if command == "pull" {
		sourceInfo, err := GetRemoteFileMetadata(sourceFile)
		if err != nil {
			if debug {
				fmt.Println(color.RedString("%s doesn't exist! ", sourceFile) + color.MagentaString("needsCopy: false"))
			}
			return false
		}

		destInfo, err := os.Stat(destFile)
		if err != nil {
			if debug {
				fmt.Println(color.RedString("%s doesn't exist! ", destFile) + color.MagentaString("needsCopy: true"))
			}
			return true
		}

		sourceSize = sourceInfo.Size
		destSize = uint64(destInfo.Size())

		sourceModified = *sourceInfo.Modified
		destModified = destInfo.ModTime()

		if checksum {
			sourceHash := ComputeMD5Checksum(sourceFile, true)
			destHash := ComputeMD5Checksum(destFile, false)

			if debug {
				fmt.Printf("MD5 diff: %s -> %s\nSource MD5: %s\nDest MD5: %s\n",
					color.CyanString(sourceFile),
					color.CyanString(destFile),
					color.MagentaString(sourceHash),
					color.MagentaString(destHash),
				)
			}

			if sourceSize != destSize && sourceModified.After(destModified) || sourceHash != destHash {
				if debug {
					fmt.Println(color.RedString("File Size & Last Modified or checksum mismatch! ") + color.MagentaString("needsCopy: true"))
				}
				return true
			}
		} else {
			if sourceSize != destSize && sourceModified.After(destModified) {
				if debug {
					fmt.Println(color.RedString("File Size & Last Modified mismatch! ") + color.MagentaString("needsCopy: true"))
				}
				return true
			}
		}

	} else if command == "push" {
		sourceInfo, err := os.Stat(sourceFile)
		if err != nil {
			if debug {
				fmt.Println(color.RedString("%s doesn't exist! ", sourceFile) + color.MagentaString("needsCopy: false"))
			}
			return false
		}

		destInfo, err := GetRemoteFileMetadata(destFile)
		if err != nil {
			if debug {
				fmt.Println(color.RedString("%s doesn't exist! ", destFile) + color.MagentaString("needsCopy: true"))
			}
			return true
		}

		sourceSize = uint64(sourceInfo.Size())
		destSize = destInfo.Size

		sourceModified = sourceInfo.ModTime()
		destModified = *destInfo.Modified

		if checksum {
			sourceHash := ComputeMD5Checksum(sourceFile, false)
			destHash := ComputeMD5Checksum(destFile, true)

			if debug {
				fmt.Printf("MD5 diff: %s -> %s\nSource MD5: %s\nDest MD5: %s\n",
					color.CyanString(sourceFile),
					color.CyanString(destFile),
					color.MagentaString(sourceHash),
					color.MagentaString(destHash),
				)
			}

			if sourceSize != destSize && sourceModified.After(destModified) || sourceHash != destHash {
				if debug {
					fmt.Println(color.RedString("File Size & Last Modified or checksum mismatch! ") + color.MagentaString("needsCopy: true"))
				}
				return true
			}
		} else {
			if sourceSize != destSize && sourceModified.After(destModified) {
				if debug {
					fmt.Println(color.RedString("File Size & Last Modified mismatch! ") + color.MagentaString("needsCopy: true"))
				}
				return true
			}
		}
	}

	return false
}

func SanitizeAndroidPath(path string) string {
	return filepath.ToSlash(path)
}

func GetFilesRecursive(path string, command string) ([]string, error) {
	if command == "push" {
		var files []string
		err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, file)
			}
			return nil
		})
		return files, err

	} else if command == "pull" {
		// Run adb command to list files and directories recursively
		cmd := exec.Command("adb", "shell", "ls", "-R", path)
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		// Parse the output to extract file paths
		filePaths := parseAdbLsOutput(string(output))

		return filePaths, nil
	}
	return nil, errors.New("Invalid command")
}

func ComputeMD5Checksum(filePath string, useADB bool) string {
	if useADB {
		output, err := exec.Command("adb", "shell", "md5sum", fmt.Sprintf(`"%s"`, filePath)).Output()
		if err != nil {
			return ""
		}

		if len(output) > 0 {
			outputStr := string(output)
			split := strings.Split(strings.TrimSpace(outputStr), " ")
			if len(split) > 0 {
				md5sum := split[0]
				return md5sum
			}
		}

		return ""
	} else {
		fileContent, err := ioutil.ReadFile(filePath)
		if err != nil {
			return ""
		}

		digest := md5.Sum(fileContent)
		md5sum := hex.EncodeToString(digest[:])
		return md5sum
	}
}

// Parse adb ls output to extract file paths
func parseAdbLsOutput(output string) []string {
	lines := strings.Split(output, "\n")

	var filePaths []string
	dirPath := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ":") {
			// New directory path
			dirPath = strings.TrimSuffix(line, ":")
		} else if line != "" {
			// File path
			filePath := filepath.Join(dirPath, line)
			filePaths = append(filePaths, filePath)
		}
	}

	return filePaths
}
