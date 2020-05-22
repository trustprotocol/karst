package chain

import (
	"encoding/json"
	"errors"
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

type SOrderRequest struct {
	SOrder string `json:"sorder"`
	Backup string `json:"backup"`
}

type StorageOrder struct {
	Provider       string `json:"provider"`
	Amount         uint64 `json:"amount"`
	FileIdentifier string `json:"fileIdentifier"`
	FileSize       uint64 `json:"fileSize"`
	Duration       uint64 `json:"duration"`
}

type FullStorageOrder struct {
	Provider       string `json:"provider"`
	Client         string `json:"client"`
	Amount         uint64 `json:"amount"`
	FileIdentifier string `json:"file_identifier"`
	FileSize       uint64 `json:"file_size"`
	Duration       uint64 `json:"duration"`
	CreatedOn      uint64 `json:"created_on"`
	ExpiredOn      uint64 `json:"expired_on"`
	OrderStatus    string `json:"order_status"`
}

type SOrderResponse struct {
	OrderId string `json:"orderId"`
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

func PlaceStorageOrder(baseUrl string, backup string, pwd string, provider string, fId string, fSize uint64) (string, error) {
	header := req.Header{
		"password": pwd,
	}

	sOrder := StorageOrder{
		Provider:       provider,
		Amount:         0,
		FileIdentifier: fId,
		FileSize:       fSize,
		Duration:       320,
	}

	sOrderStr, err := json.Marshal(sOrder)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}

	sOrderReq := SOrderRequest{
		SOrder: string(sOrderStr),
		Backup: backup,
	}

	body := req.BodyJSON(&sOrderReq)

	r, err := req.Post(baseUrl+"/api/v1/market/sorder", header, body)
	if err != nil {
		return "", err
	}

	if r.Response().StatusCode != 200 {
		return "", fmt.Errorf("Place storage order failed, error code: %d", r.Response().StatusCode)
	}
	logger.Debug("Response from sorder: %s", r)

	sOrderRes := SOrderResponse{}
	if err = r.ToJSON(&sOrderRes); err != nil {
		return "", err
	}
	return sOrderRes.OrderId, nil
}

func GetStorageOrder(baseUrl string, orderId string) (FullStorageOrder, error) {
	param := req.Param{
		"orderId": orderId,
	}
	r, err := req.Get(baseUrl+"/api/v1/market/sorder", param)
	sOrder := FullStorageOrder{}

	if err != nil {
		return sOrder, err
	}

	if r.Response().StatusCode == 200 {
		err := r.ToJSON(&sOrder)
		if err != nil {
			return sOrder, err
		}
		return sOrder, nil
	}

	return sOrder, errors.New("Error from crust api")
}
