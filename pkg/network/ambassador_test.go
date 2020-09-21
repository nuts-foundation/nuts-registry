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
	"github.com/nuts-foundation/nuts-crypto/pkg"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-go-test/io"
	pkg2 "github.com/nuts-foundation/nuts-network/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
	"time"
)

const eventType = events.EventType("testEvent")

func Test_ambassador_Send(t *testing.T) {
	eventSystem, networkInstance := createInstance(t)
	// Test that we can send a registry event through the network
	err := eventSystem.ProcessEvent(events.CreateEvent(eventType, "Hello, rest of Network!", nil))
	if !assert.NoError(t, err) {
		return
	}
	time.Sleep(500 * time.Millisecond) // Async process, wait for event to be processed
	documents, err := networkInstance.ListDocuments()
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, documents, 1)
}

func Test_ambassador_Receive(t *testing.T) {
	eventSystem, networkInstance := createInstance(t)
	var eventsHandled sync.WaitGroup
	eventsHandled.Add(1)
	eventSystem.RegisterEventHandler(eventType, func(event events.Event, lookup events.EventLookup) error {
		eventsHandled.Done()
		return nil
	})
	event := events.CreateEvent(eventType, "Hello from Network!", nil)
	_, err := networkInstance.AddDocumentWithContents(time.Now(), documentType, event.Marshal())
	if !assert.NoError(t, err) {
		return
	}
	eventsHandled.Wait()
}

func createInstance(t *testing.T) (events.EventSystem, *pkg2.Network) {
	os.Setenv("NUTS_IDENTITY", test.VendorID("4").String())
	core.NutsConfig().Load(&cobra.Command{})
	testDirectory := io.TestDirectory(t)
	eventSystem := events.NewEventSystem(eventType)
	eventSystem.Configure(testDirectory)
	networkInstance := pkg2.NewTestNetworkInstance(testDirectory)
	ambassador := NewAmbassador(networkInstance, pkg.NewTestCryptoInstance(testDirectory), eventSystem)
	ambassador.RegisterEventHandlers(eventSystem.RegisterEventHandler, []events.EventType{eventType})
	ambassador.Start()
	return eventSystem, networkInstance
}
