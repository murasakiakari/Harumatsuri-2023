package main

import (
	"fmt"
	"regexp"
	"syscall"
	"gatewayserver/backend"

	"golang.org/x/term"
)

var validPattern = regexp.MustCompile(`^[0-9a-zA-Z]+$`).MatchString

func createAccount() error {
	// username
	fmt.Print("Please enter username: ")
	var username string
	_, err := fmt.Scanln(&username)
	if err != nil {
		return err
	}
	if !validPattern(username) {
		return fmt.Errorf("username contain invalid character")
	}

	// first password
	fmt.Print("Please enter password: ")
	input, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return err
	}
	password := string(input)
	if !validPattern(password) {
		return fmt.Errorf("password contain invalid character")
	}

	// second password
	fmt.Println()
	fmt.Print("Please enter password again: ")
	input, err = term.ReadPassword(syscall.Stdin)
	if err != nil {
		return err
	}
	if password != string(input) {
		return fmt.Errorf("the input password are not the same")
	}

	// permission
	fmt.Println()
	fmt.Print("Please enter permission: ")
	var permission string
	_, err = fmt.Scanln(&permission)
	if err != nil {
		return err
	}
	if permission != "admin" {
		permission = "user"
	}

	account := &backend.Account{Username: username, Password: password, Permission: permission}
	if err := backend.CreateAccount(account); err != nil {
		return err
	}
	return nil
}
