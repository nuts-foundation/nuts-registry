package events

import (
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
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

func TestCreateEvent(t *testing.T) {
	event, err := CreateEvent(RegisterOrganization, RegisterOrganizationEvent{db.Organization{Name: "bla"}})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, RegisterOrganization, event.Type())
	assert.Equal(t, int64(0), time.Now().Unix() - event.IssuedAt().Unix())
}

func TestIsEventType(t *testing.T) {
	assert.True(t, IsEventType(RegisterOrganization))
	assert.False(t, IsEventType("NonExistingEvent"))
}

func readTestEvent() ([]byte, error) {
	return ioutil.ReadFile("../../test_data/valid_files/events/20200123091400001-RegisterOrganizationEvent.json")
}
