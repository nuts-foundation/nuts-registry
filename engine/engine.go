package engine

import (
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
	engine "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-registry/api"
	"github.com/nuts-foundation/nuts-registry/client"
	"github.com/nuts-foundation/nuts-registry/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// NewRegistryEngine returns the engine definition for the registry
func NewRegistryEngine() *engine.Engine {
	r := pkg.RegistryInstance()

	return &engine.Engine{
		Cmd: cmd(),
		Configure: r.Configure,
		Config:    &r.Config,
		ConfigKey: "registry",
		FlagSet:   flagSet(),
		Name:      "Registry",
		Routes: func(router runtime.EchoRouter) {
			api.RegisterHandlers(router, &api.ApiWrapper{R: r})
		},
	}
}

func flagSet() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("registry", pflag.ContinueOnError)

	flagSet.String(pkg.ConfDataDir, "./data", "Location of data files")
	flagSet.String(pkg.ConfMode, "server", "server or client, when client it uses the HttpClient")
	flagSet.String(pkg.ConfAddress, "localhost:1323", "Interface and port for http server to bind to")

	return flagSet
}

func cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "registry",
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
		Args: cobra.ExactArgs(1),
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

			logrus.Fatal(echo.Start(i.Config.Address))
		},
	})

	return cmd
}