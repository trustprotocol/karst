package model

import "karst/merkletree"

type PutInfo struct {
	InputfilePath    string
	Md5              string
	MerkleTree       *merkletree.MerkleTreeNode
	MerkleTreeSealed *merkletree.MerkleTreeNode
	StoredPath       string
}
