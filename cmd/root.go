/*
 * Nuts registry
 * Copyright (C) 2019 Nuts community
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
 */

package cmd

import (
	goflag "flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-registry/api"
	"github.com/nuts-foundation/nuts-registry/db"
	"github.com/nuts-foundation/nuts-registry/generated"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "nuts-registry",
	Short: "The Nuts registry",
	Long:  `The Nuts registry, containing Organisation to endpoint mappings`,
	Run: func(cmd *cobra.Command, args []string) {

		// load static db
		memoryDb := db.New()
		err := memoryDb.Load(viper.GetString(db.ConfDataDir))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// start webserver
		e := echo.New()
		generated.RegisterHandlers(e, api.ApiResource{Db: memoryDb})
		e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", viper.GetString("interface"), viper.GetInt("port"))))
	},
}

func init() {
	//commands
	rootCmd.Flags().StringP(confPort, "p", "1323", "Server listen port")
	rootCmd.Flags().String(confInterface, "localhost", "Server interface binding")
	rootCmd.Flags().String(db.ConfDataDir, "./data", "Location of data files")

	viper.BindPFlag(confPort, rootCmd.Flags().Lookup(confPort))
	viper.BindPFlag(confInterface, rootCmd.Flags().Lookup(confInterface))
	viper.BindPFlag(db.ConfDataDir, rootCmd.Flags().Lookup(db.ConfDataDir))

	viper.SetEnvPrefix("NUTS_REGISTRY")
	viper.BindEnv(confPort)
	viper.BindEnv(confInterface)
	viper.BindEnv(db.ConfDataDir)

	rootCmd.AddCommand(NewVersionCmd())

	// initialize logging
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
}

func Execute() {
	rootCmd.Execute()
}
