package model

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestOIDURN_UnmarshalJSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		expected := "urn:oid:1.3.6.1.4.1.54851.2:type"
		input := `"` + expected + `"`
		actual := OIDURNValue{}
		err := json.Unmarshal([]byte(input), &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual.String())
	})
	t.Run("error - invalid URN", func(t *testing.T) {
		input := `"not a URN"`
		actual := OIDURNValue{}
		err := json.Unmarshal([]byte(input), &actual)
		assert.EqualError(t, err, "invalid URN: not a URN")
		assert.Empty(t, actual.String())
	})
	t.Run("error - invalid OID URN (incorrect scheme)", func(t *testing.T) {
		input := `"urn:NotOID:1.3.6.1.4.1.54851.2:foobar"`
		actual := OIDURNValue{}
		err := json.Unmarshal([]byte(input), &actual)
		assert.EqualError(t, err, "invalid OID URN: urn:NotOID:1.3.6.1.4.1.54851.2:foobar")
		assert.Empty(t, actual.String())
	})
	t.Run("error - invalid OID URN (OID invalid)", func(t *testing.T) {
		input := `"urn:oid:1.3.6.1.NotANumber.1.54851.2:foobar"`
		actual := OIDURNValue{}
		err := json.Unmarshal([]byte(input), &actual)
		assert.EqualError(t, err, "invalid OID URN: urn:oid:1.3.6.1.NotANumber.1.54851.2:foobar")
		assert.Empty(t, actual.String())
	})
	t.Run("error - invalid OID URN (no value)", func(t *testing.T) {
		input := `"urn:oid:1.3.6.1.4.1.54851.2"`
		actual := OIDURNValue{}
		err := json.Unmarshal([]byte(input), &actual)
		assert.EqualError(t, err, "invalid OID URN: urn:oid:1.3.6.1.4.1.54851.2")
		assert.Empty(t, actual.String())
	})
}

func TestURL_Marshaling(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		expected := "https://nuts.nl"
		plainURL, _ := url.Parse(expected)
		urlAsBytes, err := json.Marshal(NewURL(*plainURL))

		t.Run("marshaling", func(t *testing.T) {
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, fmt.Sprintf(`"%s"`, expected), string(urlAsBytes))
		})
		t.Run("unmarshaling", func(t *testing.T) {
			var actual URL
			err = json.Unmarshal(urlAsBytes, &actual)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, expected, actual.String())
		})
	})
	t.Run("error - can't unmarshal", func(t *testing.T) {
		var actual URL
		err := json.Unmarshal([]byte(`"not a URL"`), &actual)
		assert.Error(t, err)
		assert.Empty(t, actual.String())
	})
}