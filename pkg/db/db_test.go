/*
 * Nuts registry
 * Copyright (C) 2019. Nuts community
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

package db

import (
	"encoding/json"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrganization_KeysAsSet(t *testing.T) {
	valid := "{\"name\": \"Zorggroep Nuts 2\",\"identifier\": \"urn:oid:2.16.840.1.113883.2.4.6.1:00000001\",\"keys\": [{\"kty\":\"EC\",\"crv\":\"P-256\",\"x\":\"MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4\",\"y\":\"4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM\",\"use\":\"enc\",\"kid\":\"2\"}]}"
	invalidKeys := "{\"name\": \"Zorggroep Nuts 2\",\"identifier\": \"urn:oid:2.16.840.1.113883.2.4.6.1:00000001\",\"keys\": {\"kty\":\"EC\"}}"
	invalidKty := "{\"name\": \"Zorggroep Nuts 2\",\"identifier\": \"urn:oid:2.16.840.1.113883.2.4.6.1:00000001\",\"keys\": [{\"kty\":\"error\"}]}"
	invalidData := "{\"name\": \"Zorggroep Nuts 2\",\"identifier\": \"urn:oid:2.16.840.1.113883.2.4.6.1:00000001\",\"keys\": [{\"kty\":\"EC\",\"crv\":\"P-256\",\"x\":\"MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4\",\"z\":\"4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM\",\"use\":\"enc\",\"kid\":\"2\"}]}"

	t.Run("No keys returns empty set", func(t *testing.T) {
		o := Organization{}
		set, err := o.KeysAsSet()
		if assert.NoError(t, err) && assert.NotNil(t, set) {
			assert.Len(t, set.Keys, 0)
		}
	})

	t.Run("JWK set is parsed", func(t *testing.T) {
		o := Organization{}
		if assert.NoError(t, json.Unmarshal([]byte(valid), &o)) {
			set, err := o.KeysAsSet()
			if assert.NoError(t, err) && assert.NotNil(t, set) {
				assert.Len(t, set.Keys, 1)
				assert.Equal(t, jwa.EC, set.Keys[0].KeyType())
			}
		}
	})

	t.Run("invalid JWK set in json returns error", func(t *testing.T) {
		o := Organization{}
		assert.Error(t, json.Unmarshal([]byte(invalidKeys), &o))
	})

	t.Run("invalid contents in JWK set returns error", func(t *testing.T) {
		o := Organization{}
		if assert.NoError(t, json.Unmarshal([]byte(invalidKty), &o)) {
			set, err := o.KeysAsSet()
			assert.Error(t, err)
			assert.Nil(t, set)
		}
	})

	t.Run("invalid combination in JWK set returns error", func(t *testing.T) {
		o := Organization{}
		if assert.NoError(t, json.Unmarshal([]byte(invalidData), &o)) {
			set, err := o.KeysAsSet()
			assert.Error(t, err)
			assert.Nil(t, set)
		}
	})
}

func TestOrganization_CurrentPublicKey(t *testing.T) {
	oldKey := "{\"name\": \"Zorggroep Nuts 2\",\"identifier\": \"urn:oid:2.16.840.1.113883.2.4.6.1:00000001\",\"actors\": [{\"identifier\": \"urn:oid:2.16.840.1.113883.2.4.6.1:00000007\"}],\"publicKey\": \"-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9wJQN59PYsvIsTrFuTqS\\nLoUBgwdRfpJxOa5L8nOALxNk41MlAg7xnPbvnYrOHFucfWBTDOMTKBMSmD4WDkaF\\ndVrXAML61z85Le8qsXfX6f7TbKMDm2u1O3cye+KdJe8zclK9sTFzSD0PP0wfw7wf\\nlACe+PfwQgeOLPUWHaR6aDfaA64QEdfIzk/IL3S595ixaEn0huxMHgXFX35Vok+o\\nQdbnclSTo6HUinkqsHUu/hGHApkE3UfT6GD6SaLiB9G4rAhlrDQ71ai872t4FfoK\\n7skhe8sP2DstzAQRMf9FcetrNeTxNL7Zt4F/qKm80cchRZiFYPMCYyjQphyBCoJf\\n0wIDAQAB\\n-----END PUBLIC KEY-----\"}"

	newKey := "{\"name\": \"Zorggroep Nuts 2\",\"identifier\": \"urn:oid:2.16.840.1.113883.2.4.6.1:00000001\",\"actors\": [{\"identifier\": \"urn:oid:2.16.840.1.113883.2.4.6.1:00000007\"}],\"publicKey\": \"-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9wJQN59PYsvIsTrFuTqS\\nLoUBgwdRfpJxOa5L8nOALxNk41MlAg7xnPbvnYrOHFucfWBTDOMTKBMSmD4WDkaF\\ndVrXAML61z85Le8qsXfX6f7TbKMDm2u1O3cye+KdJe8zclK9sTFzSD0PP0wfw7wf\\nlACe+PfwQgeOLPUWHaR6aDfaA64QEdfIzk/IL3S595ixaEn0huxMHgXFX35Vok+o\\nQdbnclSTo6HUinkqsHUu/hGHApkE3UfT6GD6SaLiB9G4rAhlrDQ71ai872t4FfoK\\n7skhe8sP2DstzAQRMf9FcetrNeTxNL7Zt4F/qKm80cchRZiFYPMCYyjQphyBCoJf\\n0wIDAQAB\\n-----END PUBLIC KEY-----\",\"keys\": [{\"kty\":\"EC\",\"crv\":\"P-256\",\"x\":\"MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4\",\"y\":\"4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM\",\"use\":\"enc\",\"kid\":\"2\"}]}"
	
	t.Run("using old public key", func(t *testing.T) {
		o := Organization{}
		if assert.NoError(t, json.Unmarshal([]byte(oldKey), &o)) {
			key, err := o.CurrentPublicKey()
			if assert.NoError(t, err) {
				assert.Equal(t, jwa.RSA, key.KeyType())
			}
		}
	})

	t.Run("using new JWK style keys", func(t *testing.T) {
		o := Organization{}
		if assert.NoError(t, json.Unmarshal([]byte(newKey), &o)) {
			key, err := o.CurrentPublicKey()
			if assert.NoError(t, err) {
				assert.Equal(t, jwa.EC, key.KeyType())
			}
		}
	})
}
