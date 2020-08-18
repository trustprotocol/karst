package filesystem

import (
	"fmt"
	"io/ioutil"
	"karst/config"
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

type Ipfs struct {
	sh *shell.Shell
}

func OpenIpfs(cfg *config.Configuration) (*Ipfs, error) {
	sh := shell.NewShell(cfg.Fs.Ipfs.BaseUrl)
	if sh.IsUp() {
		return nil, fmt.Errorf("Ipfs dosen't start")
	}
	return &Ipfs{sh: sh}, nil
}

func (this *Ipfs) Close() {

}

func (this *Ipfs) Put(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return this.sh.Add(f)
}

func (this *Ipfs) Get(key string, outFileName string) error {
	return this.sh.Get(key, outFileName)
}

func (this *Ipfs) Delete(key string) error {
	return this.sh.Unpin(key)
}

func (this *Ipfs) GetToBuffer(key string, size uint64) ([]byte, error) {
	dataReader, err := this.sh.Cat(key)
	if err != nil {
		return nil, err
	}
	defer dataReader.Close()
	return ioutil.ReadAll(dataReader)
}
