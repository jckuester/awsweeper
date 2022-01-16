package internal

import (
	"fmt"
	"io"

	"github.com/apex/log"
)

// UserConfirmedDeletion asks the user to confirm deletion of resources
func UserConfirmedDeletion(r io.Reader) bool {
	log.Info("Are you sure you want to delete these resources (cannot be undone)? Only YES will be accepted.")
	fmt.Print(fmt.Sprintf("%23v", "Enter a value: "))

	var response string

	_, err := fmt.Fscanln(r, &response)
	if err != nil {
		log.Fatal(err.Error())
	}

	if response == "YES" {
		return true
	}

	return false
}
