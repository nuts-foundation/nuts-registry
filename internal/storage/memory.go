package storage

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/nuts-foundation/go-did"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/shengdoushi/base58"
)

func NewMemoryDIDStore() pkg.DIDStore {
	return &memoryDIDStore{
		store: map[string]memoryDIDStoreEntry{},
	}
}

type memoryDIDStore struct {
	store map[string]memoryDIDStoreEntry
}

type memoryDIDStoreEntry struct {
	document   did.Document
	tags       []string
	privateKey *ecdsa.PrivateKey
}

func (m memoryDIDStore) Search(onlyOwn bool, tags []string) ([]did.Document, error) {
	panic("implement me")
}

func (m memoryDIDStore) Create() (*did.Document, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.PublicKey
	m.ecdsaPublicKeyToIDString(publicKey)

	did.Document{
		Context:            []string{did.DIDContextV1},
		ID:                 did.DID{},
		Controller:         nil,
		VerificationMethod: nil,
		Authentication:     nil,
		AssertionMethod:    nil,
		Service:            nil,
	}
}

func (m memoryDIDStore) ecdsaPublicKeyToIDString(publicKey ecdsa.PublicKey) string {
	pkBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	pkHash := sha256.Sum256(pkBytes)
	return base58.Encode(pkHash[:], base58.BitcoinAlphabet)
}

func (m memoryDIDStore) Get(DID did.DID) (*did.Document, *pkg.DIDDocumentMetadata, error) {
	panic("implement me")
}

func (m memoryDIDStore) GetByTag(tag string) (*did.Document, *pkg.DIDDocumentMetadata, error) {
	panic("implement me")
}

func (m memoryDIDStore) Update(DID did.DID, hash model.Hash, nextVersion did.Document) (*did.Document, error) {
	panic("implement me")
}

func (m memoryDIDStore) Tag(DID did.DID, tags []string) error {
	panic("implement me")
}
