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
package network

import (
	"bytes"
	log "github.com/nuts-foundation/nuts-crypto/log"
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
	// Start instructs the ambassador to start receiving events from the network.
	Start()
}

type ambassador struct {
	networkClient network.NetworkClient
	cryptoClient  crypto.Client
	eventSystem   events.EventSystem
}

// NewAmbassador creates a new Ambassador. Don't forget to call RegisterEventHandlers afterwards.
func NewAmbassador(networkClient network.NetworkClient, cryptoClient crypto.Client, eventSystem events.EventSystem) Ambassador {
	instance := &ambassador{
		networkClient: networkClient,
		cryptoClient:  cryptoClient,
		eventSystem:   eventSystem,
	}
	return instance
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

// Start instructs the ambassador to start receiving events from the network.
func (n *ambassador) Start() {
	queue := n.networkClient.Subscribe(documentType)
	go func() {
		for {
			document := queue.Get()
			if document == nil {
				return
			}
			n.processDocument(document)
		}
	}()
}

func (n *ambassador) sendEventToNetwork(event events.Event) {
	// For now we just send every event to the network, event other node's events. They're signed so they can't be
	// edited anyways and it assures the registry shadow copy on the network is populated ASAP.
	eventData := event.Marshal()
	document, err := n.networkClient.AddDocumentWithContents(event.IssuedAt(), documentType, eventData)
	if err != nil {
		log.Logger().Errorf("Error registering event on the network (event=%s): %v", event.IssuedAt(), err)
		return
	}
	logrus.Infof("Event registered on network (event=%s,hash=%s)", event.IssuedAt(), document.Hash)
}

func (n *ambassador) processDocument(document *model.Document) {
	log.Logger().Infof("Received event through Nuts Network: %s", document.Hash)
	reader, err := n.networkClient.GetDocumentContents(document.Hash)
	if err != nil {
		log.Logger().Errorf("Unable to retrieve document from Nuts Network (hash=%s): %v", document.Hash, err)
		return
	}
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(reader); err != nil {
		log.Logger().Errorf("Unable read document data from Nuts Network (hash=%s): %v", document.Hash, err)
		return
	}
	if event, err := events.EventFromJSON(buf.Bytes()); err != nil {
		log.Logger().Errorf("Unable parse event from Nuts Network (hash=%s): %v", document.Hash, err)
	} else {
		_ = n.eventSystem.ProcessEvent(event)
	}
}
