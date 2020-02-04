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
	"github.com/nuts-foundation/nuts-registry/pkg/events"
	"os"
	"os/signal"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/api"
	"github.com/nuts-foundation/nuts-registry/client"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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
		Routes: func(router runtime.EchoRouter) {
			api.RegisterHandlers(router, &api.ApiWrapper{R: r})
		},
		Start:    r.Start,
		Shutdown: r.Shutdown,
	}
}

func flagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("registry", pflag.ContinueOnError)

	flagSet.String(pkg.ConfDataDir, "./data", "Location of data files")
	flagSet.String(pkg.ConfMode, "server", "server or client, when client it uses the HttpClient")
	flagSet.String(pkg.ConfAddress, "localhost:1323", "Interface and port for http server to bind to")
	flagSet.String(pkg.ConfSyncMode, "fs", "The method for updating the data, 'fs' for a filesystem watch or 'github' for a periodic download from github")
	flagSet.String(pkg.ConfSyncAddress, "https://codeload.github.com/nuts-foundation/nuts-registry-development/tar.gz/master", "The remote url to download the latest registry data from github")
	flagSet.Int(pkg.ConfSyncInterval, 30, "The interval in minutes between looking for updated registry files on github")

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
			cl := client.NewRegistryClient()
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
		Use:   "register-vendor [name] [identifier]",
		Short: "Registers a vendor",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			identifier := events.Identifier(args[1])
			event, _ := events.CreateEvent(events.RegisterVendor, events.RegisterVendorEvent{Name: name, Identifier: identifier})
			logrus.Info(events.SuggestEventFileName(event))
			logrus.Info(string(event.Marshal()))
		},
	})

	return cmd
}
