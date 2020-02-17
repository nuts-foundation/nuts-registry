package events

import (
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
	assert.Equal(t, VendorClaim, event.Type())
}

func TestMarshalEvent(t *testing.T) {
	data, err := readTestEvent()
	if !assert.NoError(t, err) {
		return
	}
	event, err := EventFromJSON(data)
	assert.JSONEq(t, string(data), string(event.Marshal()))
}

func TestUnmarshalJSONPayload(t *testing.T) {
	data, err := readTestEvent()
	if !assert.NoError(t, err) {
		return
	}
	event, _ := EventFromJSON(data)
	r := VendorClaimEvent{}
	err = event.Unmarshal(&r)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Zorggroep Nuts", r.OrgName)
}

func TestUnmarshalEventWithoutPayload(t *testing.T) {
	event, err := EventFromJSON([]byte("{\"type\": \"" + RegisterVendor + "\"}"))
	if !assert.NoError(t, err) {
		return
	}
	payload := RegisterVendorEvent{}
	err = event.Unmarshal(&payload)
	assert.EqualError(t, err, "event has no payload")
}

func TestMissingEventType(t *testing.T) {
	_, err := EventFromJSON([]byte("{}"))
	assert.Error(t, err, ErrMissingEventType)
}

func TestCreateEvent(t *testing.T) {
	event := CreateEvent(RegisterVendor, RegisterVendorEvent{Name: "bla"})
	assert.Equal(t, RegisterVendor, event.Type())
	assert.Equal(t, int64(0), time.Now().Unix()-event.IssuedAt().Unix())
}

func TestIsEventType(t *testing.T) {
	assert.True(t, IsEventType(VendorClaim))
	assert.False(t, IsEventType("NonExistingEvent"))
}

func readTestEvent() ([]byte, error) {
	return ioutil.ReadFile("../../test_data/valid_files/events/20200123091400002-VendorClaimEvent.json")
}