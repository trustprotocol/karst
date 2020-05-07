package merkletree

import (
	"crypto/sha256"
	"encoding/hex"
)

type MerkleTreeNode struct {
	Hash     string           `json:"hash"`
	Size     uint64           `json:"size"`
	LinksNum uint64           `json:"links_num"`
	Links    []MerkleTreeNode `json:"links"`
}

func NewMerkleTreeNode(hash [32]byte, size uint64) *MerkleTreeNode {

	return &MerkleTreeNode{
		Hash:     hex.EncodeToString(hash[:]),
		Size:     size,
		LinksNum: 0,
		Links:    make([]MerkleTreeNode, 0),
	}
}

// TODO: Multiple depth tree, currently only supports single-layer tree
func CreateMerkleTree(hashs [][32]byte, sizes []uint64) *MerkleTreeNode {
	allHashs := make([]byte, 0)
	var totalSize uint64 = 0
	var linksNum uint64 = 0
	links := make([]MerkleTreeNode, 0)

	for index := range hashs {
		links = append(links, *NewMerkleTreeNode(hashs[index], sizes[index]))
		totalSize = totalSize + sizes[index]
		linksNum = linksNum + 1
		allHashs = append(allHashs, hashs[index][:]...)
	}

	hashBytes := sha256.Sum256(allHashs[:])
	return &MerkleTreeNode{
		Hash:     hex.EncodeToString(hashBytes[:]),
		Size:     totalSize,
		LinksNum: linksNum,
		Links:    links,
	}
}
