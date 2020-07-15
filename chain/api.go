package chain

import (
	"encoding/json"
	"errors"
	"fmt"
	"karst/config"
	"karst/logger"

	"github.com/imroc/req"
)

type RegisterRequest struct {
	AddressInfo  string `json:"addressInfo"`
	StoragePrice uint64 `json:"storagePrice"`
	Backup       string `json:"backup"`
}

type Provider struct {
	Address string `json:"address"`
}

type SOrderRequest struct {
	SOrder string `json:"sorder"`
	Backup string `json:"backup"`
}

type StorageOrder struct {
	Provider       string `json:"provider"`
	FileIdentifier string `json:"fileIdentifier"`
	FileSize       uint64 `json:"fileSize"`
	Duration       uint64 `json:"duration"`
}

type FullStorageOrder struct {
	Provider       string `json:"provider"`
	Client         string `json:"client"`
	FileIdentifier string `json:"file_identifier"`
	FileSize       uint64 `json:"file_size"`
	Duration       uint64 `json:"duration"`
	CreatedOn      uint64 `json:"created_on"`
	ExpiredOn      uint64 `json:"expired_on"`
	OrderStatus    string `json:"order_status"`
}

type SOrderResponse struct {
	OrderId string `json:"order_id"`
}

func Register(cfg *config.Configuration, karstAddr string, storagePrice uint64) error {
	header := req.Header{
		"password": cfg.Crust.Password,
	}

	regReq := RegisterRequest{
		AddressInfo:  karstAddr,
		StoragePrice: storagePrice,
		Backup:       cfg.Crust.Backup,
	}

	body := req.BodyJSON(&regReq)
	logger.Debug("Register request body: %s", body)

	r, err := req.Post("http://"+cfg.Crust.BaseUrl+"/market/register", header, body)

	if err != nil {
		return err
	}

	if r.Response().StatusCode != 200 {
		return fmt.Errorf("Register karst provider failed! Error code is: %d", r.Response().StatusCode)
	}

	logger.Debug("Register response: %s", r)

	return nil
}

func GetProviderAddr(cfg *config.Configuration, pChainAddr string) (string, error) {
	param := req.Param{
		"address": pChainAddr,
	}
	r, err := req.Get("http://"+cfg.Crust.BaseUrl+"/market/provider", param)

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

func PlaceStorageOrder(cfg *config.Configuration, provider string, duration uint64, fId string, fSize uint64) (string, error) {
	header := req.Header{
		"password": cfg.Crust.Password,
	}

	sOrder := StorageOrder{
		Provider:       provider,
		FileIdentifier: fId,
		FileSize:       fSize,
		Duration:       duration,
	}

	sOrderStr, err := json.Marshal(sOrder)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}

	sOrderReq := SOrderRequest{
		SOrder: string(sOrderStr),
		Backup: cfg.Crust.Backup,
	}

	body := req.BodyJSON(&sOrderReq)

	r, err := req.Post("http://"+cfg.Crust.BaseUrl+"/market/sorder", header, body)
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

func GetStorageOrder(cfg *config.Configuration, orderId string) (FullStorageOrder, error) {
	param := req.Param{
		"orderId": orderId,
	}
	r, err := req.Get("http://"+cfg.Crust.BaseUrl+"/market/sorder", param)
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
