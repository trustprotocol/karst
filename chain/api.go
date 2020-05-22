package chain

import (
	"github.com/imroc/req"
	"karst/logger"
)

type RegisterRequest struct {
	AddressInfo string `json:"addressInfo"`
	Backup      string `json:"backup"`
}

func Register(baseUrl string, backup string, pwd string, karstAddr string) bool {
	header := req.Header{
		"password": pwd,
	}

	regReq := RegisterRequest{
		AddressInfo: karstAddr,
		Backup:      backup,
	}

	body := req.BodyJSON(&regReq)
	logger.Debug("Register request body:", body)

	r, err := req.Post(baseUrl+"/api/v1/market/register", header, body)
	logger.Debug("Register response:", r)

	rst := err == nil && r.Response().StatusCode == 200

	if !rst {
		logger.Error(err.Error())
	}

	return rst
}
