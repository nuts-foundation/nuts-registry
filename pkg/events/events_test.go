package events

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

func TestEventFromJSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		data, err := readTestEvent()
		if !assert.NoError(t, err) {
			return
		}
		event, err := EventFromJSON(data)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, VendorClaim, event.Type())
	})
	t.Run("error - missing event type", func(t *testing.T) {
		_, err := EventFromJSON([]byte("{}"))
		assert.Error(t, err, ErrMissingEventType)
	})
}

func TestMarshalEvent(t *testing.T) {
	expected, _ := readTestEvent()
	event, _ := EventFromJSON(expected)
	// IssuedAt is not in the source JSON, so remove it before comparison
	m := map[string]interface{}{}
	json.Unmarshal(event.Marshal(), &m)
	delete(m, "issuedAt")
	actual, _ := json.Marshal(m)
	assert.JSONEq(t, string(expected), string(actual))
}

func TestUnmarshalJSONPayload(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
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
	})
	t.Run("error - no payload", func(t *testing.T) {
		event, err := EventFromJSON([]byte("{\"type\": \"" + RegisterVendor + "\"}"))
		if !assert.NoError(t, err) {
			return
		}
		payload := RegisterVendorEvent{}
		err = event.Unmarshal(&payload)
		assert.EqualError(t, err, "event has no payload")
	})
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
