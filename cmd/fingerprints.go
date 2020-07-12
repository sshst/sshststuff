package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sshst/sshststuff"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net/http"
	"os"
)

func addGithubSshKeys(listenConfig *client.ListenConfig) {
	for _, username := range listenConfig.GithubUsers {
		fingerprints, err := githubUserFingerprints(username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting SSH key for %s: %s\n", username, err.Error())
			continue
		}

		for _, fingerprint := range fingerprints {
			listenConfig.SSHFingerprints = append(listenConfig.SSHFingerprints, fingerprint)
			fmt.Fprintf(os.Stderr, "Trusting fingerprint %s for GitHub user %s\n", fingerprint, username)
		}
	}
}

func githubUserFingerprints(username string) ([]string, error) {
	url := fmt.Sprintf("https://github.com/%s.keys", username)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	authorizedKeysBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fingerprints := []string{}

	lines := bytes.Split(authorizedKeysBytes, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		pubkey, _, _, _, err := ssh.ParseAuthorizedKey(line)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		fingerprint := ssh.FingerprintSHA256(pubkey)
		fingerprints = append(fingerprints, fingerprint)
	}

	if len(fingerprints) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: User '%s' provided but they have no SSH keys defined at %s\n", username, url)
	}

	return fingerprints, nil
}
