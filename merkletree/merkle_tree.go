package merkletree

import "crypto/sha256"

type MerkleTreeNode struct {
	Hash  [32]byte
	Size  uint64
	Links []*MerkleTreeNode
}

func NewMerkleTreeNode(hash [32]byte, size uint64) *MerkleTreeNode {

	return &MerkleTreeNode{
		Hash:  hash,
		Size:  size,
		Links: nil,
	}
}

// TODO: Multiple depth tree, currently only supports single-layer tree
func CreateMerkleTree(hashs [][32]byte, sizes []uint64) *MerkleTreeNode {
	allHashs := make([]byte, 0)
	var totalSize uint64 = 0
	links := make([]*MerkleTreeNode, 0)

	for index := range hashs {
		links = append(links, NewMerkleTreeNode(hashs[index], sizes[index]))
		totalSize = totalSize + sizes[index]
		allHashs = append(allHashs, hashs[index][:]...)
	}

	return &MerkleTreeNode{
		Hash:  sha256.Sum256(allHashs[:]),
		Size:  totalSize,
		Links: links,
	}
}
