package events

import (
	"encoding/json"
	"errors"
	"fmt"
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

func TestRefCalculation(t *testing.T) {
	t.Run("ok - ref from unmarshalled v0 event", func(t *testing.T) {
		eventAsJson, _ := readTestEvent()
		event, _ := EventFromJSON(eventAsJson)
		assert.Equal(t, "7e3f9c1eef93d6685d029c058d50575b8a250960", event.Ref().String())
		assert.Equal(t, event.Ref(), event.Ref())
	})
	t.Run("ok - ref changes when included fields change", func(t *testing.T) {
		// This test assumes the event is a flat JSON object (except payload ofc)
		event := CreateEvent("Test", map[string]string{"Hello": "World"}, []byte{1, 2, 3})
		event.Sign(func([]byte) ([]byte, error) {
			return []byte("my.awesome.jws"), nil
		})
		eventAsMap := map[string]interface{}{}
		json.Unmarshal(event.Marshal(), &eventAsMap)

		includedFields := map[string][]interface{}{
			// Values are mutations we'll be testing
			"issuedAt": {time.Now(), time.Unix(0, 0)},
			"version":  {0, -1, -2},
			"prev":     {nil, Ref([]byte{3, 2, 1}).String()},
			"jws":      {"", "foobar"},
			"payload":  {struct{}{}},
			"type":     {"foobar"},
		}
		excludedFields := map[string]interface{}{
			"ref": Ref([]byte{4,5,6}).String(),
		}

		// Assert fields = includedFields ∪ excludedFields
		for field, _ := range eventAsMap {
			var included = false
			for f := range includedFields {
				if field == f {
					included = true
					break
				}
			}
			for f := range excludedFields {
				if field == f {
					included = true
					break
				}
			}
			if !included {
				assert.FailNow(t, "field is not in included or excluded set: "+field)
				return
			}
		}
		fmt.Printf("Fields included in Ref:   %v\n", includedFields)
		fmt.Printf("Fields excluded from Ref: %v\n", excludedFields)

		calcRefWithMutation := func(field string, value interface{}) (Ref, error) {
			// Copy map to avoid mutating input map
			var m = make(map[string]interface{})
			for k, v := range eventAsMap {
				m[k] = v
			}
			m[field] = value
			data, _ := json.Marshal(m)
			event, err := EventFromJSON(data)
			if err != nil {
				return nil, err
			}
			return event.Ref(), nil
		}

		// Test that Ref changes when included field changes
		for field, mutations := range includedFields {
			for _, mutation := range mutations {
				actual, err := calcRefWithMutation(field, mutation)
				if !assert.NoErrorf(t, err, "error testing mutation %s=%v", field, mutation) {
					return
				}
				assert.NotEqualf(t, event.Ref(), actual, "ref didn't change when %s=%v", field, mutation)
			}
		}

		// Test that Ref does not change when excluded field changes
		for field, mutation := range excludedFields {
			actual, err := calcRefWithMutation(field, mutation)
			if !assert.NoErrorf(t, err, "error testing mutation %s=%v", field, mutation) {
				return
			}
			assert.Equalf(t, event.Ref(), actual, "ref changed when %s=%v", field, mutation)
		}
	})
	t.Run("ok - test canonicalization invariants", func(t *testing.T) {
		// TODO when implemented!
	})
}

func TestMarshalEvent(t *testing.T) {
	t.Run("marshal v0 event", func(t *testing.T) {
		expected, _ := readTestEvent()
		event, _ := EventFromJSON(expected)
		m := map[string]interface{}{}
		json.Unmarshal(event.Marshal(), &m)
		// IssuedAt is not in the source JSON, so remove it before comparison
		delete(m, "issuedAt")
		// version and ref are >= v1
		delete(m, "version")
		delete(m, "ref")
		assert.JSONEq(t, string(expected), toJSON(m))
	})
	t.Run("marshal v1 event", func(t *testing.T) {
		payload := map[string]string{"Hello": "World"}
		event := CreateEvent("v1", payload, []byte{1, 2, 3})
		(event.(*jsonEvent)).EventIssuedAt = time.Date(1970,1,1,0,0,0,0, time.FixedZone("UTC", 0))

		expected := `{
	"issuedAt": ` + toJSON(event.IssuedAt()) + `,
	"prev":		"010203",
	"ref":		"da46d3559e472e5c7fde9d7793f2a2c5626b9f75",
	"type":		"v1",
	"version":	1,
	"payload": 	{"Hello": "World"}
}`
		assert.JSONEq(t, expected, string(event.Marshal()))
	})
}

func TestUnmarshalJSONPayload(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		data, err := readTestEvent()
		if !assert.NoError(t, err) {
			return
		}
		event, _ := EventFromJSON(data)
		// Event without version
		assert.Equal(t, Version(0), event.Version())
		payload := map[string]interface{}{}
		err = event.Unmarshal(&payload)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "Zorggroep Nuts", payload["orgName"])
	})
	t.Run("ok - with postprocessor", func(t *testing.T) {
		data := CreateEvent("testEvent", testEvent{}, nil).Marshal()
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
	event := CreateEvent("Foobar", struct{}{}, nil)
	assert.True(t, event.IssuedAt().Unix() > int64(0), "incorrect issuedAt")
	assert.Equal(t, "Foobar", string(event.Type()), "incorrect event type")
	assert.Equal(t, currentEventVersion, event.Version(), "incorrect event version")
}

func TestSignEvent(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		event := CreateEvent("Foobar", struct{}{}, nil)
		err := event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return []byte("signature"), nil
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []byte("signature"), event.Signature())
	})
	t.Run("ok - no signature", func(t *testing.T) {
		event := CreateEvent("Foobar", struct{}{}, nil)
		err := event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return nil, nil
		})
		if !assert.NoError(t, err) {
			return
		}
		assert.Empty(t, event.Signature())
	})
	t.Run("error", func(t *testing.T) {
		event := CreateEvent("Foobar", struct{}{}, nil)
		err := event.Sign(func(bytes2 []byte) (bytes []byte, err error) {
			return nil, errors.New("failed")
		})
		assert.Error(t, err)
	})
}

func TestRef_IsZero(t *testing.T) {
	assert.False(t, Ref([]byte{1, 2, 3}).IsZero())
	assert.True(t, Ref(nil).IsZero())
	assert.True(t, Ref([]byte{}).IsZero())
	assert.False(t, Ref([]byte{0}).IsZero())
}

func TestRef_String(t *testing.T) {
	assert.Equal(t, "c8c9ca", Ref([]byte{200, 201, 202}).String())
}

func TestRef_Marshal(t *testing.T) {
	t.Run("ok - roundtrip", func(t *testing.T) {
		expected := Ref([]byte{1, 2, 3})
		refAsJSON, err := expected.MarshalJSON()
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "\"010203\"", string(refAsJSON))
		actual := Ref{}
		err = actual.UnmarshalJSON(refAsJSON)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, expected, actual)
	})
}

func TestRef_Unmarshal(t *testing.T) {
	t.Run("error - invalid json", func(t *testing.T) {
		r := Ref{}
		err := r.UnmarshalJSON([]byte{1, 2, 3})
		assert.EqualError(t, err, "invalid character '\\x01' looking for beginning of value")
	})
	t.Run("error - invalid hex value", func(t *testing.T) {
		r := Ref{}
		err := r.UnmarshalJSON([]byte("\"foobar\""))
		assert.EqualError(t, err, "encoding/hex: invalid byte: U+006F 'o'")
	})
}

func readTestEvent() ([]byte, error) {
	return ioutil.ReadFile("../../test_data/valid_files/events/20200123091400002-VendorClaimEvent.json")
}

func toJSON(input interface{}) string {
	data, _ := json.Marshal(input)
	return string(data)
}
