<h1 align="center">adb-sync</h1>


[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/arzkar)

A CLI to sync between android & local system using adb
<br><br>

# Installation

## Using Go

```
go install github.com/arzkar/adb-sync@latest
```

## From Github

1. Make sure you have Git installed on your system.
2. Download the latest release of adb-sync from the [Releases](https://github.com/arzkar/adb-sync/releases) page.
3. Extract the downloaded archive to a location of your choice.
4. Add the extracted directory to your system's PATH.

# Usage

```
> adb-sync
adb-sync v0.1.0
Copyright (c) Arbaaz Laskar <arzkar.dev@gmail.com>

A CLI to sync between android & local system using adb

Usage:
  adb-sync [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  pull        Pull files from the source_path to the destination_path
  push        Push files from the source_path to the destination_path

Flags:
  -h, --help   help for adb-sync

Use "adb-sync [command] --help" for more information about a command.
```

# Example

### Pull

```
Usage:
  adb-sync pull [source_path] [destination_path] [flags]

Flags:
      --checksum   Compare files using MD5 checksums
      --debug      Enable debug mode
      --dry-run    Perform a trial run without making any changes
  -h, --help       help for pull
```

### Push

```
Usage:
  adb-sync push [source_path] [destination_path] [flags]

Flags:
      --checksum   Compare files using MD5 checksums
      --debug      Enable debug mode
      --dry-run    Perform a trial run without making any changes
  -h, --help       help for push
```
