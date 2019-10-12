package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"encoding/json"
	"net/http"
)

type User struct {
	GivenName   string          `json:"givenName"`
	FamilyName  string          `json:"familyName"`
	Email       string          `json:"email"`
	Picture     string          `json:"picture"`
	Permissions UserPermissions `json:"permissions"`
}

type UserPermissions struct {
	Admin            bool     `json:"admin"`
	CanShowHidden    bool     `json:"canShowHidden"`
	CanShowProtected bool     `json:"canShowProtected"`
	Paths            []string `json:"paths"`
	IgnoreAccess     bool     `json:"ignoreAccess"`
}

const (
	// Path to users file
	usersPath = "users.json"
)

var (
	userLogStd = log.New(os.Stdout, "[user] ", log.Ldate|log.Ltime)
	userLogErr = log.New(os.Stderr, "ERROR [user] ", log.Ldate|log.Ltime)

	// Main client to access auth server
	// Timeout set because default client may hang forever
	authClient = http.Client{
		Timeout: time.Second * 10,
	}

	// In memory user permissions
	allowedUsers map[string]UserPermissions
)

// Loads the memory with permissions from file
// Should be run synchronously before starting server
func loadUsers() error {
	file, err := os.Open(usersPath)
	if err != nil {
		return fmt.Errorf("failed to open users file: %v", err)
	}

	err = json.NewDecoder(file).Decode(&allowedUsers)
	if err != nil {
		return fmt.Errorf("failed to decode users file: %v", err)
	}

	return nil
}

// Check the user against the in memory permissions
// If successful, the user will be enriched with local permissions
func getUserPermissions(user *User) error {

	// Check if user is registered
	allowedUser, ok := allowedUsers[user.Email]

	if !ok {
		return fmt.Errorf("user '%s' is forbidden", user.Email)
	}

	// Enrich user
	user.Permissions.Admin = allowedUser.Admin
	user.Permissions.CanShowHidden = allowedUser.CanShowHidden
	user.Permissions.CanShowProtected = allowedUser.CanShowProtected
	user.Permissions.Paths = allowedUser.Paths
	user.Permissions.IgnoreAccess = allowedUser.IgnoreAccess

	return nil
}

func getUser(request *http.Request) (*User, int, error) {
	user := User{
		Email:      request.Header.Get("X-USER"),
		GivenName:  request.Header.Get("X-GIVEN-NAME"),
		FamilyName: request.Header.Get("X-FAMILY-NAME"),
		Picture:    request.Header.Get("X-PICTURE"),
	}

	// Enrich user
	err := getUserPermissions(&user)
	if err != nil {
		return nil, http.StatusForbidden, err
	}

	return &user, http.StatusOK, nil
}

//// GET
// Returns the fully enriched user as a json object
func userHandler(response http.ResponseWriter, request *http.Request) {

	// Validate method
	if request.Method != http.MethodGet {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, status, err := getUser(request)
	if err != nil {
		apiLogErr.Println(err)
		response.WriteHeader(status)
		return
	}

	userLogStd.Println("fetched user", user.Email)
	response.Header().Set("Content-type", "application/json")
	json.NewEncoder(response).Encode(user)
}
