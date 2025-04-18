package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
)

var validate *validator.Validate
var (
	contract  string
	contracts []string
)

var (
	ErrSourceNotSupported = errors.New("source type not supported")
)

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.PersistentFlags().StringArrayVarP(&contracts, "contracts", "c", []string{}, "Validate multiple Smart Legal Contracts via chaining --contracts flags or piping to stdin.")
	validateCmd.PersistentFlags().StringVar(&contract, "contract", "", "Validate a specific Smart Legal Contract. Stdin is also supported.")
	validateCmd.Flags().BoolP("report", "r", false, "Generate a report of the validation results.")
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a Smart Legal Contract",
	Long: `Validate a Smart Legal Contract to ensure it is correctly formatted and adheres to the Contract schema.

Validate will accept stdin for piping multiple Smart Legal Contracts to validate, individual Smart Legal Contracts using
the --contract flag, or even repeated --contracts flags to validate multiple Smart Legal Contracts.`,

	Run: func(cmd *cobra.Command, args []string) {

		//validate := validator.New(validator.WithRequiredStructEnabled())
		single, _ := cmd.Flags().GetString("contract")
		multi, _ := cmd.Flags().GetStringArray("contracts")

		if single == "" && len(multi) == 0 {
			if len(args) > 0 {
				for _, c := range args {
					s, err := getContractSource(c)
					if err != nil {
						fmt.Printf(ErrStyle.Render("error determining source type (file or url) of input %s: %v\n", c, err.Error()))
					}
					selectValidationStrategy(s, c)
				}
				return
			}
			//fmt.Printf(ErrStyle.Render("Error: no contracts to validate"))
		}

		if single != "" {
			s, err := getContractSource(single)
			if err != nil {
				fmt.Printf(ErrStyle.Render("error determining source type (file or url) of input %s: %v\n", single, err.Error()))
			}
			selectValidationStrategy(s, single)
			return
		}

		if len(multi) > 0 {
			for _, c := range multi {
				s, err := getContractSource(c)
				if err != nil {
					fmt.Printf(ErrStyle.Render("error determining source type (file or url) of input %s: %v\n", c, err.Error()))
				}
				selectValidationStrategy(s, c)
				return
			}
		}

		if os.Stdin != nil {
			fmt.Printf("Stdin is not nil\n")
		}
		// Set up a buffer so we can accept input from stdin
		// for the scenario where multiple Smart Legal Contracts
		// are piped to the validate command.
		reader := bufio.NewReader(cmd.InOrStdin())
		var inputs []string
		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, "\r\n")

		// Each Smart Legal Contract should be separated by a space.
		cmdPieces := strings.Split(text, " ")

		for _, piece := range cmdPieces {

			if strings.HasPrefix(piece, "-") {
				// TODO: Handle flags
				continue
			}

			inputs = append(inputs, piece)
			s, err := getContractSource(piece)
			if err != nil {
				fmt.Printf(ErrStyle.Render("error determining source type (file or url) of input %s: %v\n", s, err.Error()))
				return
			}
			selectValidationStrategy(s, piece)

		}

	},
}

// getContractSource determines the source type of the Smart Legal Contract based on the input value.
// The source type can be either a URL or a filesystem path.
func getContractSource(source string) (string, error) {

	if source == "" {
		return "", ErrSourceNotSupported
	}
	if strings.HasPrefix(source, "http") {
		return "url", nil
	}
	f, err := os.Open(source)
	if err != nil {
		return "", err
	}
	_ = f.Close()
	return "fs", nil
}

func selectValidationStrategy(source, path string) {
	switch source {
	case "fs":
		validateFSContract(path)
	case "url":
		validateURLContract(path)
	}
}

// validateFSContract validates a Smart Legal Contract from the filesystem.
func validateFSContract(path string) {
	fmt.Printf("Validating FS contract at %s\n", path)
}

// validateURLContract validates a Smart Legal Contract from a URL.
func validateURLContract(url string) {
	fmt.Printf("Validating URL contract at %s\n", url)
}
