/*
Copyright © 2025 Timothy Tavarez <timothy.tavarez@decombine.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/lipgloss"
	"github.com/decombine/slc"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

//go:embed locales/*
var embeds embed.FS
var langs = []string{"en", "es"}
var cfgFile, Lang string
var localizer *i18n.Localizer

var (
	NetworkKey = "network"
	ConfigPath = "/.config/contract/"
)

var rootLogoStyle = lipgloss.NewStyle().
	Bold(false).
	Foreground(lipgloss.Color("#227BF0")).
	PaddingTop(1).
	PaddingLeft(4).
	Width(100)

var contractBlue = lipgloss.AdaptiveColor{Light: "#227BF0", Dark: "#227BF0"}

var style = lipgloss.NewStyle().Foreground(contractBlue)

var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

var DecombineNetwork = slc.Network{
	Name:              "decombine",
	API:               "https://api.decombine.com",
	URL:               "https://decombine.com",
	ClientID:          "314914854450233349",
	Issuer:            "https://auth.decombine.com",
	DiscoveryEndpoint: "https://auth.decombine.com/.well-known/openid-configuration",
}

type NetworkClient struct {
	Address string
	Timeout time.Duration
}

var DecombineClient = NetworkClient{
	Address: "http://api.decombine.local:8025",
	Timeout: 10 * time.Second,
}

var httpClient = &http.Client{
	Timeout: DecombineClient.Timeout,
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "contract",
	Short: local("RootDescriptionShort"),
	Long:  rootLogoStyle.Render(local("RootDescriptionLong")),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/contract/contract.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home/.config/.contract with name "contract" (without extension).
		viper.AddConfigPath(home + ConfigPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("contract")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	//var networks []slc.Network
	//networks = append(networks, DecombineNetwork)
	//viper.SetDefault("networks", networks)
}

func initLang(lang string) *i18n.Localizer {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, l := range langs {
		f, err := embeds.ReadFile(fmt.Sprintf("locales/locale.%s.toml", l))
		if err != nil {
			fmt.Println(err)
		}
		bundle.MustParseMessageFileBytes(f, fmt.Sprintf("locales/locale.%s.toml", l))
	}
	return i18n.NewLocalizer(bundle, lang)
}

func local(id string) string {
	l := initLang(Lang)
	return l.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: id,
		}})
}

type ContractCLIConfig struct {
	Language                string        `yaml:"language" toml:"language" json:"language"`
	DefaultContractFileType string        `yaml:"defaultContractFileType" toml:"defaultContractFileType" json:"defaultContractFileType"`
	DefaultNetwork          string        `yaml:"defaultNetwork" toml:"defaultNetwork" json:"defaultNetwork"`
	Networks                []slc.Network `yaml:"networks" toml:"networks" json:"networks"`
}
