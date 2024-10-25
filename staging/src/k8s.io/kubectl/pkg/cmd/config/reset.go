package config

import (
	"fmt"
	"io"
	"reflect"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	currentContextPropertyName        = `current-context`
	preferencesColorsPropertyName     = `preferences.colors`
	preferencesExtensionsPropertyName = `preferences.extensions`
)

var (
	resetLong = templates.LongDesc(i18n.T(`Reset all the configs in ths specified kubeconfig file.`))

	resetExample = templates.Examples(`
		# Reset all configs in kubeconfig
		kubectl config reset`)
)

// NewCmdConfigReset returns a Command instance for 'config reset' sub command
func NewCmdConfigReset(streams genericiooptions.IOStreams, ConfigAccess clientcmd.ConfigAccess) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reset",
		Short:   i18n.T("Reset all the configs in ths specified kubeconfig file."),
		Long:    resetLong,
		Example: resetExample,
		Run: func(cmd *cobra.Command, args []string) {
			configFile := ConfigAccess.GetDefaultFilename()
			if ConfigAccess.IsExplicitFile() {
				configFile = ConfigAccess.GetExplicitFile()
			}

			config, err := ConfigAccess.GetStartingConfig()
			cmdutil.CheckErr(err)

			cmdutil.CheckErr(unsetCurrentContext(config, streams.Out, configFile))
			cmdutil.CheckErr(unsetPreferences(config, streams.Out, configFile))
			cmdutil.CheckErr(deletePrimaryConfigs(ConfigAccess, *config, streams.Out, configFile))

		},
	}

	flags := genericclioptions.NewPrintFlags("").WithTypeSetter(scheme.Scheme).WithDefaultOutput("yaml")
	flags.AddFlags(cmd)

	return cmd
}

func deletePrimaryConfigs(
	configAccess clientcmd.ConfigAccess,
	config clientcmdapi.Config,
	out io.Writer,
	configFile string,
) error {
	var (
		err            error
		deletedConfigs = struct {
			clusters []string
			contexts []string
			users    []string
		}{}
	)

	for cluster := range config.Clusters {
		delete(config.Clusters, cluster)
		deletedConfigs.clusters = append(deletedConfigs.clusters, cluster)
	}

	for context := range config.Contexts {
		delete(config.Contexts, context)
		deletedConfigs.contexts = append(deletedConfigs.contexts, context)
	}

	for user := range config.AuthInfos {
		delete(config.AuthInfos, user)
		deletedConfigs.users = append(deletedConfigs.users, user)
	}

	err = clientcmd.ModifyConfig(configAccess, config, true)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "Deleted all cluster(s) %v from %q\n", deletedConfigs.clusters, configFile)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "Deleted all context(s) %v from %q\n", deletedConfigs.contexts, configFile)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "Deleted all user(s) %v from %q\n", deletedConfigs.users, configFile)
	if err != nil {
		return err
	}

	return nil
}

func unsetCurrentContext(
	config *clientcmdapi.Config,
	out io.Writer,
	configFile string,
) error {
	steps, err := newNavigationSteps(currentContextPropertyName)
	if err != nil {
		return err
	}

	err = modifyConfig(reflect.ValueOf(config), steps, "", true, true)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "Property %q unset from %q\n", currentContextPropertyName, configFile)
	if err != nil {
		return err
	}

	return nil
}

func unsetPreferences(
	config *clientcmdapi.Config,
	out io.Writer,
	configFile string,
) error {
	steps, err := newNavigationSteps(preferencesColorsPropertyName)
	if err != nil {
		return err
	}

	err = modifyConfig(reflect.ValueOf(config), steps, "", true, true)
	if err != nil {
		return err
	}

	steps, err = newNavigationSteps(preferencesExtensionsPropertyName)
	if err != nil {
		return err
	}

	err = modifyConfig(reflect.ValueOf(config), steps, "", true, true)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "All preferences are unset from %q\n", configFile)
	if err != nil {
		return err
	}

	return nil
}
