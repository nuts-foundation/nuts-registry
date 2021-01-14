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

package api

import (
	"encoding/json"
	"fmt"
	"github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/nuts-registry/pkg"
)

func (d *DIDDocumentMetadata) FromModel(input pkg.DIDDocumentMetadata) {
	d.Created = input.Created
	if !input.Updated.IsZero() {
		d.Updated = &input.Updated
	}
	d.Hash = input.Hash.String()
	d.OriginJwsHash = input.OriginJWSHash.String()
	d.Version = input.Version
}

func (d *DIDDocument) FromModel(input did.Document) {
	asJSON, _ := json.Marshal(input)
	_ = json.Unmarshal(asJSON, d)
}

func (d DIDDocument) ToModel() (*did.Document, error) {
	var result did.Document
	if asJSON, err := json.Marshal(d); err != nil {
		return nil, err
	} else if err = json.Unmarshal(asJSON, &result); err != nil {
		return nil, fmt.Errorf("DID document is invalid: %w", err)
	} else {
		return &result, nil
	}
}

func (d *DIDResolutionResult) FromModel(document did.Document, metadata pkg.DIDDocumentMetadata) DIDResolutionResult {
	d.Document = &DIDDocument{}
	d.Document.FromModel(document)
	d.DocumentMetadata = &DIDDocumentMetadata{}
	d.DocumentMetadata.FromModel(metadata)
	// TODO: Resolution metadata
	return *d
}
