package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostSeedNotDisclosed(t *testing.T) {
	conf := ListenConfig{WebOk: true, HostSeed: "seed"}
	body, err := json.Marshal(conf)
	assert.NoError(t, err)
	assert.JSONEq(t, `
		{
		  "WebOk": true,
		  "NotifyUser": "",
		  "NotifyTitle": "",
		  "GithubUsers": null
		}
	`, string(body))
}
