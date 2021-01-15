package storage

import (
	"crypto/ecdsa"
	"errors"
	"github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/nuts-registry/pkg"
)

// TODO: Keep track of different versions when DID Documents are updated
func NewMemoryDIDStore() pkg.DIDStore {
	return &memoryDIDStore{
		store: map[string]memoryEntry{},
	}
}

type memoryDIDStore struct {
	store map[string]memoryEntry
}

type memoryEntry struct {
	document   did.Document
	tags       map[string]bool
	privateKey *ecdsa.PrivateKey
	metadata   pkg.DIDDocumentMetadata
}

func (e memoryEntry) HasTags(tags []string) bool {
	for _, tag := range tags {
		if !e.tags[tag] {
			return false
		}
	}
	return true
}

func (m memoryDIDStore) Search(onlyOwn bool, tags []string) ([]did.Document, error) {
	result := make([]did.Document, 0)
	for _, entry := range m.store {
		if onlyOwn && entry.privateKey == nil {
			continue
		}
		if !entry.HasTags(tags) {
			continue
		}
		result = append(result, entry.document)
	}
	return result, nil
}

func (m memoryDIDStore) Add(document did.Document, metadata pkg.DIDDocumentMetadata) error {
	m.store[document.ID.String()] = memoryEntry{
		document: document,
		metadata: metadata,
	}
	return nil
}

func (m memoryDIDStore) Get(DID did.DID) (*did.Document, *pkg.DIDDocumentMetadata, error) {
	entry, found := m.store[DID.String()]
	if !found {
		return nil, nil, nil
	}
	return &entry.document, &entry.metadata, nil
}

func (m memoryDIDStore) GetByTag(tag string) (*did.Document, *pkg.DIDDocumentMetadata, error) {
	if tag == "" {
		return nil, nil, nil
	}
	results := make([]memoryEntry, 0)
	for _, entry := range m.store {
		if entry.HasTags([]string{tag}) {
			if len(results) == 1 {
				return nil, nil, errors.New("multiple DIDs match the given tag")
			}
			results = append(results, entry)
		}
	}
	if len(results) == 0 {
		return nil, nil, nil
	} else {
		return &results[0].document, &results[0].metadata, nil
	}
}

func (m memoryDIDStore) Tag(DID did.DID, tags []string) error {
	entry, found := m.store[DID.String()]
	if !found {
		return errors.New("DID not found")
	}
	entry.tags = make(map[string]bool, len(tags))
	for _, tag := range tags {
		entry.tags[tag] = true
	}
	return nil
}
