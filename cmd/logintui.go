package cmd

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"time"
)

type errMsg struct{ err error }
type statusMsg int

func (e errMsg) Error() string { return e.err.Error() }

type loginModel struct {
	authenticated bool
	spinner       spinner.Model
	status        int
	err           error
	quitting      bool
}

func initialLoginModel() loginModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return loginModel{spinner: s}
}

func (m loginModel) Init() tea.Cmd {
	m.spinner.Tick()
	return loginDeviceFlowWithTea
}

func (m loginModel) View() (s string) {
	if m.err != nil {
		return m.err.Error()
	}

	starting := loginStyle.Render("Starting device authentication flow...")
	str := fmt.Sprintf("\n\n   %s %v press q to quit\n\n", m.spinner.View(), starting)

	if m.quitting {
		return str + "\n"
	}
	return str
}

func (m loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.err
		return m, nil
	case statusMsg:
		// The server returned a status message. Save it to our model. Also
		// tell the Bubble Tea runtime we want to exit because we have nothing
		// else to do. We'll still be able to render a final view with our
		// status message.
		m.status = int(msg)
		m.quitting = true
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func loginDeviceFlowWithTea() tea.Msg {
	ctx := context.Background()
	var options []rp.Option
	//options = append(options, rp.WithPKCE(cookieHandler))
	provider, err := rp.NewRelyingPartyOIDC(ctx, DecombineNetwork.Issuer, DecombineNetwork.ClientID, "", "", scopes, options...)
	if err != nil {
		return errMsg{err}
	}

	resp, err := rp.DeviceAuthorization(ctx, scopes, provider, nil)
	if err != nil {
		return errMsg{err}
	}
	fmt.Printf("\nPlease browse to %s and enter code %s\n", resp.VerificationURI, resp.UserCode)
	fmt.Printf("Waiting for authentication...")
	token, err := rp.DeviceAccessToken(ctx, resp.DeviceCode, time.Duration(resp.Interval)*time.Second, provider)
	if err != nil {
		return errMsg{err}
	}
	fmt.Printf("Successfully obtained token: %#v", token)
	return statusMsg(1)
}
