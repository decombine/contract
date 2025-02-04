package cmd

import (
	"fmt"
	"github.com/decombine/slc"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	rootCmd.AddCommand(networkCmd)
	removeNetworkCmd.Flags().StringP("network", "n", "", "The name of the Network")
	networkCmd.AddCommand(addNetworkCmd)
	networkCmd.AddCommand(removeNetworkCmd)
	networkCmd.AddCommand(setNetworkCmd)
	networkCmd.AddCommand(networkListCmd)
	setNetworkCmd.Flags().StringP("network", "n", "", "The name of the Network")
	addNetworkCmd.Flags().StringP("network", "n", "", "The name of the Network")
	addNetworkCmd.Flags().StringP("api", "p", "", "The API for the Network")
	addNetworkCmd.Flags().StringP("url", "u", "", "The URL of the Network")
	addNetworkCmd.Flags().StringP("client-id", "c", "", "The Client ID for the Network")
	addNetworkCmd.Flags().StringP("domain", "d", "", "The Domain for the Network")
	addNetworkCmd.Flags().StringP("oidc", "a", "", "The Discovery Endpoint for the Network")
	addNetworkCmd.Flags().StringP("output", "o", "json", "Define the output format for the command (json, yaml, toml)")
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Configure a Smart Legal Contract Network",
	Long: `Add, remove, and configure Smart Legal Contract networks. Networks provide the ability to deploy
Smart Legal Contacts with remote state, workloads, and policies.`,
}

var addNetworkCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a Smart Legal Contract Network",
	Long:  `Add a Smart Legal Contract Network to the available options in the Contract configuration.`,
	RunE:  addNetworkExecute,
}

func addNetworkExecute(cmd *cobra.Command, args []string) error {
	name := cmd.Flags().Lookup("network").Value.String()
	api := cmd.Flags().Lookup("api").Value.String()
	url := cmd.Flags().Lookup("url").Value.String()
	clientID := cmd.Flags().Lookup("client-id").Value.String()
	domain := cmd.Flags().Lookup("domain").Value.String()
	oidc := cmd.Flags().Lookup("oidc").Value.String()
	err := addNetwork(name, api, url, clientID, domain, oidc)
	if err != nil {
		return err
	}
	return nil
}

var removeNetworkCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Smart Legal Contract Network",
	Long:  `Remove a Smart Legal Contract Network from the available options in the Contract configuration.`,
	RunE:  removeNetworkExecute,
}

func removeNetworkExecute(cmd *cobra.Command, args []string) error {
	name := cmd.Flags().Lookup("network").Value.String()

	if name == "" {
		return fmt.Errorf("network name is required")
	}

	networks := Networks()

	for i, n := range networks {

		if name == "decombine" {
			return fmt.Errorf("cannot remove default network")
		}

		if n.Name == name {
			fmt.Printf("Removing network %s\n", name)
			networks = append(networks[:i], networks[i+1:]...)
			break
		}
	}

	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		return fmt.Errorf("no config file found")
	}
	file, err := os.ReadFile(cfg)
	if err != nil {
		return err
	}
	var config ContractCLIConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}
	config.Networks = networks
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(cfg, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Smart Legal Contract Networks",
	Long:  `List the available Smart Legal Contract Networks in the Contract configuration.`,
	RunE:  networkListExecute,
}

func networkListExecute(cmd *cobra.Command, args []string) error {

	networks := Networks()
	if len(networks) == 0 {
		fmt.Println("No networks configured.")
		return nil
	} else {
		for _, n := range networks {
			printNetwork(n)
		}
	}

	return nil
}

func addDefaultNetwork() {
	viper.SetDefault("defaultNetwork", DecombineNetwork)
	viper.SetDefault("network", DecombineNetwork.Name)
}

var setNetworkCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the default Smart Legal Contract Network",
	Long:  `Set the default Smart Legal Contract Network for the Contract configuration.`,
	RunE:  setNetworkExecute,
}

func setNetworkExecute(cmd *cobra.Command, args []string) error {
	name := cmd.Flags().Lookup("network").Value.String()
	if name == "" {
		return fmt.Errorf("network name is required")
	}
	networks := Networks()
	for _, n := range networks {
		if n.Name == name {
			viper.Set("defaultNetwork", n)
			viper.Set("network", n.Name)
			cfg, err := Config()
			if err != nil {
				return err
			}
			cfg.DefaultNetwork = n.Name
			err = UpdateConfig(viper.ConfigFileUsed(), &cfg)
			if err != nil {
				return err
			}
			fmt.Printf("Default network set to %s\n", name)
			return nil
		}
	}
	return fmt.Errorf("network %s not found", name)
}

func printNetwork(n slc.Network) {
	fmt.Printf(style.Render("Network: ")+"%s\n", n.Name)
	fmt.Printf("API: %s\n", n.API)
	fmt.Printf("URL: %s\n", n.URL)
	fmt.Printf("Client ID: %s\n", n.ClientID)
	fmt.Printf("Domain: %s\n", n.Domain)
	fmt.Printf("Discovery Endpoint: %s\n", n.DiscoveryEndpoint)
}

func UpdateConfig(path string, config *ContractCLIConfig) error {
	_, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading config file")
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Networks() []slc.Network {
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		return []slc.Network{DecombineNetwork}
	}

	file, err := os.ReadFile(cfg)

	if err != nil {
		return []slc.Network{DecombineNetwork}
	}

	var config ContractCLIConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return []slc.Network{DecombineNetwork}
	}
	return config.Networks
}

func addNetwork(name, api, url, clientID, domain, oidc string) error {
	err := checkNetworkValues(name, api, url, clientID, domain, oidc)
	if err != nil {
		return err
	}
	networks := Networks()
	networks = append(networks, slc.Network{
		Name:              name,
		API:               api,
		URL:               url,
		ClientID:          clientID,
		Domain:            domain,
		DiscoveryEndpoint: oidc,
	})
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		return fmt.Errorf("no config file found")
	}
	file, err := os.ReadFile(cfg)
	if err != nil {
		return err
	}
	var config ContractCLIConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}
	config.Networks = networks
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(cfg, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func checkNetworkValues(name, api, url, clientID, domain, oidc string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if api == "" {
		return fmt.Errorf("api is required")
	}
	if url == "" {
		return fmt.Errorf("url is required")
	}
	if clientID == "" {
		return fmt.Errorf("client id is required")
	}
	if domain == "" {
		return fmt.Errorf("domain is required")
	}
	if oidc == "" {
		return fmt.Errorf("discovery endpoint is required")
	}
	return nil
}
