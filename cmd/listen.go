package main

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sshst/sshststuff"
)

var listenConfig client.ListenConfig

func listenCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for inbound connections",
		Run:   listen,
	}

	listenCmd.PersistentFlags().StringVar(&listenConfig.APIURL, "api", "https://api.ssh.st/api/listeners", "")
	listenCmd.PersistentFlags().MarkHidden("api")

	listenCmd.PersistentFlags().IntVar(&listenConfig.IdleTimeout, "idle-timeout", 30, "")
	listenCmd.PersistentFlags().BoolVar(&listenConfig.WebOk, "web-ok", true, "")
	// TODO command
	listenCmd.PersistentFlags().StringVar(&listenConfig.NotifyUser, "notify-user", "", "")
	listenCmd.PersistentFlags().StringVar(&listenConfig.NotifyTitle, "notify-title", "", "")
	listenCmd.PersistentFlags().StringSliceVarP(&listenConfig.GithubUsers, "github", "g", []string{}, "")
	listenCmd.PersistentFlags().StringSliceVarP(&listenConfig.SSHFingerprints, "fingerprint", "f", []string{}, "")

	codebuildCmd := &cobra.Command{Use: "codebuild", Run: codebuild}
	codebuildCmd.PersistentFlags().Bool("always", false, "")
	listenCmd.AddCommand(codebuildCmd)

	githubCmd := &cobra.Command{Use: "github", Run: github}
	listenCmd.AddCommand(githubCmd)

	return listenCmd
}

func listen(cmd *cobra.Command, args []string) {
	if len(listenConfig.NotifyUser) == 0 {
		listenConfig.NotifyUser = headCommitAuthor()
	}

	listenConfig.Version = version
	addGithubSshKeys(&listenConfig)
	
	err := client.Listen(context.Background(), listenConfig)
	if err != nil {
		panic(err)
	}
}

func addGithubSshKeys(listenConfig *client.ListenConfig) {
	warn := func(username, msg string) {
		fmt.Fprintf(os.Stderr, "Error getting SSH key for %s: %s\n", username, msg)
	}

	for _, username := range listenConfig.GithubUsers {
		resp, err := http.Get(fmt.Sprintf("https://github.com/%s.keys", username))
		if err != nil {
			warn(username, err.Error())
			continue
		}

		authorizedKeysBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			warn(username, err.Error())
			continue
		}

		lines := bytes.Split(authorizedKeysBytes, []byte("\n"))
		for _, line := range lines {
			pubkey, _, _, _, err := ssh.ParseAuthorizedKey(line)
			if err != nil {
				warn(username, err.Error())
				continue
			}

			fingerprint := ssh.FingerprintSHA256(pubkey)
			listenConfig.SSHFingerprints = append(listenConfig.SSHFingerprints, fingerprint)
			fmt.Fprintf(os.Stderr, "Trusting fingerprint %s for GitHub user %s\n", fingerprint, username)
		}
	}
}

func codebuild(cmd *cobra.Command, args []string) {
	always, _ := cmd.PersistentFlags().GetBool("always")
	if os.Getenv("CODEBUILD_BUILD_SUCCEEDING") == "1" && !always {
		return
	}

	if len(listenConfig.NotifyTitle) == 0 {
		id := os.Getenv("CODEBUILD_BUILD_ID")
		split := strings.SplitN(id, ":", 2)
		listenConfig.NotifyTitle = split[0]
	}

	listen(cmd, args)
}

func headCommitAuthor() string {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ""
	}

	ref, err := repo.Head()
	if err != nil {
		return ""
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return ""
	}

	commit, err := cIter.Next()
	if err != nil {
		return ""
	}

	return commit.Author.Email
}

func github(cmd *cobra.Command, args []string) {
	listen(cmd, args)
}
