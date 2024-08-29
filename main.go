package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var debug bool

func watchDir(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	var mu sync.Mutex
	fileOffsets := make(map[string]int64)

	go handleEvents(watcher, &mu, fileOffsets)

	addWatcherForRootDir(watcher, dir)
	walkAndWatchFiles(watcher, dir, &mu, fileOffsets)

	<-make(chan struct{})
}

func handleEvents(watcher *fsnotify.Watcher, mu *sync.Mutex, fileOffsets map[string]int64) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if debug {
				log.Printf("[debug] event: %s", event)
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				handleCreateEvent(watcher, event, mu, fileOffsets)
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				handleWriteEvent(event, mu, fileOffsets)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func handleCreateEvent(watcher *fsnotify.Watcher, event fsnotify.Event, mu *sync.Mutex, fileOffsets map[string]int64) {
	info, err := os.Stat(event.Name)
	if err == nil && !info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
		mu.Lock()
		fileOffsets[event.Name] = readNewLines(event.Name, 0)
		watcher.Add(event.Name)
		mu.Unlock()
	}
}

func handleWriteEvent(event fsnotify.Event, mu *sync.Mutex, fileOffsets map[string]int64) {
	mu.Lock()
	offset := fileOffsets[event.Name]
	fileOffsets[event.Name] = readNewLines(event.Name, offset)
	mu.Unlock()
}

func addWatcherForRootDir(watcher *fsnotify.Watcher, dir string) {
	err := watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
}

func walkAndWatchFiles(watcher *fsnotify.Watcher, dir string, mu *sync.Mutex, fileOffsets map[string]int64) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != dir {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		if !info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
			mu.Lock()
			fileOffsets[path] = readNewLines(path, 0)
			mu.Unlock()
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func readNewLines(filePath string, offset int64) int64 {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("error opening:", err)
		return offset
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

	newOffset, _ := file.Seek(0, os.SEEK_CUR)
	return newOffset
}

func main() {
	dir := flag.String("dir", "", "Directory to watch")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()

	if *dir == "" {
		log.Fatal("Please provide a directory to watch using the -dir flag")
	}

	*dir = filepath.Clean(*dir) + string(os.PathSeparator)

	watchDir(*dir)
}
