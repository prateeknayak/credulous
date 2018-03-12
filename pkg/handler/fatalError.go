package handler

import (
	"os"

	"fmt"
)

func LogAndDieOnFatalError(err error) {
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
