package merkletree

type MerkleTreeNode struct {
	Hash  []byte
	Size  uint64
	Links []*MerkleTreeNode
}

func NewMerkleTreeNode(contentFilePath string) *MerkleTreeNode {
	return nil
}
