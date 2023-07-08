package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/calvinchengx/repos/github"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	orgName     string
	username    string
	accessToken string
	cloneDir    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "repos",
	Short: "clone or pull multiple repositories given org name or username",
	Long:  `clone or pull multiple repositories given org name or username`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		accessToken = os.Getenv("GITHUB_TOKEN")

		if cmd.Flags().NFlag() > 0 {

			if cmd.Flags().Changed("config") {
				if len(cfgFile) == 1 && cfgFile == " " {
					cfgFile = configPath()
				}
				v := viper.New()
				v.SetConfigType("yaml")
				v.SetConfigFile(cfgFile)
				err := v.ReadInConfig()
				if err != nil {
					log.Fatalf("Error reading config file: %s", err)
				}

				// if GITHUB_TOKEN is available, we do not need to ask for it
				if accessToken == "" {
					fmt.Print("Enter personal access token: ")
					accessTokenBytes, err := gopass.GetPasswdMasked()
					if err != nil {
						log.Fatalf("Error reading access token: %s", err)
					}
					accessToken = strings.TrimSpace(string(accessTokenBytes))
				}

				// if config file is found and has values, we will run the command as a loop with specified values
				pairs := v.GetStringSlice("pairs")
				for _, pair := range pairs {
					orgName, username, cloneDir = parsePair(pair)
					repositories, err := github.GetRepositories("https://api.github.com", orgName, username, accessToken)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Printf("Clone or pull repositories into %s \n", cloneDir)
					github.CloneRepositories(repositories, cloneDir)
					fmt.Printf("Total number of repositories: %d\n", len(repositories))

				}
				return
			}

			if cmd.Flags().Changed("org") && cmd.Flags().Changed("user") {
				log.Fatal("Please provide either an organization name or a username, not both.")
			}
			if accessToken == "" {
				log.Fatal("Error: GITHUB_TOKEN environment variable is not set.")
			}
		} else {
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Enter GitHub organization name (leave empty if cloning for a username): ")
			orgName, _ = reader.ReadString('\n')
			orgName = strings.TrimSpace(orgName)

			if orgName == "" {
				fmt.Print("Enter GitHub username (leave empty if cloning for an organization): ")
				username, _ = reader.ReadString('\n')
				username = strings.TrimSpace(username)
			}

			if orgName != "" && username != "" {
				log.Fatal("Please provide either an organization name or a username, not both.")
			}

			// if GITHUB_TOKEN is available, we do not need to ask for it
			if accessToken == "" {
				fmt.Print("Enter personal access token: ")
				accessTokenBytes, err := gopass.GetPasswdMasked()
				if err != nil {
					log.Fatalf("Error reading access token: %s", err)
				}
				accessToken = strings.TrimSpace(string(accessTokenBytes))
			}

			fmt.Print("Enter the directory where repositories should be cloned \n(if empty, repositories will use the default path as user home and orgName or username subdirectory): ")
			cloneDir, _ = reader.ReadString('\n')
			cloneDir = strings.TrimSpace(cloneDir)
		}

		if cloneDir == "" {
			homeDir, err := os.UserHomeDir() // this works on Windows, macOS, and Linux
			if err != nil {
				fmt.Println("Failed to get home directory:", err)
				return
			}
			if orgName != "" {
				cloneDir = filepath.Join(homeDir, orgName)
			} else {
				cloneDir = filepath.Join(homeDir, username)
			}
		}

		repositories, err := github.GetRepositories("https://api.github.com", orgName, username, accessToken)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Clone or pull repositories into %s \n", cloneDir)
		github.CloneRepositories(repositories, cloneDir)

		fmt.Printf("Total number of repositories: %d\n", len(repositories))

		configUpdate(cfgFile, orgName, username, cloneDir)
	},
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", configPath(), "Path to config file")
	rootCmd.PersistentFlags().Lookup("config").NoOptDefVal = " " // allows us to not provide any value for config flag
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&orgName, "org", "o", "", "GitHub organization name")
	rootCmd.PersistentFlags().StringVarP(&username, "user", "u", "", "GitHub username")
	rootCmd.PersistentFlags().StringVarP(&cloneDir, "dir", "d", "", "Directory where repositories should be cloned")
	rootCmd.Flags().SortFlags = false
}

// use yaml config file if it is available
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigFile(filepath.Join(configPath()))
	}

	viper.ReadInConfig()
}

func configPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configDir := filepath.Join(homeDir, ".repos")
	return filepath.Join(configDir, "repos.yaml")
}

func configUpdate(cfgFile string, orgName string, username string, cloneDir string) error {

	// Create config directory if it doesn't exist
	configDir := path.Dir(cfgFile)
	// Create directories recursively
	err := os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		return err
	}

	// Check if file exists
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		// File doesn't exist, create it
		file, err := os.Create(cfgFile)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return err
		}
		defer file.Close()

		fmt.Println("File created successfully.")
	} else if err != nil {
		fmt.Println("Error checking file existence:", err)
		return err
	} else {
		fmt.Println("File already exists.")
	}

	// update file with values
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(cfgFile)
	err = v.ReadInConfig()
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return err
	}

	pairs := v.GetStringSlice("pairs")
	fmt.Println("existing list:", pairs)
	newPairs := []string{}
	newItem := ""

	if orgName != "" {
		newItem = "org:" + orgName + ":" + cloneDir
		if !hasDuplicate(pairs, newItem) {
			newPairs = append(pairs, newItem)
		} else {
			newPairs = pairs
		}
	} else if username != "" {
		newItem = "user:" + username + ":" + cloneDir
		if !hasDuplicate(pairs, newItem) {
			newPairs = append(pairs, newItem)
		} else {
			newPairs = pairs
		}
	}
	fmt.Println("new list:", newPairs)
	v.Set("pairs", newPairs)

	err = v.WriteConfig()
	if err != nil {
		fmt.Println("Error writing config file:", err)
		return err
	}

	return nil
}

func hasDuplicate(list []string, newItem string) bool {
	for _, item := range list {
		if item == newItem {
			return true
		}
	}
	return false
}

func parsePair(pair string) (orgName string, username string, cloneDir string) {

	parts := strings.Split(pair, ":")
	if parts[0] == "org" {
		return parts[1], "", parts[2]
	} else if parts[0] == "user" {
		return "", parts[1], parts[2]
	}

	return "", "", ""
}
