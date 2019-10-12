package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	notifierLogStd = log.New(os.Stdout, "[notifier] ", log.Ldate|log.Ltime)
	notifierLogErr = log.New(os.Stderr, "ERROR [notifier] ", log.Ldate|log.Ltime)

	userChannels = make(map[string]chan string)
	mutex        = &sync.Mutex{}
)

func bootStrapChannel(user *User) chan string {
	channel := make(chan string)
	go bufferAndNotify(user, channel)
	return channel
}

func buffer(user *User, messages chan string) []string {
	var buffer []string
	for {
		select {
		case message := <-messages:
			buffer = append(buffer, message)
		case <-time.After(15 * time.Minute):
			mutex.Lock()
			delete(userChannels, user.Email)
			mutex.Unlock()
			return buffer
		}
	}
}

func notify(user *User, buffer []string) {
	Send(user, strings.Join(buffer[:], "\n"))
}

func bufferAndNotify(user *User, channel chan string) {
	notify(user, buffer(user, channel))
}

func PushAccessNotification(user *User, path string) {
	if user.Permissions.IgnoreAccess {
		return
	}

	if _, ok := userChannels[user.Email]; !ok {
		mutex.Lock()
		userChannels[user.Email] = bootStrapChannel(user)
		mutex.Unlock()
	}

	access := fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), path)
	userChannels[user.Email] <- access
}
