package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/howeyc/gopass"
)

type Repository struct {
	Name     string `json:"name"`
	CloneURL string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
}

func getRepositories(apiBaseURL, orgName, username, accessToken string) ([]Repository, error) {
	var url string
	if orgName != "" {
		url = fmt.Sprintf("%s/orgs/%s/repos", apiBaseURL, orgName)
	} else if username != "" {
		url = fmt.Sprintf("%s/users/%s/repos", apiBaseURL, username)
	} else {
		return nil, fmt.Errorf("orgName or username is required")
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	q := req.URL.Query()
	q.Add("type", "all")
	q.Add("per_page", "100")
	req.URL.RawQuery = q.Encode()

	client := http.DefaultClient
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var repositories []Repository
	if err := json.NewDecoder(response.Body).Decode(&repositories); err != nil {
		return nil, err
	}

	return repositories, nil
}

func cloneRepositories(repositories []Repository, cloneDir string) {

	_, err := os.Stat(cloneDir)
	if os.IsNotExist(err) {
		// directory does not exist, create it
		err := os.MkdirAll(cloneDir, 0755)
		if err != nil {
			fmt.Println("Failed to create directory:", err)
			return
		}
		fmt.Println("Directory created successfully.")
	} else if err != nil {
		// an error coccured while checking the directory
		fmt.Println("Failed to create directory:", err)
		return
	}

	var wg sync.WaitGroup
	cloneChan := make(chan Repository)

	// Start goroutines for cloning repositories
	for _, repo := range repositories {
		wg.Add(1)
		go func(repo Repository) {
			defer wg.Done()

			repoDir := filepath.Join(cloneDir, repo.Name)
			if _, err := os.Stat(repoDir); err == nil {
				// Directory already exists, perform git pull
				fmt.Printf("Repository %s already exists. Performing git pull...\n", repo.Name)
				cmd := exec.Command("git", "pull", "--all")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Dir = repoDir
				if err := cmd.Run(); err != nil {
					log.Printf("Error pulling repository %s: %s\n", repo.Name, err.Error())
				} else {
					fmt.Printf("Pulled repository: %s\n", repo.Name)
				}
			} else {
				// Directory doesn't exist, clone the repository
				fmt.Printf("Cloning repository %s...\n", repo.Name)
				cmd := exec.Command("git", "clone", repo.SSHURL)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Dir = cloneDir
				if err := cmd.Run(); err != nil {
					log.Printf("Error cloning repository %s: %s\n", repo.Name, err.Error())
				} else {
					fmt.Printf("Cloned repository: %s\n", repo.Name)
				}
			}
		}(repo)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(cloneChan)
}

func main() {
	var (
		orgName     string
		username    string
		accessToken string
		cloneDir    string
	)

	// Define command line flags
	orgNameFlag := flag.String("org", "", "GitHub organization name")
	usernameFlag := flag.String("user", "", "GitHub username")
	cloneDirFlag := flag.String("dir", "", "Directory where repositories should be cloned")

	// Parse command line arguments
	flag.Parse()

	if flag.NFlag() > 0 {
		fmt.Println("Using command line flags")
		orgName = *orgNameFlag
		username = *usernameFlag
		accessToken = os.Getenv("GITHUB_TOKEN")
		cloneDir = *cloneDirFlag
		if orgName != "" && username != "" {
			log.Fatal("Please provide either an organization name or a username, not both.")
		}
		if accessToken == "" {
			fmt.Println("Error: GITHUB_TOKEN environment variable is not set.")
			os.Exit(1)
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

		fmt.Print("Enter personal access token: ")
		accessTokenBytes, err := gopass.GetPasswdMasked()
		if err != nil {
			log.Fatalf("Error reading access token: %s", err)
		}
		accessToken = strings.TrimSpace(string(accessTokenBytes))

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

	repositories, err := getRepositories("https://api.github.com", orgName, username, accessToken)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Clone or pull repositories into %s \n", cloneDir)
	cloneRepositories(repositories, cloneDir)

	fmt.Printf("Total number of repositories: %d\n", len(repositories))
}
