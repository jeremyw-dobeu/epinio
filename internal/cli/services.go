package cli

import (
	"fmt"

	"github.com/epinio/epinio/internal/cli/usercmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CmdServices implements the command: epinio services
var CmdServices = &cobra.Command{
	Use:           "service",
	Aliases:       []string{"services"},
	Short:         "Epinio service management",
	Long:          `Manage the epinio services`,
	SilenceErrors: false,
	Args:          cobra.ExactArgs(1),
}

func init() {
	CmdServiceDelete.Flags().Bool("unbind", false, "Unbind from applications before deleting")
	CmdServices.AddCommand(CmdServiceCatalog)
	CmdServices.AddCommand(CmdServiceCreate)
	CmdServices.AddCommand(CmdServiceBind)
	CmdServices.AddCommand(CmdServiceUnbind)
	CmdServices.AddCommand(CmdServiceShow)
	CmdServices.AddCommand(CmdServiceDelete)
	CmdServices.AddCommand(CmdServiceList)

	CmdServiceList.Flags().Bool("all", false, "list all services")
}

var CmdServiceCatalog = &cobra.Command{
	Use:               "catalog [NAME]",
	Short:             "Lists all available Epinio catalog services, or show the details of the specified one",
	ValidArgsFunction: matchingCatalogFinder,
	Args:              cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New()
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		if len(args) == 0 {
			err = client.ServiceCatalog()
			return errors.Wrap(err, "error listing Epinio catalog services")
		}

		if len(args) == 1 {
			serviceName := args[0]
			err = client.ServiceCatalogShow(serviceName)
			return errors.Wrap(err, fmt.Sprintf("error showing %s Epinio catalog service", serviceName))
		}

		return nil
	},
}

var CmdServiceCreate = &cobra.Command{
	Use:               "create CATALOGSERVICENAME SERVICENAME",
	Short:             "Create a service SERVICENAME of an Epinio catalog service CATALOGSERVICENAME",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: matchingCatalogFinder,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New()
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		catalogServiceName := args[0]
		serviceName := args[1]

		err = client.ServiceCreate(catalogServiceName, serviceName)
		return errors.Wrap(err, "error creating service")
	},
}

var CmdServiceShow = &cobra.Command{
	Use:               "show SERVICENAME",
	Short:             "Show details of a service SERVICENAME",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: matchingServiceFinder,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New()
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		serviceName := args[0]

		err = client.ServiceShow(serviceName)
		return errors.Wrap(err, "error showing service")
	},
}

var CmdServiceDelete = &cobra.Command{
	Use:               "delete SERVICENAME",
	Short:             "Delete service SERVICENAME",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: matchingServiceFinder,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		unbind, err := cmd.Flags().GetBool("unbind")
		if err != nil {
			return errors.Wrap(err, "error reading option --unbind")
		}

		client, err := usercmd.New()
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		serviceName := args[0]

		err = client.ServiceDelete(serviceName, unbind)
		return errors.Wrap(err, "error deleting service")
	},
}
var CmdServiceBind = &cobra.Command{
	Use:               "bind SERVICENAME APPNAME",
	Short:             "Bind a service SERVICENAME to an Epinio app APPNAME",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: findServiceApp,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New()
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		serviceName := args[0]
		appName := args[1]

		err = client.ServiceBind(serviceName, appName)
		return errors.Wrap(err, "error binding service")
	},
}

var CmdServiceUnbind = &cobra.Command{
	Use:               "unbind SERVICENAME APPNAME",
	Short:             "Unbinds a service SERVICENAME from an Epinio app APPNAME",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: findServiceApp,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New()
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		serviceName := args[0]
		appName := args[1]

		err = client.ServiceUnbind(serviceName, appName)
		return errors.Wrap(err, "error unbinding service")
	},
}

var CmdServiceList = &cobra.Command{
	Use:   "list",
	Short: "List all the services in the targeted namespace",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New()
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			return errors.Wrap(err, "error reading option --all")
		}

		if all {
			err = client.ServiceListAll()
		} else {
			err = client.ServiceList()
		}

		return errors.Wrap(err, "error listing services")
	},
}

func findServiceApp(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	app, err := usercmd.New()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(args) == 1 {
		// #args == 1: app name.
		matches := app.AppsMatching(toComplete)
		return matches, cobra.ShellCompDirectiveNoFileComp
	}

	// #args == 0: configuration name.

	matches := app.ServiceMatching(toComplete)
	return matches, cobra.ShellCompDirectiveNoFileComp
}
