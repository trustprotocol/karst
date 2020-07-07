package tee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"karst/config"
	"karst/logger"
	"karst/merkletree"
	"net/http"

	"github.com/gorilla/websocket"
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

// TODO: change to wss
func Seal(tee *config.TeeConfiguration, path string, merkleTree *merkletree.MerkleTreeNode) (*merkletree.MerkleTreeNode, string, error) {
	// Connect to tee
	url := tee.WsBaseUrl + "/storage/seal"
	logger.Info("Connecting to TEE '%s' to seal file", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, "", err
	}
	defer c.Close()

	// Send file to seal
	reqBody := map[string]interface{}{
		"backup": tee.Backup,
		"body":   merkleTree,
		"path":   path,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", err
	} else {
		logger.Debug("Request body for sealing: %s", string(reqBodyBytes))
	}

	err = c.WriteMessage(websocket.TextMessage, reqBodyBytes)
	if err != nil {
		return nil, "", err
	}

	// Deal result
	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, "", err
	}
	logger.Debug("Recv: %s", message)

	var sealedMsg sealedMessage
	err = json.Unmarshal([]byte(message), &sealedMsg)
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

func Unseal(tee *config.TeeConfiguration, path string) (*merkletree.MerkleTreeNode, string, error) {
	// Connect to tee
	url := tee.WsBaseUrl + "/storage/unseal"
	logger.Info("Connecting to TEE '%s' to unseal file", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, "", err
	}
	defer c.Close()

	// Send file to unseal
	reqBody := map[string]interface{}{
		"backup": tee.Backup,
		"path":   path,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", err
	} else {
		logger.Debug("Request body for unsealing: %s", string(reqBodyBytes))
	}

	err = c.WriteMessage(websocket.TextMessage, reqBodyBytes)
	if err != nil {
		return nil, "", err
	}

	// Deal result
	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, "", err
	}
	logger.Debug("Recv: %s", message)

	var unsealBackMes unsealBackMessage
	err = json.Unmarshal([]byte(message), &unsealBackMes)
	if err != nil {
		return nil, "", fmt.Errorf("Unmarshal unseal back message failed: %s", err)
	}
	if unsealBackMes.Status != 200 {
		return nil, "", fmt.Errorf("Unseal failed: %s", unsealBackMes.Body)
	}

	return nil, unsealBackMes.Path, nil
}

func Confirm(tee *config.TeeConfiguration, sealedHash string) error {
	// Generate request
	url := tee.HttpBaseUrl + "/storage/confirm"
	reqBody := map[string]interface{}{
		"hash": sealedHash,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("backup", tee.Backup)

	// Request
	client := &http.Client{}
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
