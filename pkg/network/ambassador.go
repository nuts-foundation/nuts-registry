package network

import (
	"github.com/labstack/gommon/log"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	network "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	errors2 "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

const documentType = "nuts.registry-event"

// Ambassador acts as integration point between the registry and network by sending registry events to the
// network and (later on) process notifications of new documents on the network that might be of interest to the registyr.
type Ambassador interface {
	RegisterEventHandlers(fn events.EventRegistrar, eventType []events.EventType)
}

type ambassador struct {
	networkClient network.NetworkClient
	cryptoClient  crypto.Client
}

func NewAmbassador(networkClient network.NetworkClient, cryptoClient crypto.Client) Ambassador {
	return &ambassador{
		networkClient: networkClient,
		cryptoClient:  cryptoClient,
	}
}

func (n *ambassador) RegisterEventHandlers(fn events.EventRegistrar, eventType []events.EventType) {
	for _, eventType := range eventType {
		fn(eventType, func(event events.Event, lookup events.EventLookup) error {
			if event.Type() == domain.RegisterVendor {
				if err := n.registerVendorCertificateInTrustStore(event); err != nil {
					logrus.Errorf("Error while registering vendor certificates in the truststore (event=%s): %v", event.Ref(), err)
				}
			}
			go n.sendEventToNetwork(event)
			return nil
		})
	}
}

func (n *ambassador) registerVendorCertificateInTrustStore(event events.Event) error {
	payload := domain.RegisterVendorEvent{}
	if err := event.Unmarshal(&payload); err != nil {
		return err
	}
	for _, trustedCert := range cert.GetActiveCertificates(payload.Keys, time.Now()) {
		if err := n.cryptoClient.TrustStore().AddCertificate(trustedCert); err != nil {
			return errors2.Wrapf(err, "can't add vendor certificate to truststore (subject=%s,serial=%d)", trustedCert.Subject.String(), trustedCert.SerialNumber)
		}
	}
	return nil
}

func (n *ambassador) sendEventToNetwork(event events.Event) {
	// TODO: Should we be able to put other vendor's events on the network?
	eventData := event.Marshal()
	hash := model.CalculateDocumentHash(documentType, event.IssuedAt(), eventData)
	existingDocument, err := n.networkClient.GetDocument(hash)
	if err != nil {
		log.Errorf("Error while checking whether document exists on the network (event=%s,hash=%s): %v", event.IssuedAt(), hash, err)
		return
	} else if existingDocument != nil && existingDocument.HasContents {
		log.Debugf("Document already exists on the network (event=%s,hash=%s): %v", event.IssuedAt(), hash, err)
		return
	}
	document, err := n.networkClient.AddDocumentWithContents(event.IssuedAt(), documentType, eventData)
	if err != nil {
		log.Errorf("Error registering event on the network (event=%s,hash=%s): %v", event.IssuedAt(), hash, err)
		return
	}
	logrus.Infof("Event registered on network (event=%s,hash=%s)", event.IssuedAt(), document.Hash)
}
