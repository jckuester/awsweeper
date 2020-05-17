package internal_test

import (
	"strings"
	"testing"

	"github.com/jckuester/terradozer/internal"
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
		{
			name:                 "force mode",
			force:                true,
			expectedConfirmation: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualConfirmation := internal.UserConfirmedDeletion(strings.NewReader(tc.userInput), tc.force)
			assert.Equal(t, tc.expectedConfirmation, actualConfirmation)
		})
	}
}
