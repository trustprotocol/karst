package chain

import (
	"encoding/json"
	"errors"
	"fmt"
	"karst/config"
	"karst/logger"

	"github.com/imroc/req"
)

type registerRequest struct {
	AddressInfo  string `json:"addressInfo"`
	StoragePrice uint64 `json:"storagePrice"`
	Backup       string `json:"backup"`
}

type merchant struct {
	Address string              `json:"address"`
	FileMap map[string][]string `json:"file_map"`
}

type sOrderRequest struct {
	SOrder string `json:"sorder"`
	Backup string `json:"backup"`
}

type storageOrder struct {
	Merchant       string `json:"merchant"`
	FileIdentifier string `json:"fileIdentifier"`
	FileSize       uint64 `json:"fileSize"`
	Duration       uint64 `json:"duration"`
}

type fullStorageOrder struct {
	Merchant       string `json:"merchant"`
	Client         string `json:"client"`
	FileIdentifier string `json:"file_identifier"`
	FileSize       uint64 `json:"file_size"`
	Duration       uint64 `json:"duration"`
	CreatedOn      uint64 `json:"created_on"`
	ExpiredOn      uint64 `json:"expired_on"`
	OrderStatus    string `json:"order_status"`
}

type sOrderResponse struct {
	OrderId string `json:"order_id"`
}

type systemHealth struct {
	Peers           uint64 `json:"peers"`
	IsSyncing       bool   `json:"isSyncing"`
	ShouldHavePeers bool   `json:"shouldHavePeers"`
}

func Register(cfg *config.Configuration, karstAddr string, storagePrice uint64) error {
	header := req.Header{
		"password": cfg.Crust.Password,
	}

	regReq := registerRequest{
		AddressInfo:  karstAddr,
		StoragePrice: storagePrice,
		Backup:       cfg.Crust.Backup,
	}

	body := req.BodyJSON(&regReq)
	logger.Debug("Register request body: %s", body)

	r, err := req.Post("http://"+cfg.Crust.BaseUrl+"/api/v1/market/register", header, body)

	if err != nil {
		return err
	}

	if r.Response().StatusCode != 200 {
		return fmt.Errorf("Register karst merchant failed! Error code is: %d", r.Response().StatusCode)
	}

	logger.Debug("Register response: %s", r)

	return nil
}

func GetMerchantAddr(cfg *config.Configuration, pChainAddr string) (string, error) {
	param := req.Param{
		"address": pChainAddr,
	}
	r, err := req.Get("http://"+cfg.Crust.BaseUrl+"/api/v1/market/merchant", param)

	if err != nil {
		return "", err
	}

	if r.Response().StatusCode != 200 {
		return "", fmt.Errorf("Get merchant failed! Error code is: %d", r.Response().StatusCode)
	}
	logger.Debug("Get merchant response: %s", r)

	merchant := merchant{}
	if err = r.ToJSON(&merchant); err != nil {
		return "", err
	}

	return merchant.Address, nil
}

func GetMerchantFileMap(cfg *config.Configuration, pChainAddr string) (map[string][]string, error) {
	param := req.Param{
		"address": pChainAddr,
	}
	r, err := req.Get("http://"+cfg.Crust.BaseUrl+"/api/v1/market/merchant", param)

	if err != nil {
		return nil, err
	}

	if r.Response().StatusCode != 200 {
		return nil, fmt.Errorf("Get merchant failed! Error code is: %d", r.Response().StatusCode)
	}
	logger.Debug("Get merchant response: %s", r)

	merchant := merchant{}
	if err = r.ToJSON(&merchant); err != nil {
		return nil, err
	}

	return merchant.FileMap, nil
}

func PlaceStorageOrder(cfg *config.Configuration, merchant string, duration uint64, fId string, fSize uint64) (string, error) {
	header := req.Header{
		"password": cfg.Crust.Password,
	}

	sOrder := storageOrder{
		Merchant:       merchant,
		FileIdentifier: fId,
		FileSize:       fSize,
		Duration:       duration,
	}

	sOrderStr, err := json.Marshal(sOrder)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}

	sOrderReq := sOrderRequest{
		SOrder: string(sOrderStr),
		Backup: cfg.Crust.Backup,
	}

	body := req.BodyJSON(&sOrderReq)

	r, err := req.Post("http://"+cfg.Crust.BaseUrl+"/api/v1/market/sorder", header, body)
	if err != nil {
		return "", err
	}

	if r.Response().StatusCode != 200 {
		return "", fmt.Errorf("Place storage order failed, error code: %d", r.Response().StatusCode)
	}
	logger.Debug("Response from sorder: %s", r)

	sOrderRes := sOrderResponse{}
	if err = r.ToJSON(&sOrderRes); err != nil {
		return "", err
	}
	return sOrderRes.OrderId, nil
}

func GetStorageOrder(cfg *config.Configuration, orderId string) (fullStorageOrder, error) {
	param := req.Param{
		"orderId": orderId,
	}
	r, err := req.Get("http://"+cfg.Crust.BaseUrl+"/api/v1/market/sorder", param)
	sOrder := fullStorageOrder{}

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

func IsReady(cfg *config.Configuration) bool {
	r, err := req.Get("http://" + cfg.Crust.BaseUrl + "/api/v1/system/health")
	if err != nil {
		return false
	}

	if r.Response().StatusCode != 200 {
		return false
	}

	sh := systemHealth{
		Peers:           0,
		IsSyncing:       true,
		ShouldHavePeers: false,
	}

	err = r.ToJSON(&sh)
	if err != nil {
		logger.Error("%s", err)
		return false
	}

	return !sh.IsSyncing
}
