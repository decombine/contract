package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/decombine/slc"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

type initContract struct {
	Name      string
	SourceURL string
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.AddCommand(initConfigFileCmd)
	initCmd.AddCommand(initSLCCmd)
	initCmd.AddCommand(initProjectCmd)
	initSLCCmd.PersistentFlags().String("name", "My Contract", "The name of the Smart Legal Contract")
	initSLCCmd.PersistentFlags().String("url", "https://github.com/myorg/myrepo", "The Git URL of the Smart Legal Contract")
	initSLCCmd.Flags().BoolP("with-project", "g", false, "Include an example project structure for a Git repository")
	initSLCCmd.Flags().StringP("output", "o", "json", "The output format of the Smart Legal Contract. Options: json, toml, yaml")
	initCmd.Flags().StringP("path", "p", "", "The output path")

	initProjectCmd.PersistentFlags().String("name", "My Contract", "The name of the Smart Legal Contract")
	initProjectCmd.PersistentFlags().String("url", "https://github.com/myorg/myrepo", "The Git URL of the Smart Legal Contract")
	initProjectCmd.MarkPersistentFlagRequired("name")
	initProjectCmd.MarkPersistentFlagRequired("url")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new Smart Legal Contract template",
	Long:  `Generate templates to get started with Contract development.`,
}

var initConfigFileCmd = &cobra.Command{
	Use:   "config",
	Short: "Create a new Contract CLI configuration file",
	Long:  "Create a Contract CLI configuration file to store your preferred settings for the Contract CLI.",
	Run:   initConfigFileExecute,
}

var initSLCCmd = &cobra.Command{
	Use:   "slc",
	Short: "Create a new Smart Legal Contract template",
	Long:  `Generate templates to get started with Contract development.`,
	RunE:  initExecute,
}

var initProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Create a new Smart Legal Contract project with templates",
	Long:  `Generate a recommended project structure to get started with Contract development.`,
	RunE:  initProjectExecute,
}

func initExecute(cmd *cobra.Command, args []string) error {
	in := initContract{
		Name:      "My Contract",
		SourceURL: "https://github.com/myorg/myrepo",
	}
	c := makeContract(&in)
	p, _ := cmd.Flags().GetString("path")
	proj, _ := cmd.Flags().GetBool("with-project")

	format, _ := cmd.Flags().GetString("output")

	// If the user has requested a project structure, create the project
	// and return early.
	if proj {
		if err := initProject(in.Name, c); err != nil {
			return fmt.Errorf("error creating project: %v", err)
		}
		return nil
	}

	switch format {
	case "json":
		if err := makeJSONContract(p, c); err != nil {
			return fmt.Errorf("error creating contract: %v", err)
		}
		fmt.Printf("Contract created at %s\n", p+"contract.json")
		return nil
	case "toml":
		if err := makeTOMLContract(p, c); err != nil {
			return fmt.Errorf("error creating contract: %v", err)
		}
		fmt.Printf("Contract created at %s\n", p+"contract.toml")
		return nil
	case "yaml":
		if err := makeYAMLContract(p, c); err != nil {
			return fmt.Errorf("error creating contract: %v", err)
		}
		fmt.Printf("Contract created at %s\n", p+"contract.yaml")
		return nil
	}

	return nil
}

func initConfigFileExecute(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("path")
	err := makeYAMLConfigFile(path)
	if err != nil {
		fmt.Println("Error creating config file:", err)
	}
}

func initProjectExecute(cmd *cobra.Command, args []string) error {
	var in initContract
	in.Name, _ = cmd.Flags().GetString("name")
	in.SourceURL, _ = cmd.Flags().GetString("url")
	c := makeContract(&in)
	if err := initProject(in.Name, c); err != nil {
		return fmt.Errorf("error creating project: %v", err)
	}
	return nil
}

func makeConfigDir() error {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	err = os.MkdirAll(home+ConfigPath, 0700)
	if err != nil {
		return err
	}
	return nil
}

func makeYAMLConfigFile(path string) error {
	if path == "" {
		path = "contract.yaml"
	}
	var networks []slc.Network
	networks = append(networks, DecombineNetwork)
	newConfig := ContractCLIConfig{
		Language:                "en",
		DefaultContractFileType: "json",
		DefaultNetwork:          "decombine",
		Networks:                networks,
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := yaml.Marshal(newConfig)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func initProject(n string, c *slc.Contract) error {
	// Create a root directory based on the passed in name that will be compatible
	// with Git.
	rootDirName := strings.ReplaceAll(strings.ToLower(n), " ", "-")
	err := makeProjectDirs(rootDirName)
	if err != nil {
		return fmt.Errorf("error creating project directories: %v", err)
	}

	rootDirName += "/"

	err = makeHTMLFile(rootDirName + "text/")
	if err != nil {
		return fmt.Errorf("error creating Contract Text %v", err)
	}

	if err = makeJSONContract(rootDirName+"contracts/", c); err != nil {
		return fmt.Errorf("error creating contract: %v", err)
	}

	if err = makeTOMLContract(rootDirName+"contracts/", c); err != nil {
		return fmt.Errorf("error creating contract: %v", err)
	}

	if err = makeYAMLContract(rootDirName+"contracts/", c); err != nil {
		return fmt.Errorf("error creating contract: %v", err)
	}

	if err = makeProjectREADMEFiles(rootDirName); err != nil {
		return fmt.Errorf("error creating README: %v", err)
	}

	enumeratorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).MarginRight(1)
	rootStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#227BF0"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("35"))

	t := tree.Root(rootDirName).
		Child(
			tree.New().
				Root("/contracts").
				RootStyle(rootStyle).
				Child("contract.json").
				Child("contract.toml").
				Child("contract.yaml").
				Child("README.md"),
		).
		Child("/policies").
		Child(
			tree.New().Root("/text").
				RootStyle(rootStyle).
				Child("index.html"),
		).
		Child("README.md").
		Enumerator(tree.RoundedEnumerator).
		EnumeratorStyle(enumeratorStyle).
		RootStyle(rootStyle).
		ItemStyle(itemStyle)

	fmt.Println(t)

	return nil
}

// makeJSONContract creates a sample JSON-based Smart Legal Contract file.
func makeJSONContract(p string, c *slc.Contract) error {
	f, _ := os.Create(p + "contract.json")
	defer f.Close()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling contract.json: %v", err)
	}
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("error writing contract file: %v", err)
	}
	return nil
}

// makeTOMLContract creates a sample TOML-based Smart Legal Contract file.
func makeTOMLContract(p string, c *slc.Contract) error {
	f, _ := os.Create(p + "contract.toml")
	defer f.Close()
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshalling contract.toml: %v", err)
	}
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("error writing contract file: %v", err)
	}
	return nil
}

// makeYAMLContract creates a sample YAML-based Smart Legal Contract file.
func makeYAMLContract(p string, c *slc.Contract) error {
	f, _ := os.Create(p + "contract.yaml")
	defer f.Close()
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshalling contract.yaml: %v", err)
	}
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("error writing contract file: %v", err)
	}
	return nil
}

func makeProjectDirs(n string) error {
	err := os.MkdirAll(n, 0700)
	if err != nil {
		return fmt.Errorf("error creating project directory: %v", err)
	}

	for _, d := range []string{"contracts", "policies", "text"} {
		err = os.MkdirAll(n+"/"+d, 0700)
		if err != nil {
			return fmt.Errorf("error creating project directory: %v", err)
		}
	}
	return nil
}

func makeHTMLFile(p string) error {
	f, err := os.Create(p + "index.html")
	if err != nil {
		return fmt.Errorf("error creating project directory: %v", err)
	}
	defer f.Close()
	// TODO: Write some example text to the file
	return nil
}

func makeREADMEFile(p string) error {
	f, err := os.Create(p + "README.md")
	if err != nil {
		return fmt.Errorf("error creating README.md file: %v", err)
	}
	defer f.Close()
	_, err = f.WriteString(local("InitREADME"))
	if err != nil {
		return fmt.Errorf("error writing to README.md: %v", err)
	}
	return nil
}

func makeProjectREADMEFiles(p string) error {
	err := makeREADMEFile(p)
	if err != nil {
		return fmt.Errorf("error creating README.md file: %v", err)
	}
	cf, err := os.Create(p + "/contracts/" + "README.md")
	if err != nil {
		return fmt.Errorf("error creating README.md file: %v", err)
	}
	defer cf.Close()
	_, err = cf.WriteString(local("InitContractREADME"))
	if err != nil {
		return fmt.Errorf("error writing to README.md: %v", err)
	}
	return nil
}

func makeContract(in *initContract) *slc.Contract {
	c := &slc.Contract{
		Version: slc.Version,
		Name:    in.Name,
		Source: slc.GitSource{
			URL:    in.SourceURL,
			Branch: "main",
			Path:   "contract.json",
		},
		Text: slc.ContractText{
			URL: in.SourceURL + "/text/index.html",
		},
		Policy: slc.PolicySource{
			URL:       in.SourceURL,
			Branch:    "main",
			Directory: "/policies",
		},
		State: slc.StateConfiguration{
			Initial: "Draft",
			URL:     in.SourceURL,
			States: []slc.State{
				{
					Name: "Draft",
					Entry: slc.Action{
						ActionType: "KubernetesAction",
						KubernetesActions: []slc.KubernetesAction{
							{
								Namespace: "default",
								KustomizationSpec: &kustomizev1.KustomizationSpec{
									Path:       "contracts/workloads/draft",
									Prune:      true,
									NamePrefix: "draft-",
								},
							},
						},
					},
					Exit: slc.Action{},
					Transitions: []slc.Transition{
						{
							Name: "Signing",
							To:   "In Process",
							On:   "com.decombine.signature.sign",
							Conditions: []slc.Condition{
								{
									Name:  "data.signature.validated",
									Value: "true",
								},
							},
						},
						{
							Name: "Expired",
							To:   "Expired",
							On:   "com.decombine.contract.expirationReached",
						},
					},
				},
			},
		},
	}

	return c
}
