package chain

import (
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
func Register(baseUrl string, backup string, pwd string, karstAddr string) bool {
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
	logger.Debug("Register response: %s", r)

	rst := err == nil && r.Response().StatusCode == 200

	if !rst {
		logger.Error(err.Error())
	}

	return rst
}

func GetProvideAddr(baseUrl string, pChainAddr string) (string, error) {
	param := req.Param{
		"address": pChainAddr,
	}
	r, err := req.Get(baseUrl+"/api/v1/market/provider", param)

	if r.Response().StatusCode == 200 {
		provider := Provider{}
		r.ToJSON(&provider)
		return provider.Address, nil
	}

	logger.Error(err.Error())
	return "", err
}
