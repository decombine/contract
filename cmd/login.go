package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/decombine/slc"
	"github.com/spf13/cobra"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

var (
	scopes       = []string{"openid", "offline_access", "email", "profile"}
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC1A"))
	highlight    = lipgloss.NewStyle().Background(lipgloss.Color("#0268A8"))
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	loginStyle   = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A881FC"))
)

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("network", "n", "", "The name of the Network to login to")
	loginCmd.Flags().BoolP("device-flow", "d", true, "Use device flow for authentication")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to a Contract Network",
	Long:  `Login to a Contract Network to deploy Smart Legal Contracts.`,
	RunE:  loginExecute,
}

func loginExecute(cmd *cobra.Command, args []string) error {
	deviceFlow, _ := cmd.Flags().GetBool("device-flow")
	name := cmd.Flags().Lookup("network").Value.String()

	if name == "" {
		cfg, err := Config()
		if err != nil {
			fmt.Printf(helpStyle("No configuration file found. Consider running contract init config."))
			return err
		}
		var netNames []string
		for _, n := range cfg.Networks {
			netNames = append(netNames, n.Name)
		}
		fmt.Printf("Please provide a network name with the -n flag.")
		fmt.Printf("The following Networks were found in your configuration file: " + "\n\n")
		for _, n := range netNames {
			fmt.Printf(n + "\n")
		}
		return err
	}

	if deviceFlow {
		token, err := loginDeviceFlow(name)
		if err != nil {
			return err
		}
		fmt.Printf(successStyle.Render("Success! Authenticated to Network."))
		// Store the token on the FS
		err = storeResponse(name, token)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func networkDetails(name string) (slc.Network, error) {

	if name == "" {
		return DecombineNetwork, nil
	}

	config, err := Config()
	if err != nil {
		return slc.Network{}, err
	}

	if len(config.Networks) == 0 {
		return slc.Network{}, fmt.Errorf("no networks configured. please configure a network first")
	}
	for _, n := range config.Networks {
		if n.Name == name {
			return n, nil
		}
	}

	return slc.Network{}, fmt.Errorf("network %s not found", name)
}

func defaultNetwork() (slc.Network, error) {
	cfg, err := Config()
	if err != nil {
		return slc.Network{}, err
	}
	for _, n := range cfg.Networks {
		if n.Name == cfg.DefaultNetwork {
			return n, nil
		}
	}
	return slc.Network{}, fmt.Errorf("default network not found")
}

func loginDeviceFlow(name string) (*oidc.AccessTokenResponse, error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer stop()

	network, err := networkDetails(name)
	if err != nil {
		return nil, err
	}

	var options []rp.Option
	provider, err := rp.NewRelyingPartyOIDC(ctx, network.Issuer, network.ClientID, "", "", scopes, options...)

	if err != nil {
		return nil, err
	}
	fmt.Printf(loginStyle.Render("Starting device flow authentication..."))
	resp, err := rp.DeviceAuthorization(ctx, scopes, provider, nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\nPlease browse to %s and enter code %s\n", resp.VerificationURI, highlight.Render(resp.UserCode))
	fmt.Printf("\nGotta go fast? Direct URL: %s"+"?user_code=%s\n", resp.VerificationURI, resp.UserCode)
	fmt.Printf("Waiting for authentication...\n")
	token, err := rp.DeviceAccessToken(ctx, resp.DeviceCode, time.Duration(resp.Interval)*time.Second, provider)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func loginPKCEFlow() {
	fmt.Printf("Starting PKCE authentication flow...")
	fmt.Printf("SYKE!")
}

func GetIDToken(name string) (string, error) {
	t, err := unmarshalToken(name)
	if err != nil {
		return "", err
	}
	return t.IDToken, nil
}

func unmarshalToken(name string) (*oidc.AccessTokenResponse, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	tokenDir := home + `/.config/contract/.` + name + `/`
	tokenPath := tokenDir + `token.json`
	tokenF, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer tokenF.Close()
	var (
		t oidc.AccessTokenResponse
		b []byte
	)
	b, err = io.ReadAll(tokenF)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func storeResponse(name string, token *oidc.AccessTokenResponse) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	tokenDir := home + `/.config/contract/.` + name + `/`
	err = os.MkdirAll(tokenDir, 0700)
	if err != nil {
		return err
	}
	tokenPath := tokenDir + `token.json`
	tokenF, err := os.Create(tokenPath)
	if err != nil {
		return err
	}
	defer tokenF.Close()

	resp, err := json.Marshal(token)
	if err != nil {
		return err
	}
	_, err = tokenF.Write(resp)
	if err != nil {
		return err
	}
	fmt.Printf("\nToken stored in %s", tokenPath)
	return nil
}
