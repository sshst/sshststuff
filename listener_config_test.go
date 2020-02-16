package client

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
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
