// +build !race

/*
 * Nuts registry
 * Copyright (C) 2019. Nuts community
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

package pkg

import (
	"github.com/google/uuid"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-go-test/io"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/nuts-foundation/nuts-registry/pkg/types"
	"github.com/nuts-foundation/nuts-registry/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)


// Test_Functional_OutOfOrder_Registrations tests that events that are created in chronological order but processed
// out of order (which can be the case when received through Nuts Network) are still processed later on.
func Test_Functional_OutOfOrder_Registrations(t *testing.T) {
	os.Setenv("NUTS_IDENTITY", test.VendorID("4").String())
	core.NutsConfig().Load(&cobra.Command{})
	configureIdleTimeout()
	registry := NewTestRegistryInstance(io.TestDirectory(t))
	vendorID := test.VendorID("1234")
	organizationID := test.OrganizationID("5678")
	registerVendorEvent := events.CreateEvent(domain.RegisterVendor, domain.RegisterVendorEvent{
		Identifier: vendorID,
		Name:       "Awesome Vendor",
		Domain:     types.HealthcareDomain,
	}, nil)
	time.Sleep(1 * time.Millisecond) // Make sure events have incrementing timestamps
	vendorClaimEvent := events.CreateEvent(domain.VendorClaim, domain.VendorClaimEvent{
		OrganizationID: organizationID,
		VendorID:       vendorID,
		OrgName:        "Awesome Organization",
		Start:          time.Now(),
	}, nil)
	time.Sleep(1 * time.Millisecond) // Make sure events have incrementing timestamps
	registerEndpointEvent := events.CreateEvent(domain.RegisterEndpoint, domain.RegisterEndpointEvent{
		Organization: organizationID,
		URL:          "http://foobar",
		EndpointType: "test",
		Identifier:   types.EndpointID(uuid.New().String()),
		Status:       "active",
	}, nil)

	err := registry.EventSystem.ProcessEvent(registerEndpointEvent)
	assert.Error(t, err)
	err = registry.EventSystem.ProcessEvent(vendorClaimEvent)
	assert.Error(t, err)
	err = registry.EventSystem.ProcessEvent(registerVendorEvent)
	assert.NoError(t, err)
	endpoints, _ := registry.Db.FindEndpointsByOrganizationAndType(organizationID, nil)
	assert.Len(t, endpoints, 1)
}