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

type URL struct {
	url.URL
}

func NewURL(url url.URL) URL {
	return URL{url}
}

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

func (U URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(U.String())
}

// OIDURNValue is a URN in the OID scheme in the format of urn:oid:1.2.3.4:some-value
type OIDURNValue struct {
	urn   urn.URN
	oid   string
	value string
}

func NewOIDURN(oid string, value string) (OIDURNValue, error) {
	for _, idPart := range strings.Split(oid, ".") {
		for _, idPartRune := range idPart {
			if !unicode.IsNumber(idPartRune) {
				return OIDURNValue{}, fmt.Errorf("OID is invalid: %s", oid)
			}
		}
	}
	urn, _ := urn.Parse([]byte(fmt.Sprintf("urn:oid:%s:%s", oid, value)))
	return OIDURNValue{
		urn:   *urn,
		oid:   oid,
		value: value,
	}, nil
}

func (e OIDURNValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.urn.String())
}

func (e *OIDURNValue) UnmarshalJSON(bytes []byte) error {
	var urn urn.URN
	if err := json.Unmarshal(bytes, &urn); err != nil {
		return err
	}
	errInvalid := fmt.Errorf(errInvalidOIDURNFmt, urn.String())
	if urn.ID != oidURNNamespace {
		return errInvalid
	}
	idAndValue := urn.SS
	sepIdx := strings.Index(idAndValue, ":")
	if sepIdx == -1 {
		return errInvalid
	}
	id := idAndValue[:sepIdx]
	// Assert OID type is valid
	for _, idPart := range strings.Split(id, ".") {
		for _, idPartRune := range idPart {
			if !unicode.IsNumber(idPartRune) {
				return errInvalid
			}
		}
	}
	e.urn = urn
	e.oid = id
	e.value = idAndValue[sepIdx+1:]
	return nil
}

func (e OIDURNValue) String() string {
	return e.urn.String()
}
