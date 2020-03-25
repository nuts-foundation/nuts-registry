package events

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

type testEvent struct {
	unmarshalPostProcCalled bool
}

func (t *testEvent) PostProcessUnmarshal(event Event) error {
	t.unmarshalPostProcCalled = true
	return nil
}

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
		assert.Equal(t, "VendorClaimEvent", string(event.Type()))
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
		r := map[string]interface{}{}
		err = event.Unmarshal(&r)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "Zorggroep Nuts", r["orgName"])
	})
	t.Run("ok - with postprocessor", func(t *testing.T) {
		data := CreateEvent("testEvent", testEvent{}).Marshal()
		event, err := EventFromJSON(data)
		if !assert.NoError(t, err) {
			return
		}
		e := testEvent{}
		err = event.Unmarshal(&e)
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, e.unmarshalPostProcCalled)
	})
	t.Run("error - no payload", func(t *testing.T) {
		event, err := EventFromJSON([]byte("{\"type\": \"RegisterVendorEvent\"}"))
		if !assert.NoError(t, err) {
			return
		}
		payload := map[string]interface{}{}
		err = event.Unmarshal(&payload)
		assert.EqualError(t, err, "event has no payload")
	})
}

func TestCreateEvent(t *testing.T) {
	event := CreateEvent("Foobar", struct{}{})
	assert.Equal(t, "Foobar", string(event.Type()))
	assert.Equal(t, int64(0), time.Now().Unix()-event.IssuedAt().Unix())
}

func TestSignEvent(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		event := CreateEvent("Foobar", struct{}{})
		err := event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return []byte("signature"), nil
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []byte("signature"), event.Signature())
	})
	t.Run("error", func(t *testing.T) {
		event := CreateEvent("Foobar", struct{}{})
		err := event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return nil, errors.New("failed")
		})
		assert.Error(t, err)
	})
}

func readTestEvent() ([]byte, error) {
	return ioutil.ReadFile("../../test_data/valid_files/events/20200123091400002-VendorClaimEvent.json")
}
