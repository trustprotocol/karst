package fs

type FsInterface interface {
	Close()
	Put(fileName string) (string, error)
	Get(key string, outFileName string) error
	Delete(key string) error
}
