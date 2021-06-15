package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {
	flag.Parse()
	commands := flag.Args()
	files := listFiles(readFromStdin())
	if len(files) == 0 {
		log.Fatal("no file matched to input pattern")
	}

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
				log.Println("modified file:", event.Name)
				log.Printf("%+v", commands)
				if err := execCommands(commands); err != nil {
					log.Fatal(err)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	for _, file := range files {
		err := watcher.Add(file)
		if err != nil {
			log.Fatal(err)
		}
	}

	// prevent terminating
	<-done
}

func readFromStdin() []string {
	s := bufio.NewScanner(os.Stdin)
	s.Split(bufio.ScanWords)

	var result []string
	for s.Scan() {
		result = append(result, s.Text())
	}
	return result
}

// listFiles return filenames which match to given glob patterns.
func listFiles(globs []string) []string {
	var files []string
	for _, g := range globs {
		matches, err := filepath.Glob(g)
		if err != nil {
			continue
		}
		files = append(files, matches...)
	}
	return files
}

func execCommands(commands []string) error {
	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
