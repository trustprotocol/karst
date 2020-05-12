package model

import "karst/merkletree"

type StorePermissionMessage struct {
	ChainAccount   string                     `json:"chain_account"`
	StoreOrderHash string                     `json:"store_order"`
	MekleTree      *merkletree.MerkleTreeNode `json:"merkle_tree"`
}

type BackupMessage struct {
	Backup string `json:"back_up"`
}

type NodeDataMessage struct {
	FileHash  string `json:"file_hash"`
	NodeHash  string `json:"node_hash"`
	NodeIndex uint64 `json:"node_index"`
}
