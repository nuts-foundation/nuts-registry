/*
 * Nuts registry
 * Copyright (C) 2020. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */
package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

func CreateTestEvent(eventType EventType, payload interface{}, previousEvent Ref, issuedAt time.Time) Event {
	return &jsonEvent{
		EventVersion:  currentEventVersion,
		EventType:     string(eventType),
		PreviousEvent: previousEvent,
		EventIssuedAt: issuedAt,
		EventPayload:  payload,
	}
}

type testEvent struct {
	unmarshalPostProcCalled bool
}

func (t *testEvent) PostProcessUnmarshal(event Event) error {
	t.unmarshalPostProcCalled = true
	return nil
}

func TestEventsFromJSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		singleEvent := `{
 "version": 1,
 "type": "Test",
 "issuedAt": "2020-04-27T10:25:22.861204915+02:00",
 "ref": "b1aa2fd05d040d20fe55152e14c336c0ab9d0e79",
 "prev": "010203",
 "jws": "my.awesome.jws",
 "payload": {
   "Hello": "World"
 }
}`
		input := "[" + singleEvent + ", " + singleEvent + "]"
		events, err := EventsFromJSON([]byte(input))
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, events, 2)
	})
	t.Run("error - unmarshal", func(t *testing.T) {
		events, err := EventsFromJSON([]byte("foobar"))
		assert.Error(t, err)
		assert.Empty(t, events)
	})
}

func TestEventFromJSONWithIssuedAt(t *testing.T) {
	t.Run("ok - issuedAt not present", func(t *testing.T) {
		data, err := readTestEvent()
		if !assert.NoError(t, err) {
			return
		}
		expected := time.Now()
		event, err := EventFromJSONWithIssuedAt(data, expected)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, expected, event.IssuedAt())
	})
	t.Run("ok - issuedAt present", func(t *testing.T) {
		unexpected := time.Date(2020, 10, 20, 20, 30, 0, 0, time.UTC)
		event, err := EventFromJSONWithIssuedAt(CreateEvent("someType", "Hello, World!", nil).Marshal(), unexpected)
		if !assert.NoError(t, err) {
			return
		}
		assert.NotEqual(t, unexpected, event.IssuedAt())
		assert.False(t, event.IssuedAt().IsZero())
	})
	t.Run("error - unable to unmarshal event", func(t *testing.T) {
		event, err := EventFromJSONWithIssuedAt([]byte("not JSON"), time.Now())
		assert.Nil(t, event)
		assert.Error(t, err)
	})
}

func TestEventFromJSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		t.Run("v0", func(t *testing.T) {
			data, err := readTestEvent()
			if !assert.NoError(t, err) {
				return
			}
			event, err := EventFromJSON(data)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, "VendorClaimEvent", string(event.Type()))
			// v0 events have no issuedAt field, so using EventFromJSON should lead to a zero timestamp.
			assert.True(t, event.IssuedAt().IsZero())
		})
		t.Run("v1", func(t *testing.T) {
			input := `{
  "version": 1,
  "type": "Test",
  "issuedAt": "2020-04-27T10:25:22.861204915+02:00",
  "ref": "b1aa2fd05d040d20fe55152e14c336c0ab9d0e79",
  "prev": "010203",
  "jws": "my.awesome.jws",
  "payload": {
    "Hello": "World"
  }
}`
			event, err := EventFromJSON([]byte(input))
			assert.NoError(t, err)
			assert.NotNil(t, event)
		})
	})
	t.Run("error - unmarshalling", func(t *testing.T) {
		event, err := EventFromJSON([]byte("foobar"))
		assert.Error(t, err)
		assert.Nil(t, event)
	})
	t.Run("error - missing event type", func(t *testing.T) {
		_, err := EventFromJSON([]byte("{}"))
		assert.Error(t, err, ErrMissingEventType)
	})
	t.Run("error - invalid ref", func(t *testing.T) {
		input := `{
  "version": 1,
  "type": "Test",
  "issuedAt": "2020-04-27T10:25:22.861204915+02:00",
  "ref": "b1aa2fd05d040d20fe55152e14c336c0ab9d0d79",
  "prev": "010203",
  "jws": "my.awesome.jws",
  "payload": {
    "Hello": "World"
  }
}`
		event, err := EventFromJSON([]byte(input))
		assert.Nil(t, event)
		assert.EqualError(t, err, "event ref is invalid (specified: b1aa2fd05d040d20fe55152e14c336c0ab9d0d79, actual: b1aa2fd05d040d20fe55152e14c336c0ab9d0e79)")
	})
}

func TestRefCalculation(t *testing.T) {
	t.Run("ok - ref from unmarshalled v0 event", func(t *testing.T) {
		eventAsJson, _ := readTestEvent()
		event, _ := EventFromJSON(eventAsJson)
		assert.Equal(t, "eb0268837459c1cb5505fe9032caa50ab297de40", event.Ref().String())
		assert.Equal(t, event.Ref(), event.Ref())
	})
	t.Run("ok - ref changes when included fields change", func(t *testing.T) {
		// This test assumes the event is a flat JSON object (except payload ofc)

		v0Fields := map[string][]interface{}{
			"issuedAt": {time.Now(), time.Unix(0, 0)},
			"jws":      {"", "foobar"},
			"payload":  {struct{}{}},
			"type":     {"foobar"},
		}

		test := func(t *testing.T, event Event, includedFields map[string][]interface{}, excludedFields map[string]interface{}) {
			eventAsMap := map[string]interface{}{}
			json.Unmarshal(event.Marshal(), &eventAsMap)
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
			fmt.Printf("Fields included in Ref:   %v\n", reflect.ValueOf(includedFields).MapKeys())
			fmt.Printf("Fields excluded from Ref: %v\n", reflect.ValueOf(excludedFields).MapKeys())

			calcRefWithMutation := func(field string, value interface{}) (Ref, error) {
				// Copy map to avoid mutating input map
				var m = make(map[string]interface{})
				for k, v := range eventAsMap {
					m[k] = v
				}
				m[field] = value
				data, _ := json.Marshal(m)
				event := jsonEvent{}
				err := json.Unmarshal(data, &event)
				if err != nil {
					return nil, err
				}
				return event.Ref(), nil
			}

			t.Run("included fields", func(t *testing.T) {
				// Test that Ref changes when included field changes
				for field, mutations := range includedFields {
					for _, mutation := range mutations {
						t.Run(field, func(t *testing.T) {
							actual, err := calcRefWithMutation(field, mutation)
							if !assert.NoErrorf(t, err, "error testing mutation %s=%v", field, mutation) {
								return
							}
							assert.NotEqualf(t, event.Ref(), actual, "ref didn't change when %s=%v", field, mutation)
						})
					}
				}
			})
			t.Run("excluded fields", func(t *testing.T) {
				// Test that Ref does not change when excluded field changes
				for field, mutation := range excludedFields {
					t.Run(field, func(t *testing.T) {
						actual, err := calcRefWithMutation(field, mutation)
						if !assert.NoErrorf(t, err, "error testing mutation %s=%v", field, mutation) {
							return
						}
						assert.Equalf(t, event.Ref(), actual, "ref changed when %s=%v", field, mutation)
					})
				}
			})
		}
		t.Run("event version = 0", func(t *testing.T) {
			event := CreateEvent("Test", map[string]interface{}{"Hello": "World"}, nil)
			(event.(*jsonEvent)).EventVersion = 0
			event.Sign(func([]byte) ([]byte, error) {
				return []byte("my.awesome.jws"), nil
			})
			excludedFields := map[string]interface{}{
				"ref":     Ref([]byte{4, 5, 6}).String(),
				"version": -1,
			}
			test(t, event, v0Fields, excludedFields)
		})
		t.Run("event version = 1", func(t *testing.T) {
			event := CreateEvent("Test", map[string]string{"Hello": "World"}, []byte{1, 2, 3})
			(event.(*jsonEvent)).EventVersion = 1
			event.Sign(func([]byte) ([]byte, error) {
				return []byte("my.awesome.jws"), nil
			})
			includedFields := make(map[string][]interface{}, 0)
			for k, v := range v0Fields {
				includedFields[k] = v
			}
			includedFields["version"] = []interface{}{0, 2}
			includedFields["prev"] = []interface{}{nil, Ref([]byte{3, 2, 1}).String()}
			excludedFields := map[string]interface{}{
				"ref": Ref([]byte{4, 5, 6}).String(),
			}
			test(t, event, includedFields, excludedFields)
		})
	})
	t.Run("test canonicalization", func(t *testing.T) {
		test := func(expectedOutput, input string) {
			bytes, err := canonicalizeJSON([]byte(input))
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, expectedOutput, string(bytes))
		}
		test("{\"Array\":[1,2,3],\"Bsort\":true,\"Hello\":\"World\",\"Nested\":{\"Bool\":true,\"Scaler\":1.213123123,\"Scientic\":100}}", `{
"Hello": "World",

"Nested": {
    "Scaler": 1.213123123,
  "Scientic": 1E2,
"Bool": true
},
"Bsort": true,
"Array": [ 1, 2, 3 ]
}`)
		test(`{}`, `{   }`)
	})
}

func TestMarshalEvent(t *testing.T) {
	t.Run("marshal v0 event", func(t *testing.T) {
		expected, _ := readTestEvent()
		event, _ := EventFromJSON(expected)
		m := map[string]interface{}{}
		marshalled := event.Marshal()
		json.Unmarshal(marshalled, &m)
		// IssuedAt is not in the source JSON, so remove it before comparison
		delete(m, "issuedAt")
		// version and ref are >= v1
		delete(m, "version")
		delete(m, "ref")
		assert.JSONEq(t, string(expected), toJSON(m))

		t.Run("can be unmarshalled", func(t *testing.T) {
			actual, err := EventFromJSON(marshalled)
			if !assert.NoError(t, err) {
				return
			}
			(actual.(*jsonEvent)).ThisEventRef = nil
			assert.Equal(t, event, actual)
			assert.Nil(t, event.PreviousRef())
		})
	})
	t.Run("marshal v1 event", func(t *testing.T) {
		payload := map[string]interface{}{"Hello": "World"}
		event := CreateEvent("v1", payload, []byte{1, 2, 3})
		l, _ := time.LoadLocation("UTC")
		(event.(*jsonEvent)).EventIssuedAt = time.Date(1970, 1, 1, 0, 0, 0, 0, l)

		expected := `{
	"issuedAt": ` + toJSON(event.IssuedAt()) + `,
	"prev":		"010203",
	"ref":		"da46d3559e472e5c7fde9d7793f2a2c5626b9f75",
	"type":		"v1",
	"version":	1,
	"payload": 	{"Hello": "World"}
}`
		marshalled := event.Marshal()
		assert.JSONEq(t, expected, string(marshalled))

		t.Run("can be unmarshalled", func(t *testing.T) {
			actual, err := EventFromJSON(marshalled)
			if !assert.NoError(t, err) {
				return
			}
			eventAsJsonEvent := actual.(*jsonEvent)
			eventAsJsonEvent.ThisEventRef = nil
			eventAsJsonEvent.cachedData = nil
			assert.Equal(t, event, actual)
			assert.NotNil(t, event.PreviousRef())
		})
	})
	t.Run("marshalling read event should return read event unchanged", func(t *testing.T) {
		// We added some random whitespace and extra fields
		expected := `{
	"issuedAt": ` + toJSON(time.Unix(5000, 0).UTC()) + `,
	"prev":		"010203",


	"ref":		"f06ec8d10533e57c27d27936adff5ebbd66ec226",
	"type":		"v1",
	"version":	1,
	"payload": 	{"Hello": "World"},
	"extra-field": "some-value"
}`
		event, err := EventFromJSON([]byte(expected))
		if !assert.NoError(t, err) {
			return
		}
		marshalled := event.Marshal()
		assert.JSONEq(t, expected, string(marshalled))
		assert.Equal(t, expected, string(marshalled))
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

func TestRef_Equal(t *testing.T) {
	assert.True(t, Ref([]byte{1, 2, 3}).Equal([]byte{1, 2, 3}))
	assert.False(t, Ref([]byte{1, 2, 3}).Equal([]byte{1, 2, 3, 4}))
	assert.False(t, Ref(nil).Equal([]byte{1, 2, 3, 4}))
	assert.True(t, Ref(nil).Equal(Ref(nil)))
}

func readTestEvent() ([]byte, error) {
	return ioutil.ReadFile("../../test_data/valid_files/events/20200123091400002-VendorClaimEvent.json")
}

func toJSON(input interface{}) string {
	data, _ := json.Marshal(input)
	return string(data)
}
