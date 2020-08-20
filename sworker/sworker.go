package sworker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"karst/config"
	"karst/logger"
	"karst/merkletree"
	"net/http"
	"time"
)

type sealedMessage struct {
	Status int
	Body   string
	Path   string
}

type unsealBackMessage struct {
	Status int
	Body   string
	Path   string
}

func Seal(sworker *config.SworkerConfiguration, path string, merkleTree *merkletree.MerkleTreeNode) (*merkletree.MerkleTreeNode, string, error) {
	// Generate request
	url := sworker.HttpBaseUrl + "/api/v0/storage/seal"
	reqBody := map[string]interface{}{
		"body": merkleTree,
		"path": path,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("backup", sworker.Backup)

	// Request
	client := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		returnBody, _ := ioutil.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("Request seal failed, error is: %s, error code is: %d", string(returnBody), resp.StatusCode)
	}

	returnBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var sealedMsg sealedMessage
	err = json.Unmarshal([]byte(returnBody), &sealedMsg)
	if err != nil {
		return nil, "", fmt.Errorf("Unmarshal seal result failed: %s", err)
	}

	if sealedMsg.Status != 200 {
		return nil, "", fmt.Errorf("Seal failed, error code is %d", sealedMsg.Status)
	}

	var merkleTreeSealed merkletree.MerkleTreeNode
	if err = json.Unmarshal([]byte(sealedMsg.Body), &merkleTreeSealed); err != nil {
		return nil, "", fmt.Errorf("Unmarshal sealed merkle tree failed: %s", err)
	}

	return &merkleTreeSealed, sealedMsg.Path, nil
}

func Unseal(sworker *config.SworkerConfiguration, path string) (*merkletree.MerkleTreeNode, string, error) {
	// Generate request
	url := sworker.HttpBaseUrl + "/api/v0/storage/unseal"
	reqBody := map[string]interface{}{
		"path": path,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("backup", sworker.Backup)

	// Request
	client := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		returnBody, _ := ioutil.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("Request unseal failed, error is: %s, error code is: %d", string(returnBody), resp.StatusCode)
	}

	returnBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var unsealBackMes unsealBackMessage
	err = json.Unmarshal([]byte(returnBody), &unsealBackMes)
	if err != nil {
		return nil, "", fmt.Errorf("Unmarshal unseal back message failed: %s", err)
	}
	if unsealBackMes.Status != 200 {
		return nil, "", fmt.Errorf("Unseal failed: %s", unsealBackMes.Body)
	}

	return nil, unsealBackMes.Path, nil
}

func Confirm(sworker *config.SworkerConfiguration, sealedHash string) error {
	// Generate request
	url := sworker.HttpBaseUrl + "/api/v0/storage/confirm"
	reqBody := map[string]interface{}{
		"hash": sealedHash,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("backup", sworker.Backup)

	// Request
	client := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		returnBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Request confirm failed, error is: %s, error code is: %d", string(returnBody), resp.StatusCode)
	}

	returnBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.Debug(string(returnBody))
	return nil
}

func Delete(sworker *config.SworkerConfiguration, sealedHash string) error {
	// Generate request
	url := sworker.HttpBaseUrl + "/api/v0/storage/delete"
	reqBody := map[string]interface{}{
		"hash": sealedHash,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("backup", sworker.Backup)

	// Request
	client := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		returnBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Request delete failed, error is: %s, error code is: %d", string(returnBody), resp.StatusCode)
	}

	returnBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.Debug(string(returnBody))
	return nil
}
