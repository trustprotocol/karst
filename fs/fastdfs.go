package fs

import (
	"karst/config"
	"karst/fs/fastdfs"
)

type Fastdfs struct {
	client *fastdfs.Client
}

func OpenFastdfs(cfg *config.Configuration) (*Fastdfs, error) {
	client, err := fastdfs.NewClientWithConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Fastdfs{
		client: client,
	}, nil
}

func (this *Fastdfs) Close() {
	this.client.Destory()
}

func (this *Fastdfs) Put(fileName string) (string, error) {
	return this.client.UploadByFilename(fileName)
}

func (this *Fastdfs) Get(key string, outFileName string) error {
	return this.client.DownloadToFile(key, outFileName, 0, 0)
}

func (this *Fastdfs) Delete(key string) error {
	return this.client.DeleteFile(key)
}
