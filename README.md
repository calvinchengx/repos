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
# export GITHUB_TOKEN="your personal access token"

$ source .env
$ go run main.go -user calvinchengx

# or
# go run main.go -org your_org_name

# see available flags
$ go run main.go -h
clone or pull multiple repositories given org name or username

Usage:
  repos [flags]

Flags:
  -d, --dir string    Directory where repositories should be cloned
  -h, --help          help for repos
  -o, --org string    GitHub organization name
  -u, --user string   GitHub username
```

### Non-interactive mode with configuration file

Every time we run `repos` in interactive mode, we will automatically save the repository-clone-directory mapping to a configuration file in a default configuration file `$HOME/.repos/repos.yaml`. 

The next time we run `repos -c` with `GITHUB_TOKEN` environment variable, it will automatically clone or pull all the repository-clone-directory mapping pairs in the configuration file.

This allows us to easily run all the `repos` commands in a single command.

```bash
# example with .env
$ source .env && repos -c

# example with doppler
$ doppler run -- repos -c
```

## Run Tests

```bash
$ go test -v ./...
```

## Build and Install

```bash
$ go build && go install

# The repos command is now available
$ repos

$ repos -h
clone or pull multiple repositories given org name or username

Usage:
  repos [flags]

Flags:
  -d, --dir string    Directory where repositories should be cloned
  -h, --help          help for repos
  -o, --org string    GitHub organization name
  -u, --user string   GitHub username
```