package main

import (
	"log"
	"os"
	"strings"
	"time"

	"net/http"
)

var (
	mainLogStd = log.New(os.Stdout, "[main] ", log.Ldate|log.Ltime)
	mainLogErr = log.New(os.Stderr, "ERROR [main] ", log.Ldate|log.Ltime)
)

// Matches the path requested with a handler
// Will garantee that the path ends with a '/'
func matchHandler(request *http.Request, target string) bool {
	path := request.URL.Path

	if path == target {
		request.URL.Path = "/"
		return true
	}

	if strings.HasPrefix(path, target) && path[len(target)] == '/' {
		request.URL.Path = path[len(target):]
		return true
	}

	return false
}

// Main server handler
// Will dispatch to other handlers based on the path
func handler(response http.ResponseWriter, request *http.Request) {
	mainLogStd.Println("access to:", request.URL.Path)

	// Known handlers
	var handler http.HandlerFunc
	if matchHandler(request, config.Paths.User) {
		handler = userHandler
	} else if matchHandler(request, config.Paths.Api) {
		handler = apiHandler
	}

	// Handler was found
	if handler != nil {
		handler(response, request)
		return
	}

	// Invalid path
	mainLogErr.Println("no handler")
	response.WriteHeader(http.StatusNotFound)
	return
}

func main() {

	// Load users
	err := loadUsers()
	if err != nil {
		mainLogErr.Fatal(err)
	}
	mainLogStd.Printf("allowed users\n%#v\n", allowedUsers)

	// Load configuration
	err = loadConfig("config.json")
	if err != nil {
		mainLogErr.Println(err)
		mainLogErr.Println("using default configuration")
	}
	mainLogStd.Printf("using config\n%#v\n", config)

	// Build server
	server := http.Server{
		Addr:         config.Server.Address,
		Handler:      http.HandlerFunc(handler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     mainLogErr,
	}

	// Listen
	err = server.ListenAndServe()
	if err != nil {
		mainLogErr.Fatal(err)
	}

	mainLogStd.Println("server exited")
}
