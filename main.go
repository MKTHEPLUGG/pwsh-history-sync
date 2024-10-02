package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "log"
    "runtime"

    git "gopkg.in/src-d/go-git.v4"
    "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
    "gopkg.in/src-d/go-git.v4/config"
    "gopkg.in/yaml.v2"
)

// Config struct to hold Git credentials from YAML
type Config struct {
    Git struct {
        Username string `yaml:"username"`
        Token    string `yaml:"token"`
        Repo     string `yaml:"repo"`
    } `yaml:"git"`
}

// Global-like variable for the PowerShell history file path
var historyFilePath string
var gitRepoPath string
var homeDir string
var configFilePath string

func init() {
    // Get the APPDATA environment variable
    appDataPath := os.Getenv("APPDATA")

    if appDataPath == "" {
        fmt.Println("APPDATA environment variable is not set.")
        return
    }

    // Build the full path to the PowerShell history file
    historyFilePath = filepath.Join(appDataPath, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt")

    // Set the Git repository path to the same directory as the history file
    gitRepoPath = filepath.Join(appDataPath, "Microsoft", "Windows", "PowerShell", "PSReadLine")
}

func main() {
    log.Println("Setting the users home directory")
    homeDir := getHomeDir()
    if homeDir == "" {
        fmt.Println("Could not determine the home directory.")
    } else {
        fmt.Println("User's home directory is:", homeDir)
    }

    log.Println("Checking if history file path was set by init function")
    if historyFilePath == "" {
        fmt.Println("Failed to set history file path.")
        return
    }

    // Load Git credentials
    username, password, repo, err := loadCredentials()
    if err != nil {
        fmt.Printf("Error loading credentials: %s\n", err)
        return
    }

    // Link the local repo with the remote and pull changes
    err = linkAndPullFromRemote(username, password, repo)
    if err != nil {
        fmt.Printf("Error pulling from the remote repository: %s\n", err)
    } else {
        fmt.Println("Repository linked and pulled successfully.")
    }
}

// getHomeDir gets the user's home directory based on the operating system.
func getHomeDir() string {
    if runtime.GOOS == "windows" {
        // On Windows, use USERPROFILE or HOMEPATH
        home := os.Getenv("USERPROFILE")
        if home == "" {
            home = os.Getenv("HOMEPATH")
        }
        if home == "" {
            fmt.Println("Could not determine the home directory on Windows.")
        }
        return home
    }

    // On Linux/macOS, use HOME environment variable
    return os.Getenv("HOME")
}


// loadCredentials loads credentials from environment variables or config file
func loadCredentials() (string, string, string, error) {
    configFilePath = filepath.Join(homeDir, ".config", "config.yaml")
    // Check if credentials are set in environment variables
    username := os.Getenv("GIT_USERNAME")
    password := os.Getenv("GIT_TOKEN")
    repo := os.Getenv("GIT_REPO")

    if username != "" && password != "" && repo != "" {
        fmt.Println("Credentials loaded from environment variables.")
        return username, password, repo, nil
    }

    // If environment variables are not set, load from config file
    fmt.Println("Loading credentials from config file.")
    config, err := loadConfig(configFilePath)
    if err != nil {
        return "", "", "", err
    }

    return config.Git.Username, config.Git.Token, config.Git.Repo, nil
}

// loadConfig reads and parses the YAML config file
func loadConfig(filePath string) (*Config, error) {
    config := &Config{}

    // Read the config file
    fileData, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    // Unmarshal the YAML into the Config struct
    err = yaml.Unmarshal(fileData, config)
    if err != nil {
        return nil, err
    }

    return config, nil
}

// linkAndPullFromRemote links to the remote and pulls changes
func linkAndPullFromRemote(username, password, repoURL string) error {
    // First, check if the directory is already a Git repository
    repo, err := git.PlainOpen(gitRepoPath)
    if err != nil {
        fmt.Println("Directory is not a Git repository. Initializing a new Git repository.")
        repo, err = git.PlainInit(gitRepoPath, false)
        if err != nil {
            return fmt.Errorf("failed to initialize Git repository: %w", err)
        }
    } else {
        fmt.Println("Directory is already a Git repository.")
    }

    // Check if the remote is already set
    remotes, err := repo.Remotes()
    if err != nil {
        return fmt.Errorf("failed to list remotes: %w", err)
    }

    remoteExists := false
    for _, remote := range remotes {
        if remote.Config().Name == "origin" {
            remoteExists = true
            fmt.Println("Remote 'origin' is already set.")
            break
        }
    }

    if !remoteExists {
        // Add the remote for pulling
        remoteConfig := fmt.Sprintf("https://%s:%s@%s", username, password, repoURL)

        _, err = repo.CreateRemote(&config.RemoteConfig{
            Name: "origin",
            URLs: []string{remoteConfig},
        })
        if err != nil {
            return fmt.Errorf("failed to add remote: %w", err)
        }
        fmt.Println("Remote 'origin' added successfully.")
    }

    // Now pull from the remote
    err = pullFromRemote(repo, username, password)
    if err != nil {
        return fmt.Errorf("failed to pull from remote: %w", err)
    }

    return nil
}

// pullFromRemote pulls the latest changes from the remote repository
func pullFromRemote(repo *git.Repository, username, password string) error {
    worktree, err := repo.Worktree()
    if err != nil {
        return fmt.Errorf("failed to get worktree: %w", err)
    }

    // Pull the latest changes from the origin
    err = worktree.Pull(&git.PullOptions{
        RemoteName: "origin",
        Auth: &http.BasicAuth{
            Username: username, // GitHub username
            Password: password, // Personal access token
        },
    })

    if err != nil && err == git.NoErrAlreadyUpToDate {
        fmt.Println("Already up to date.")
        return nil
    } else if err != nil {
        return fmt.Errorf("failed to pull from remote: %w", err)
    }

    fmt.Println("Pulled latest changes from remote.")
    return nil
}


// --------------------------------------------------------------------------------------------------------------------- //

// func syncHistory() error {
//     repoPath := "/path/to/repo"
//     historyFile := "/path/to/ConsoleHost_history.txt"
//
//     // Open the repo (assumes it's already cloned)
//     repo, err := git.PlainOpen(repoPath)
//     if err != nil {
//         return err
//     }
//
//     // Pull latest changes
//     w, err := repo.Worktree()
//     if err != nil {
//         return err
//     }
//
//     err = w.Pull(&git.PullOptions{RemoteName: "origin"})
//     if err != nil && err != git.NoErrAlreadyUpToDate {
//         return err
//     }
//
//     // Read the local history file
//     localHistory, err := os.ReadFile(historyFile)
//     if err != nil {
//         return err
//     }
//
//     // Compare and commit new history entries (this is just an example)
//     // You can implement a more robust diffing method here
//     if len(localHistory) > 0 {
//         // Write new history to the repo file, commit, and push
//         err = os.WriteFile("/path/to/repo/ConsoleHost_history.txt", localHistory, 0644)
//         if err != nil {
//             return err
//         }
//
//         _, err = w.Commit("Sync shell history", &git.CommitOptions{})
//         if err != nil {
//             return err
//         }
//
//         err = repo.Push(&git.PushOptions{})
//         if err != nil {
//             return err
//         }
//
//         fmt.Println("History synced successfully!")
//     }
//
//     return nil
// }

//    // Set up ticker to run every X minutes
//     ticker := time.NewTicker(10 * time.Minute)
//     defer ticker.Stop()

//     for range ticker.C {
//         err := syncHistory()
//         if err != nil {
//             fmt.Println("Error syncing history:", err)
//         }
//     }

//
// package main
//
// import (
//     "fmt"
//     "os"
//     "path/filepath"
//     "gopkg.in/src-d/go-git.v4"          // Import the go-git package
//     "gopkg.in/src-d/go-git.v4/plumbing/transport/http" // For HTTP authentication
// )
//
// // Global-like variable for the PowerShell history file path
// var historyFilePath string
//
// // init function to initialize the global-like variable
// func init() {
//     // Get the APPDATA environment variable
//     appDataPath := os.Getenv("APPDATA")
//
//     if appDataPath == "" {
//         fmt.Println("APPDATA environment variable is not set.")
//         return
//     }
//
//     // Build the full path to the PowerShell history file
//     historyFilePath = filepath.Join(appDataPath, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt")
// }
//
// func main() {
//     if historyFilePath == "" {
//         fmt.Println("Failed to set history file path.")
//         return
//     }
//
//     // Now historyFilePath is accessible throughout your program
//     fmt.Println("PowerShell History Path:", historyFilePath)
//     // You can use historyFilePath here and in other functions
//
//     err := cloneGit()
//     if err != nil {
//         fmt.Printf("Error cloning the repository: %s\n", err)
//     } else {
//         fmt.Println("Repository cloned successfully.")
//     }
// }
//
//
// func cloneGit() error {
//     // Set the Git credentials (username and personal access token)
//     username := "your-username" // Replace with your GitHub username
//     password := "your-token"    // Replace with your personal access token
//     repo := "github.com/your-org/your-repo.git" // Replace with your repository URL
//
//     // Construct the URL for HTTPS clone
//     url := fmt.Sprintf("https://%s:%s@%s", username, password, repo)
//
//     // Clone options, including progress display and authentication
//     options := &git.CloneOptions{
//         URL:      url,
//         Progress: os.Stdout,
//         Auth: &http.BasicAuth{
//             Username: username, // Username must be provided even if using a token
//             Password: password, // This is your personal access token
//         },
//     }
//
//     // Clone the repository into the ./src directory
//     r, err := git.PlainClone("./src", false, options)
//     if err != nil {
//         return err
//     }
//
//     // Optionally, log the head of the cloned repository
//     head, err := r.Head()
//     if err != nil {
//         return err
//     }
//
//     fmt.Printf("Repository cloned to: ./src\nCurrent branch: %s\n", head.Name())
//     return nil
// }
