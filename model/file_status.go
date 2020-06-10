package model

type FileStatus struct {
	Hash       string `json:"hash"`
	Size       uint64 `json:"size"`
	SealedHash string `json:"sealed_hash"`
	SealedSize uint64 `json:"sealed_size"`
}
