package main

import (
	"context"
	"crypto/md5"
	"log"
	"os"
	"sync"

	"golang.org/x/sync/semaphore"
)

const maxGoroutines = 1

type Watcher struct {
	Events chan string
	files  []string
	hashes map[string][16]byte
	mu     sync.RWMutex
}

func NewWatcher() *Watcher {
	return &Watcher{
		Events: make(chan string),
		hashes: make(map[string][16]byte),
	}
}

func (w *Watcher) Add(filename string) {
	w.files = append(w.files, filename)

	w.mu.Lock()
	defer w.mu.Unlock()
	w.hashes[filename] = hashFile(filename)
}

func (w *Watcher) Start() {
	go w.poll()
}

// goroutinesの数だけセマフォを作って、ファイルを監視
// 変更はhashで確認する
func (w *Watcher) poll() {
	ctx := context.Background()
	sem := semaphore.NewWeighted(maxGoroutines)
	var idx int
	for {
		// セマフォを獲得できるまでブロック
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Print(err)
			continue
		}
		go func(filename string) {
			defer sem.Release(1)
			if w.hasChanged(filename) {
				newHash := hashFile(filename)
				w.mu.Lock()
				w.hashes[filename] = newHash
				w.mu.Unlock()
				w.Events <- filename
			}
		}(w.files[idx])

		if idx < len(w.files)-1 {
			idx++
		} else {
			idx = 0
		}
	}
}

func (w *Watcher) hasChanged(filename string) bool {
	newHash := hashFile(filename)
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.hashes[filename] != newHash
}

func hashFile(filename string) [16]byte {
	b, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return md5.Sum(b)
}
