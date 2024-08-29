package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	debug       bool
	fileOffsets = make(map[string]int64)
	mu          sync.Mutex
)

func newOffset(file *os.File) int64 {
	newOffset, _ := file.Seek(0, io.SeekCurrent)
	return newOffset
}

func processFile(filePath string) {
	mu.Lock()
	offset := fileOffsets[filePath]
	mu.Unlock()

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0444)
	if err != nil {
		log.Println("error:", err)
		return
	}
	defer file.Close()

	file.Seek(offset, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			fmt.Println(line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Println("error reading file:", err)
	}

	updateOffset(filePath, newOffset(file))
}

func updateOffset(filePath string, offset int64) {
	mu.Lock()
	defer mu.Unlock()
	fileOffsets[filePath] = offset
}

func handleFileEvent(event fsnotify.Event) {
	if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
		if debug {
			log.Println(event.Op, event.Name)
		}
		processFile(event.Name)
	}
}

func watchDir(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				handleFileEvent(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	if err = watcher.Add(dir); err != nil {
		log.Fatal(err)
	}
	<-done
}

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create directory: %s", dir)
		}
	}
}

func processInitialFiles(dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			processFile(path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to process initial files: %s", err)
	}
}

func main() {
	dir := flag.String("dir", "", "Directory to watch")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()

	if *dir == "" {
		log.Fatal("Please provide a directory to watch using the -dir flag")
	}

	*dir = filepath.Clean(*dir) + string(os.PathSeparator)
	ensureDirExists(*dir)
	processInitialFiles(*dir)
	watchDir(*dir)
}
