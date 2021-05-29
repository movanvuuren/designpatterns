package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type (
	Event struct {
		data     int
		filename string
		fileop   fsnotify.Op
	}

	Observer interface {
		NotifyCallback(Event)
	}

	Subject interface {
		AddListener(Observer)
		RemoveListener(Observer)
		Notify(Event)
	}

	eventObserver struct {
		id int
		//	time time.Time
	}

	eventSubject struct {
		observers sync.Map
	}
)

func (e *eventObserver) NotifyCallback(event Event) {
	fmt.Printf(" - Observer %d event number: %d\n", e.id, event.data)
	fmt.Printf(" - File: %s changed with file operation: %d\n", event.filename, event.fileop)
}

func (s *eventSubject) AddListener(obs Observer) {
	fmt.Printf("Added: %v\n", obs)
	s.observers.Store(obs, struct{}{})
}

func (s *eventSubject) RemoveListener(obs Observer) {
	fmt.Printf("REMOVE: %v\n", obs)
	s.observers.Delete(obs)
}

func (s *eventSubject) Notify(event Event) {
	s.observers.Range(func(key interface{}, value interface{}) bool {
		if key == nil || value == nil {
			return false
		}
		key.(Observer).NotifyCallback(event)
		return true
	})
}

func watch(n *eventSubject, filename string) {
	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		x := 0
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				fmt.Printf("EVENT: %#v\n", event)
				n.Notify(Event{data: x, filename: event.Name, fileop: event.Op})
				x++
				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR:", err)
			}
		}
	}()

	if err := watcher.Add(filename); err != nil {
		fmt.Println("ERROR:", err)
	}

	<-done
}

func main() {
	filename := "./"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	fmt.Println("WATCHING:", filename)

	n := eventSubject{
		observers: sync.Map{},
	}

	var obs1 = eventObserver{id: 1}
	var obs2 = eventObserver{id: 2}

	n.AddListener(&obs1)
	n.AddListener(&obs2)

	//Remove observer 2 after 10 seconds
	go func() {
		select {
		case <-time.After(time.Second * 10):
			n.RemoveListener(&obs2)
		}
	}()

	watch(&n, filename)
}
