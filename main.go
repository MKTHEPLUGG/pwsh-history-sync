package main

import (
    "fmt"
    "os"
    "os/exec"
    "time"

    git "github.com/go-git/go-git/v5"
)

func main() {
    // Set up ticker to run every X minutes
    ticker := time.NewTicker(10 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        err := syncHistory()
        if err != nil {
            fmt.Println("Error syncing history:", err)
        }
    }
}

func syncHistory() error {
    repoPath := "/path/to/repo"
    historyFile := "/path/to/ConsoleHost_history.txt"

    // Open the repo (assumes it's already cloned)
    repo, err := git.PlainOpen(repoPath)
    if err != nil {
        return err
    }

    // Pull latest changes
    w, err := repo.Worktree()
    if err != nil {
        return err
    }

    err = w.Pull(&git.PullOptions{RemoteName: "origin"})
    if err != nil && err != git.NoErrAlreadyUpToDate {
        return err
    }

    // Read the local history file
    localHistory, err := os.ReadFile(historyFile)
    if err != nil {
        return err
    }

    // Compare and commit new history entries (this is just an example)
    // You can implement a more robust diffing method here
    if len(localHistory) > 0 {
        // Write new history to the repo file, commit, and push
        err = os.WriteFile("/path/to/repo/ConsoleHost_history.txt", localHistory, 0644)
        if err != nil {
            return err
        }

        _, err = w.Commit("Sync shell history", &git.CommitOptions{})
        if err != nil {
            return err
        }

        err = repo.Push(&git.PushOptions{})
        if err != nil {
            return err
        }

        fmt.Println("History synced successfully!")
    }

    return nil
}
