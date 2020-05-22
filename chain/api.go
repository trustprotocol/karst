package chain

import (
	"fmt"
	"karst/logger"

	"github.com/imroc/req"
)

type RegisterRequest struct {
	AddressInfo string `json:"addressInfo"`
	Backup      string `json:"backup"`
}

type Provider struct {
	Address string `json:"address"`
	FileMap string `json:"file_map"`
}

// TODO: extract baseUrl, backup and pwd to common structure
func Register(baseUrl string, backup string, pwd string, karstAddr string) error {
	header := req.Header{
		"password": pwd,
	}

	regReq := RegisterRequest{
		AddressInfo: karstAddr,
		Backup:      backup,
	}

	body := req.BodyJSON(&regReq)
	logger.Debug("Register request body: %s", body)

	r, err := req.Post(baseUrl+"/api/v1/market/register", header, body)

	if err != nil {
		return err
	}

	if r.Response().StatusCode != 200 {
		return fmt.Errorf("Register karst provider failed! Error code is: %d", r.Response().StatusCode)
	}

	logger.Debug("Register response: %s", r)

	return nil
}

func GetProviderAddr(baseUrl string, pChainAddr string) (string, error) {
	param := req.Param{
		"address": pChainAddr,
	}
	r, err := req.Get(baseUrl+"/api/v1/market/provider", param)

	if err != nil {
		return "", err
	}

	if r.Response().StatusCode != 200 {
		return "", fmt.Errorf("Get provider failed! Error code is: %d", r.Response().StatusCode)
	}
	logger.Debug("Get provider address response: %s", r)

	provider := Provider{}
	if err = r.ToJSON(&provider); err != nil {
		return "", err
	}

	return provider.Address, nil
}
