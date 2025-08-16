package functions

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		name        string
		setUserID   interface{}
		expectUser  string
		expectError bool
	}{
		{
			name:        "Valid user ID",
			setUserID:   "user123",
			expectUser:  "user123",
			expectError: false,
		},
		{
			name:        "Empty user ID",
			setUserID:   "",
			expectUser:  "",
			expectError: true,
		},
		{
			name:        "No user ID in context",
			setUserID:   nil,
			expectUser:  "",
			expectError: true,
		},
		{
			name:        "Wrong type user ID",
			setUserID:   123,
			expectUser:  "",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(nil)

			if test.setUserID != nil {
				c.Set("userID", test.setUserID)
			}

			userID, err := GetUser(c)

			if test.expectError {
				assert.Error(t, err)
				assert.Equal(t, "unauthorized", err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectUser, userID)
			}
		})
	}
}
