package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	dir, err := ioutil.TempDir("", "e2etest")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	binPath := filepath.Join(dir, "bin")

	cmd := exec.Command("go", "build", "-o", binPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	pkeyPath := filepath.Join(dir, "id_rsa")
	err = ioutil.WriteFile(pkeyPath, pkeyBytes, 0600)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(privateKey)
	require.NoError(t, err)

	pr, pw := io.Pipe()
	lmulti := io.MultiWriter(pw, os.Stderr)

	fingerprint := gossh.FingerprintSHA256(signer.PublicKey())
	lcmd := exec.CommandContext(ctx, binPath, "listen", "-f", fingerprint)
	lcmd.Stderr = lmulti
	lcmd.Stdout = lmulti

	go func(cmd *exec.Cmd) {
		err := cmd.Run()
		if err == ctx.Err() {
			return
		}
		require.NoError(t, err)
	}(lcmd)

	scan := bufio.NewScanner(pr)
	host := ""
	for scan.Scan() {
		split := strings.Split(scan.Text(), "Listener ID: ")
		if len(split) == 2 {
			host = split[1]
			go ioutil.ReadAll(pr) // drain the pipe
			break
		}
	}

	buf := &bytes.Buffer{}
	multi := io.MultiWriter(buf, os.Stderr)

	cmd = exec.CommandContext(ctx, "ssh",  "-i", pkeyPath, "-o", fmt.Sprintf("ProxyCommand %s connect --sni %s", binPath, host), "test@"+host, "/bin/sh -c 'echo $SSHST'")
	cmd.Stdout = multi
	cmd.Stderr = multi
	err = cmd.Run()
	require.NoError(t, err)
	assert.Equal(t, "true\n", buf.String())

	cancel()
}
