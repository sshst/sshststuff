package client

type ListenConfig struct {
	Version         string   `json:"-"`
	APIURL          string   `json:"-"`
	IdleTimeout     int      `json:"-"`
	MaxTimeout      int      `json:"-"`
	HostSeed        string   `json:"-"`
	Command         []string `json:"-"`
	SSHFingerprints []string `json:"-"`
	WebOk           bool
	NotifyUser      string
	NotifyTitle     string
	GithubUsers     []string
}
