# repos

Simple command line utility to manage multiple github repos.

## Usage

### Interactive mode

```bash
$ go run main.go

Enter GitHub organization name (leave empty if cloning for a username): 
Enter GitHub username (leave empty if cloning for an organization): calvinchengx
Enter personal access token: ****************************************
Enter the directory where repositories should be cloned: /Users/calvin/calvinchengx
(if empty, repositories will use the default path as user home and orgName or username subdirectory): 
```

### Non-interactive mode

```bash
# place your github personal access token a .env file
# GITHUB_TOKEN="your personal access token"

$ source .env
$ go run main.go -user calvinchengx

# or
# go run main.go -org your_org_name

# see available flags
$ go run main.go -h
Usage of repos:
  -dir string
        Directory where repositories should be cloned
  -org string
        GitHub organization name
  -user string
        GitHub username
```


## Run Tests

```bash
$ go test -v
```

## Build and Install

```bash
$ go build && go install

# The repos command is now available
$ repos

$ repos -h
Usage of repos:
  -dir string
    	Directory where repositories should be cloned
  -org string
    	GitHub organization name
  -user string
    	GitHub username
```