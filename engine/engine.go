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

package engine

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nuts-foundation/nuts-crypto/pkg/cert"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/api"
	"github.com/nuts-foundation/nuts-registry/client"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/nuts-foundation/nuts-registry/pkg/db"
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"github.com/nuts-foundation/nuts-registry/pkg/events/domain"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"time"
)

// registryClientCreator is a variable to aid testability
var registryClientCreator = client.NewRegistryClient

// NewRegistryEngine returns the core definition for the registry
func NewRegistryEngine() *core.Engine {
	r := pkg.RegistryInstance()

	return &core.Engine{
		Cmd:       cmd(),
		Configure: r.Configure,
		Config:    &r.Config,
		ConfigKey: "registry",
		FlagSet:   flagSet(),
		Name:      pkg.ModuleName,
		Routes: func(router core.EchoRouter) {
			api.RegisterHandlers(router, &api.ApiWrapper{R: r})
		},
		Start:    r.Start,
		Shutdown: r.Shutdown,
	}
}

func flagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("registry", pflag.ContinueOnError)

	defs := pkg.DefaultRegistryConfig()
	flagSet.String(pkg.ConfDataDir, defs.Datadir, fmt.Sprintf("Location of data files, default: %s", defs.Datadir))
	flagSet.String(pkg.ConfMode, defs.Mode, fmt.Sprintf("server or client, when client it uses the HttpClient, default: %s", defs.Mode))
	flagSet.String(pkg.ConfAddress, defs.Address, fmt.Sprintf("Interface and port for http server to bind to, default: %s", defs.Address))
	flagSet.String(pkg.ConfSyncMode, defs.SyncMode, fmt.Sprintf("The method for updating the data, 'fs' for a filesystem watch or 'github' for a periodic download, default: %s", defs.SyncMode))
	flagSet.String(pkg.ConfSyncAddress, defs.SyncAddress, fmt.Sprintf("The remote url to download the latest registry data from, default: %s", defs.SyncAddress))
	flagSet.Int(pkg.ConfSyncInterval, defs.SyncInterval, fmt.Sprintf("The interval in minutes between looking for updated registry files on github, default: %d", defs.SyncInterval))
	flagSet.Int(pkg.ConfVendorCACertificateValidity, defs.VendorCACertificateValidity, fmt.Sprintf("Number of days vendor CA certificates are valid, default: %d", defs.VendorCACertificateValidity))
	flagSet.Int(pkg.ConfOrganisationCertificateValidity, defs.OrganisationCertificateValidity, fmt.Sprintf("Number of days organisation certificates are valid, default: %d", defs.OrganisationCertificateValidity))
	flagSet.Int(pkg.ConfClientTimeout, defs.ClientTimeout, fmt.Sprintf("Time-out for the client in seconds (e.g. when using the CLI), default: %d", defs.ClientTimeout))

	return flagSet
}

func cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "registry commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of the Nuts registry",
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Errorf("version 0.0.0")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "search [organization]",
		Short: "Find organizations within the registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cl := registryClientCreator()
			os, _ := cl.SearchOrganizations(args[0])

			logrus.Errorf("Found %d organizations\n", len(os))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "Run standalone api server",
		Run: func(cmd *cobra.Command, args []string) {
			i := pkg.RegistryInstance()

			echo := echo.New()
			echo.HideBanner = true
			echo.Use(middleware.Logger())
			api.RegisterHandlers(echo, &api.ApiWrapper{R: i})

			// todo move to nuts-go-core
			sigc := make(chan os.Signal, 1)
			signal.Notify(sigc, os.Interrupt, os.Kill)

			recoverFromEcho := func() {
				defer func() {
					recover()
				}()
				echo.Start(i.Config.Address)
			}

			go recoverFromEcho()
			<-sigc
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "register-vendor [name] [(optional, default=healthcare) domain]",
		Short: "Registers a vendor",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := registryClientCreator()
			var vendorDomain = domain.FallbackDomain
			if len(args) == 2 {
				vendorDomain = args[1]
			}
			event, err := cl.RegisterVendor(args[0], vendorDomain)
			if err != nil {
				logrus.Errorf("Unable to register vendor: %v", err)
				return err
			}
			logrus.Info("Vendor registered.")
			logEventToConsole(event)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "refresh-vendor-cert",
		Short: "Issues a new vendor certificate using existing keys (or newly generated if none exist)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := registryClientCreator()
			event, err := cl.RefreshVendorCertificate()
			if err != nil {
				logrus.Errorf("Unable to refresh vendor certificate: %v", err)
				return err
			}
			payload := domain.RegisterVendorEvent{}
			err = event.Unmarshal(&payload)
			if err != nil {
				logrus.Error("Unable to parse event payload.", err)
			}
			certificates := cert.GetActiveCertificates(payload.Keys, time.Now())
			if len(certificates) == 0 {
				logrus.Error("Certificate refresh succeeded, but couldn't find any activate certificate.")
			} else {
				// GetActiveCertificates returns the certificate with the longest validity first.
				logrus.Infof("Vendor certificate refreshed, new certificate is valid until %v", certificates[0].NotAfter)
			}
			logEventToConsole(event)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "refresh-organization-cert [orgID]",
		Short: "Issues a new organization certificate using existing keys (or newly generated if none exist)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := registryClientCreator()
			event, err := cl.RefreshOrganizationCertificate(args[0])
			if err != nil {
				logrus.Errorf("Unable to refresh organization certificate: %v", err)
				return err
			}
			payload := domain.VendorClaimEvent{}
			err = event.Unmarshal(&payload)
			if err != nil {
				logrus.Error("Unable to parse event payload.", err)
			}
			certificates := cert.GetActiveCertificates(payload.OrgKeys, time.Now())
			if len(certificates) == 0 {
				logrus.Error("Certificate refresh succeeded, but couldn't find any activate certificate.")
			} else {
				// GetActiveCertificates returns the certificate with the longest validity first.
				logrus.Infof("Organization certificate refreshed, new certificate is valid until %v", certificates[0].NotAfter)
			}
			logEventToConsole(event)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "vendor-claim [org-identifier] [org-name]",
		Short: "Registers a vendor claim.",
		Long:  "Registers a vendor claiming a care organization as its client.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := registryClientCreator()
			event, err := cl.VendorClaim(args[0], args[1], nil)
			if err != nil {
				logrus.Errorf("Unable to register vendor organisation claim: %v", err)
				return err
			}
			logrus.Info("Vendor organisation claim registered.")
			logEventToConsole(event)
			return nil
		},
	})

	{
		var fix *bool
		command := &cobra.Command{
			Use:   "verify",
			Short: "Verifies data in the registry.",
			Long:  "Verifies the vendor's own data in the registry, use --fix or -f to fix/upgrade data.",
			RunE: func(cmd *cobra.Command, args []string) error {
				cl := registryClientCreator()
				logrus.Info("Verifying...")
				resultingEvents, needsFixing, err := cl.Verify(*fix)
				if err != nil {
					logrus.Errorf("Verification error: %v", err)
					return err
				}
				if needsFixing {
					logrus.Warn("Verification complete, data must be fixed. Please rerun command with --fix or -f")
				} else {
					logrus.Info("Verification complete")
				}
				if len(resultingEvents) > 0 {
					logrus.Infof("Data was fixed and %d events were emitted:", len(resultingEvents))
					for _, event := range resultingEvents {
						logEventToConsole(event)
					}
				}
				return nil
			},
		}
		flagSet := pflag.NewFlagSet("verify", pflag.ContinueOnError)
		fix = flagSet.BoolP("fix", "f", false, "fix/upgrade data")
		command.Flags().AddFlagSet(flagSet)
		cmd.AddCommand(command)
	}

	{
		var properties *[]string
		var id *string
		command := &cobra.Command{
			Use:   "register-endpoint [org-identifier] [type] [url]",
			Short: "Registers an endpoint",
			Long:  "Registers an endpoint for an organization.",
			Args:  cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				cl := registryClientCreator()
				event, err := cl.RegisterEndpoint(args[0], *id, args[2], args[1], db.StatusActive, parseCLIProperties(*properties))
				if err != nil {
					logrus.Errorf("Unable to register endpoint: %v", err)
					return err
				}
				logrus.Info("Endpoint registered.")
				logEventToConsole(event)
				return nil
			},
		}
		flagSet := pflag.NewFlagSet("register-endpoint", pflag.ContinueOnError)
		properties = flagSet.StringArrayP("property", "p", nil, "extra properties for the endpoint, in the format: key=value")
		id = flagSet.StringP("id", "i", "", "endpoint identifier, defaults to a random GUID when not set")
		command.Flags().AddFlagSet(flagSet)
		cmd.AddCommand(command)
	}

	return cmd
}

// parseCLIProperties parses a slice of key-value entries (key=value) to a map.
func parseCLIProperties(keysAndValues []string) map[string]string {
	result := make(map[string]string, 0)
	for _, keyAndValue := range keysAndValues {
		parts := strings.SplitN(keyAndValue, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func logEventToConsole(event events.Event) {
	println("Event:", events.SuggestEventFileName(event))
	println(string(event.Marshal()))
}
