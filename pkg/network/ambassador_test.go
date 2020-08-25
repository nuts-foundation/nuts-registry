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

func Test_ambassador_SendAndReceive(t *testing.T) {
	os.Setenv("NUTS_IDENTITY", test.VendorID("4").String())
	core.NutsConfig().Load(&cobra.Command{})
	testDirectory := io.TestDirectory(t)
	const eventType = events.EventType("testEvent")
	var eventsHandled sync.WaitGroup
	eventsHandled.Add(2)
	eventSystem := events.NewEventSystem(eventType)
	eventSystem.RegisterEventHandler(eventType, func(event events.Event, lookup events.EventLookup) error {
		eventsHandled.Done()
		return nil
	})
	eventSystem.Configure(testDirectory)
	networkInstance := pkg2.NewTestNetworkInstance(testDirectory)
	ambassador := NewAmbassador(networkInstance, pkg.NewTestCryptoInstance(testDirectory), eventSystem)
	ambassador.RegisterEventHandlers(eventSystem.RegisterEventHandler, []events.EventType{eventType})
	ambassador.Start()
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
	// Test that we can receive a registry event through the network
	event := events.CreateEvent(eventType, "Hello from Network!", nil)
	_, err = networkInstance.AddDocumentWithContents(time.Now(), documentType, event.Marshal())
	if !assert.NoError(t, err) {
		return
	}
	eventsHandled.Wait()
}