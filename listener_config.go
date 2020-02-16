package client

type ListenConfig struct {
	Version         string   `json:"-"`
	ApiUrl          string   `json:"-"`
	IdleTimeout     int      `json:"-"`
	MaxTimeout      int      `json:"-"`
	HostSeed        string   `json:"-"`
	Command         []string `json:"-"`
	SshFingerprints []string `json:"-"`
	WebOk           bool
	NotifyUser      string
	NotifyTitle     string
	GithubUsers     []string
}
