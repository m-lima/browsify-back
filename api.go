package main

import (
	"log"
	"os"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type Entry struct {
	Name      string
	Directory bool
	Size      int64
	Date      time.Time
}

var (
	apiLogStd = log.New(os.Stdout, "[api] ", log.Ldate|log.Ltime)
	apiLogErr = log.New(os.Stderr, "ERROR [api] ", log.Ldate|log.Ltime)
)

func pathAllowed(user *User, file string, isDir bool) bool {
	if user.Permissions.Admin {
		return true
	}

	for _, pattern := range user.Permissions.Paths {
		if isDir && strings.HasPrefix(pattern, file) {
			return true
		}

		if pattern[len(pattern)-1:] == "*" && strings.HasPrefix(file, pattern[:len(pattern)-1]) {
			return true
		}

		if pattern == file {
			return true
		}
	}

	return false
}

func shouldDisplayEntry(user *User, showHidden bool, showProtected bool, file os.FileInfo) bool {

	if file.Name()[0] == '.' && (!showHidden || !user.Permissions.CanShowHidden) {
		return false
	}

	if file.Mode()&0004 == 0 && (!showProtected || !user.Permissions.CanShowProtected) {
		return false
	}

	return true
}

//// GET
// Returns requested path as either a json directory or the file itself
func apiHandler(response http.ResponseWriter, request *http.Request) {

	// Validate method
	if request.Method != http.MethodGet {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check if logged in
	user, status, err := getUser(request)
	if err != nil {
		apiLogErr.Println(err)
		response.WriteHeader(status)
		return
	}

	// Map request to filesystem path
	urlPath := request.URL.Path
	systemPath := config.System.Root + urlPath
	apiLogStd.Println(user.Email, "is requesting", urlPath)

	// Not found
	if _, err := os.Stat(systemPath); err != nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	// Extract query options
	showHidden := false
	showProtected := false
	query := strings.Split(request.URL.RawQuery, "&")
	for _, value := range query {
		switch value {
		case "showHidden":
			showHidden = true
		case "showProtected":
			showProtected = true
		}
	}

	// Should not display
	if file, err := os.Stat(systemPath); err != nil || !shouldDisplayEntry(user, showHidden, showProtected, file) || !pathAllowed(user, urlPath, file.IsDir()) {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	// Try folder
	{
		files, err := ioutil.ReadDir(systemPath)
		if err == nil {
			var entries []Entry
			for _, file := range files {
				if shouldDisplayEntry(user, showHidden, showProtected, file) && pathAllowed(user, filepath.Join(urlPath, file.Name()), file.IsDir()) {
					entries = append(entries, Entry{
						Name:      file.Name(),
						Directory: file.IsDir(),
						Size:      file.Size(),
						Date:      file.ModTime(),
					})
				}
			}

			response.Header().Set("Content-type", "application/json")
			json.NewEncoder(response).Encode(entries)

			return
		}
	}

	// Try file
	{
		http.ServeFile(response, request, systemPath)
		return
	}
}
