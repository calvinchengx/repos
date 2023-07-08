package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/calvinchengx/repos/github"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
)

var (
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

		if cmd.Flags().NFlag() > 0 {
			accessToken = os.Getenv("GITHUB_TOKEN")
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
			accessToken = os.Getenv("GITHUB_TOKEN")
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
	rootCmd.PersistentFlags().StringVarP(&orgName, "org", "o", "", "GitHub organization name")
	rootCmd.PersistentFlags().StringVarP(&username, "user", "u", "", "GitHub username")
	rootCmd.PersistentFlags().StringVarP(&cloneDir, "dir", "d", "", "Directory where repositories should be cloned")
}
