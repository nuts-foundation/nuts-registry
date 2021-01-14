package pkg

import (
	"github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"time"
)

// DIDStore is the interface for the low level DID operations.
type DIDStore interface {
	// Search searches for DID documents that match the given conditions;
	// - onlyOwn: only return documents which contain a verificationMethod which' private key is present in this node.
	// - tags: only return documents that match ALL of the given tags.
	// If something goes wrong an error is returned.
	Search(onlyOwn bool, tags []string) ([]did.Document, error)
	// Create creates a new DID document and returns it. If something goes wrong an error is returned.
	Create() (*did.Document, error)
	// Get returns the DID document using on the given DID or nil if not found. If something goes wrong an error is returned.
	Get(DID did.DID) (*did.Document, *DIDDocumentMetadata, error)
	// GetByTag gets a DID document using the given tag or nil if not found. If multiple documents match the given tag
	// or something else goes wrong, an error is returned.
	GetByTag(tag string) (*did.Document, *DIDDocumentMetadata, error)
	// Update replaces the DID document identified by DID with the nextVersion if the given hash matches the current valid DID document hash.
	Update(DID did.DID, hash model.Hash, nextVersion did.Document) (*did.Document, error)
	// Tag replaces all tags on a DID document given the DID.
	Tag(DID did.DID, tags []string) error
}

// DIDDocumentMetadata holds the metadata of a DID document
type DIDDocumentMetadata struct {
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated,omitempty"`
	// Version contains the semantic version of the DID document.
	Version int `json:"version"`
	// OriginJWSHash contains the hash of the JWS envelope of the first version of the DID document.
	OriginJWSHash model.Hash `json:"originJwsHash"`
	// Hash of DID document bytes. Is equal to payloadHash in network layer.
	Hash model.Hash `json:"hash"`
	// Tags of the DID document.
	Tags []string `json:"tags,omitempty"`
}

//type StoreWrapper struct {
//	networkClient networkPkg.NetworkClient
//	store         DIDStore
//}
//
//func wrap(store DIDStore) DIDStore {
//	return &StoreWrapper(store: store)
//}
