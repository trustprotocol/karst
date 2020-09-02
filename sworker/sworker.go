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
	Body string
	Path string
}

func httpRetryHandle(client *http.Client, req *http.Request, cfg *config.Configuration) ([]byte, error) {
	tryTimes := 0

	for {
		tryTimes++
		resp, err := client.Do(req)
		if err != nil {
			if tryTimes > cfg.RetryTimes {
				return nil, err
			}
		} else {
			if resp.StatusCode == 200 {
				returnBody, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				return returnBody, err
			} else if resp.StatusCode == 503 {
				resp.Body.Close()
				logger.Debug("SWorker is updating, will still wait for %d s", cfg.RetryInterval.Milliseconds()*(int64)(cfg.RetryTimes*180-tryTimes))
				if tryTimes > cfg.RetryTimes*180 {
					return nil, fmt.Errorf("SWorker updates too slow")
				}
			} else {
				resp.Body.Close()
				return nil, fmt.Errorf("Error code is: %d", resp.StatusCode)
			}
		}
		time.Sleep(cfg.RetryInterval)
	}
}

func Seal(cfg *config.Configuration, path string, merkleTree *merkletree.MerkleTreeNode) (*merkletree.MerkleTreeNode, string, error) {
	// Generate request
	url := cfg.Sworker.HttpBaseUrl + "/api/v0/storage/seal"
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
	req.Header.Set("backup", cfg.Sworker.Backup)

	// Request
	client := &http.Client{
		Timeout: 1000 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	returnBody, err := httpRetryHandle(client, req, cfg)
	if err != nil {
		return nil, "", err
	}

	var sealedMsg sealedMessage
	err = json.Unmarshal([]byte(returnBody), &sealedMsg)
	if err != nil {
		return nil, "", fmt.Errorf("Unmarshal seal result failed: %s", err)
	}

	var merkleTreeSealed merkletree.MerkleTreeNode
	if err = json.Unmarshal([]byte(sealedMsg.Body), &merkleTreeSealed); err != nil {
		return nil, "", fmt.Errorf("Unmarshal sealed merkle tree failed: %s", err)
	}

	return &merkleTreeSealed, sealedMsg.Path, nil
}

func Unseal(cfg *config.Configuration, path string) (string, error) {
	// Generate request
	url := cfg.Sworker.HttpBaseUrl + "/api/v0/storage/unseal"
	reqBody := map[string]interface{}{
		"path": path,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("backup", cfg.Sworker.Backup)

	// Request
	client := &http.Client{
		Timeout: 1000 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	returnBody, err := httpRetryHandle(client, req, cfg)
	if err != nil {
		return "", err
	}

	return string(returnBody), nil
}

func Confirm(cfg *config.Configuration, sealedHash string) error {
	// Generate request
	url := cfg.Sworker.HttpBaseUrl + "/api/v0/storage/confirm"
	reqBody := map[string]interface{}{
		"hash": sealedHash,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("backup", cfg.Sworker.Backup)

	// Request
	client := &http.Client{
		Timeout: 1000 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	returnBody, err := httpRetryHandle(client, req, cfg)
	if err != nil {
		return err
	}

	logger.Debug(string(returnBody))
	return nil
}

func Delete(cfg *config.Configuration, sealedHash string) error {
	// Generate request
	url := cfg.Sworker.HttpBaseUrl + "/api/v0/storage/delete"
	reqBody := map[string]interface{}{
		"hash": sealedHash,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("backup", cfg.Sworker.Backup)

	// Request
	client := &http.Client{
		Timeout: 1000 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	returnBody, err := httpRetryHandle(client, req, cfg)
	if err != nil {
		return err
	}

	logger.Debug(string(returnBody))
	return nil
}
