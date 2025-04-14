package cmd

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the Contract CLI",
	Long:  `Configure the Contract CLI with the necessary settings to interact with the Contract network.`,
	Run:   configExecute,
}

func configExecute(cmd *cobra.Command, args []string) {
	config, err := Config()
	if err != nil {
		fmt.Println("error reading config file:", err)
		return
	}
	echoConfig(config)
}

func Config() (ContractCLIConfig, error) {
	cfg := viper.ConfigFileUsed()
	if cfg == "" {
		return ContractCLIConfig{}, fmt.Errorf("no config file found")
	}
	file, err := os.ReadFile(cfg)
	if err != nil {
		return ContractCLIConfig{}, fmt.Errorf("error reading config file: %w", err)
	}
	var config ContractCLIConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return ContractCLIConfig{}, fmt.Errorf("error unmarshalling config file: %w", err)
	}
	return config, nil
}

func echoConfig(config ContractCLIConfig) {
	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println("error marshalling config file:", err)
		return
	}
	fmt.Println(string(data))
}
