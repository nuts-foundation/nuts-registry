package model

import (
	"encoding/json"
	"fmt"
	urn "github.com/leodido/go-urn"
	"net/url"
	"strings"
	"unicode"
)

const oidURNNamespace = "oid"

const errInvalidOIDURNFmt = "invalid OID URN: %s"

// URL wraps url.URL to aid JSON marshaling.
type URL struct {
	url.URL
}

// NewURL constructs a new URL wrapped around url.URL.
func NewURL(value url.URL) URL {
	return URL{value}
}

// UnmarshalJSON unmarshals a JSON URL (e.g. "https://nuts.nl"). If the URL can't be parsed or isn't abosolute an error is returned.
func (U *URL) UnmarshalJSON(bytes []byte) error {
	var str string
	if err := json.Unmarshal(bytes, &str); err != nil {
		return err
	}
	if value, err := url.Parse(str); err != nil {
		return err
	} else {
		if !value.IsAbs() {
			return fmt.Errorf("URL is not absolute: %s", str)
		}
		*U = URL{
			URL: *value,
		}
	}
	return nil
}

// MarshalJSON marshals the URL into a JSON string.
func (U URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(U.String())
}

// OIDURNValue is a URN in the OID scheme in the format of `urn:oid:(some-oid):(some-value)`
type OIDURNValue struct {
	urn   urn.URN
	oid   string
	value string
}

// NewOIDURNValue constructs a new OIDURNValue given the OID (e.g. `1.2.3.4`) and value (e.g. `some-value`).
func NewOIDURNValue(oid string, value string) (OIDURNValue, error) {
	for _, idPart := range strings.Split(oid, ".") {
		for _, idPartRune := range idPart {
			if !unicode.IsNumber(idPartRune) {
				return OIDURNValue{}, fmt.Errorf("OID is invalid: %s", oid)
			}
		}
	}
	underlyingURN, _ := urn.Parse([]byte(fmt.Sprintf("urn:oid:%s:%s", oid, value)))
	return OIDURNValue{
		urn:   *underlyingURN,
		oid:   oid,
		value: value,
	}, nil
}

// MarshalJSON marshals the OIDURNValue into a JSON string.
func (e OIDURNValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.urn.String())
}

// UnmarshalJSON unmarshals a JSON string into a OIDURNValue.
// An error is returned if the format isn't `urn:oid:(some-oid):(some-value)`
func (e *OIDURNValue) UnmarshalJSON(bytes []byte) error {
	var underlyingURN urn.URN
	if err := json.Unmarshal(bytes, &underlyingURN); err != nil {
		return err
	}
	errInvalid := fmt.Errorf(errInvalidOIDURNFmt, underlyingURN.String())
	if underlyingURN.ID != oidURNNamespace {
		return errInvalid
	}
	idAndValue := underlyingURN.SS
	sepIdx := strings.Index(idAndValue, ":")
	if sepIdx == -1 {
		return errInvalid
	}
	id := idAndValue[:sepIdx]
	if value, err := NewOIDURNValue(id, idAndValue[sepIdx+1:]); err != nil {
		return err
	} else {
		*e = value
		return nil
	}
}

// String formats the OIDURNValue as a string.
func (e OIDURNValue) String() string {
	return e.urn.String()
}
