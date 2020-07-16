package network

import (
	"github.com/labstack/gommon/log"
	crypto "github.com/nuts-foundation/nuts-crypto/pkg"
	network "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-network/pkg/model"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/sirupsen/logrus"
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

// NewAmbassador creates a new Ambassador. Don't forget to call RegisterEventHandlers afterwards.
func NewAmbassador(networkClient network.NetworkClient, cryptoClient crypto.Client) Ambassador {
	return &ambassador{
		networkClient: networkClient,
		cryptoClient:  cryptoClient,
	}
}

// RegisterEventHandlers this event handler which is required for it to actually work.
func (n *ambassador) RegisterEventHandlers(fn events.EventRegistrar, eventType []events.EventType) {
	for _, eventType := range eventType {
		fn(eventType, func(event events.Event, lookup events.EventLookup) error {
			go n.sendEventToNetwork(event)
			return nil
		})
	}
}

func (n *ambassador) sendEventToNetwork(event events.Event) {
	// For now we just send every event to the network, event other node's events. They're signed so they can't be
	// edited anyways and it assures the registry shadow copy on the network is populated ASAP.
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
