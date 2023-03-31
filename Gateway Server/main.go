package main

import (
	"flag"
	"fmt"
	"gatewayserver/service"
)

var (
	isCreateAccountMode = flag.Bool("CreateAccount", false, "Create Account")
	isServingMode       = flag.Bool("Serving", true, "Start Serving")
)

func main() {
	var err error
	defer func() {
		if err != nil {
			fmt.Printf("exit with error: %v\n", err)
		}
	}()

	flag.Parse()
	if *isCreateAccountMode {
		if err = createAccount(); err != nil {
			err = fmt.Errorf("failed to create account: %w", err)
			return
		}
	}

	if *isServingMode {
		if err = service.New(); err != nil {
			err = fmt.Errorf("failed to start serving: %w", err)
			return
		}
	}
}
