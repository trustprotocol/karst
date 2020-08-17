package filesystem

import (
	"karst/config"

	shell "github.com/ipfs/go-ipfs-api"
)

type Ipfs struct {
}

func OpenIpfs(cfg *config.Configuration) (*Ipfs, error) {
	shell.NewShell("localhost:5001")
	return &Ipfs{}, nil
}
