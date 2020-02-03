package events

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestEventFromJSON(t *testing.T) {
	data, err := readTestEvent()
	if !assert.NoError(t, err) {
		return
	}
	event, err := EventFromJSON(data)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, RegisterOrganization, event.Type())
}

func TestMarshalEvent(t *testing.T) {
	data, err := readTestEvent()
	if !assert.NoError(t, err) {
		return
	}
	event := jsonEvent{
		data: data,
	}
	assert.Equal(t, data, event.Marshal())
}

func TestUnmarshalJSONPayload(t *testing.T) {
	data, err := readTestEvent()
	if !assert.NoError(t, err) {
		return
	}
	event := jsonEvent{
		data: data,
	}

	r := RegisterOrganizationEvent{}
	err = event.Unmarshal(&r)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Zorggroep Nuts", r.Organization.Name)
}

func TestUnmarshalEventWithoutPayload(t *testing.T) {
	event, err := EventFromJSON([]byte("{\"type\": \"" + RegisterOrganization + "\"}"))
	if !assert.NoError(t, err) {
		return
	}
	payload := RegisterOrganizationEvent{}
	err = event.Unmarshal(&payload)
	assert.EqualError(t, err, "event has no payload")
}

func TestMissingEventType(t *testing.T) {
	_, err := EventFromJSON([]byte("{}"))
	assert.Error(t, err, ErrMissingEventType)
}

func readTestEvent() ([]byte, error) {
	return ioutil.ReadFile("../../test_data/valid_files/events/20200123091400001-RegisterOrganizationEvent.json")
}
