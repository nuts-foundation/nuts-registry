package events

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestEventMarshalling(t *testing.T) {
	// Setup
	const fileName = "../../test_data/valid_files/20200123091400001-RegisterOrganizationEvent.json"
	data, err := ioutil.ReadFile(fileName)
	assert.NoError(t, err)

	// Unmarshal event
	event, err := EventFromJson(data)
	assert.NoError(t, err)
	assert.Equal(t, RegisterOrganization, event.Type())

	// Unmarshal payload
	payload := &RegisterOrganizationEvent{}
	err = event.Unmarshal(payload)
	assert.NoError(t, err)
	assert.Equal(t, "Zorggroep Nuts", payload.Organization.Name)

	// Marshal event
	assert.Equal(t, data, event.Marshal())
}

func TestUnsupportedEventType(t *testing.T) {
	data := "{}"
	event, err := EventFromJson([]byte(data))
	assert.NoError(t, err)
	event.Type()
}