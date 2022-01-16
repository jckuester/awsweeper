package internal_test

import (
	"strings"
	"testing"

	"github.com/jckuester/awsweeper/internal"
	"github.com/stretchr/testify/assert"
)

func TestUserConfirmedDeletion(t *testing.T) {
	tests := []struct {
		name                 string
		force                bool
		userInput            string
		expectedConfirmation bool
	}{
		{
			name:                 "confirmed with YES",
			userInput:            "YES",
			expectedConfirmation: true,
		},
		{
			name:      "confirmed with yes",
			userInput: "yes",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualConfirmation := internal.UserConfirmedDeletion(strings.NewReader(tc.userInput))
			assert.Equal(t, tc.expectedConfirmation, actualConfirmation)
		})
	}
}
