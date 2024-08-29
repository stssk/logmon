# logmon

A simple file watcher to log new lines.

## Overview

`logmon` is a Go-based utility that monitors a specified directory for file changes and logs new lines added to the files. It uses the `fsnotify` package to watch for file system events.

## Features

- Watches a specified directory for file changes.
- Logs new lines added to the files.
- Skips hidden files and directories.
- Supports debug logging.

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/stssk/logmon.git
    cd logmon
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

## Usage

Run the `logmon` command with the `-dir` flag to specify the directory to watch:

```sh
go run main.go -dir /path/to/directory