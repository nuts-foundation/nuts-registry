package pkg

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/nuts-foundation/go-did"
	"github.com/shengdoushi/base58"
	"net/url"
)

func jwkToVerificationMethod(DID did.DID, key jwk.Key) (*did.VerificationMethod, error) {
	publicKeyAsJWKAsMap, err := key.AsMap(context.Background())
	if err != nil {
		return nil, err
	}
	DID.Fragment = key.KeyID()
	id, err := url.Parse(DID.String())
	if err != nil {
		return nil, err
	}
	return &did.VerificationMethod{
		ID:           did.URI{URL: *id},
		Type:         did.JsonWebKey2020,
		PublicKeyJwk: publicKeyAsJWKAsMap,
	}, nil
}

func ecdsaPublicKeyToNutsDID(publicKey ecdsa.PublicKey) did.DID {
	if result, err := did.ParseDID("nuts:did:" + ecdsaPublicKeyToIDString(publicKey)); err != nil {
		panic(err)
	} else {
		return *result
	}
}

func ecdsaPublicKeyToIDString(publicKey ecdsa.PublicKey) string {
	pkBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	pkHash := sha256.Sum256(pkBytes)
	return base58.Encode(pkHash[:], base58.BitcoinAlphabet)
}
