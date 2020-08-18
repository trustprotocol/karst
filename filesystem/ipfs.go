package filesystem

import (
	"karst/config"

	shell "github.com/ipfs/go-ipfs-api"
)

type Ipfs struct {
	sh *shell.Shell
}

func OpenIpfs(cfg *config.Configuration) (*Ipfs, error) {
	return &Ipfs{
		sh: shell.NewShell(cfg.Fs.Ipfs.BaseUrl),
	}, nil
}
