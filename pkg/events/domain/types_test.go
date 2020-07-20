package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetEventTypes(t *testing.T) {
	assert.NotEmpty(t, GetEventTypes())
	for _, eventType := range GetEventTypes() {
		assert.NotEqual(t, "", string(eventType))
	}
}
