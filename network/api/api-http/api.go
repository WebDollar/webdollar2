package api_http

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common"
	"strconv"
)

type API struct {
	GetMap    map[string]func(values *url.Values) (interface{}, error)
	chain     *blockchain.Blockchain
	apiCommon *api_common.APICommon
	apiStore  *api_common.APIStore
}

func (api *API) getBlockchain(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetBlockchain()
}

func (api *API) getBlockchainSync(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetBlockchainSync()
}

func (api *API) getInfo(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetInfo()
}

func (api *API) getPing(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetPing()
}

func (api *API) getBlockComplete(values *url.Values) (out interface{}, err error) {

	request := &api_common.APIBlockCompleteRequest{}
	request.ReturnType = api_common.GetReturnType(values.Get("type"), api_common.RETURN_JSON)

	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	} else {
		err = errors.New("parameter 'hash' or 'height' are missing")
	}
	if err != nil {
		return
	}

	if out, err = api.apiCommon.GetBlockComplete(request); err != nil {
		return
	}

	if request.ReturnType == api_common.RETURN_SERIALIZED {
		out = helpers.HexBytes(out.([]byte))
	}

	return
}

func (api *API) getBlockHash(values *url.Values) (interface{}, error) {
	if values.Get("height") != "" {
		height, err := strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}

		out, err := api.apiCommon.GetBlockHash(height)
		if err != nil {
			return nil, err
		}
		return helpers.HexBytes(out.([]byte)), nil
	}
	return nil, errors.New("parameter `height` is missing")
}

func (api *API) getBlock(values *url.Values) (out interface{}, err error) {

	request := &api_common.APIBlockRequest{}

	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	} else {
		err = errors.New("parameter 'hash' or 'height' are missing")
	}
	if err != nil {
		return
	}

	return api.apiCommon.GetBlock(request)
}

func (api *API) getBlockInfo(values *url.Values) (out interface{}, err error) {

	request := &api_common.APIBlockRequest{}

	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	} else {
		err = errors.New("parameter 'hash' or 'height' are missing")
	}
	if err != nil {
		return
	}

	return api.apiCommon.GetBlockInfo(request)
}

func (api *API) getTx(values *url.Values) (out interface{}, err error) {

	request := &api_common.APITransactionRequest{}
	request.ReturnType = api_common.GetReturnType(values.Get("type"), api_common.RETURN_JSON)

	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	} else {
		err = errors.New("parameter 'hash' or 'height' are missing")
	}
	if err != nil {
		return
	}

	return api.apiCommon.GetTx(request)
}

func (api *API) getTxHash(values *url.Values) (interface{}, error) {
	if values.Get("height") != "" {
		height, err := strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}

		out, err := api.apiCommon.GetTxHash(height)
		if err != nil {
			return nil, err
		}
		return helpers.HexBytes(out.([]byte)), nil
	}
	return nil, errors.New("parameter `height` is missing")
}

func (api *API) getAccount(values *url.Values) (out interface{}, err error) {
	request := &api_common.APIAccountRequest{}
	request.ReturnType = api_common.GetReturnType(values.Get("type"), api_common.RETURN_JSON)

	if values.Get("address") != "" {
		request.Address = values.Get("address")
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	} else {
		err = errors.New("parameter 'address' or 'hash' was not specified")
	}
	if err != nil {
		return
	}
	return api.apiCommon.GetAccount(request)
}

func (api *API) getToken(values *url.Values) (interface{}, error) {
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		return nil, err
	}
	return api.apiCommon.GetToken(hash)
}

func (api *API) getMempool(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetMempool()
}

func (api *API) getMempoolExists(values *url.Values) (interface{}, error) {
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		return nil, err
	}
	return api.apiCommon.GetMempoolExists(hash)
}

func (api *API) postMempoolInsert(values *url.Values) (interface{}, error) {

	tx := &transaction.Transaction{}
	var err error

	if values.Get("type") == "json" {
		data := values.Get("tx")
		err = json.Unmarshal([]byte(data), tx)
	} else if values.Get("type") == "binary" {
		data, err := hex.DecodeString(values.Get("tx"))
		if err != nil {
			return nil, err
		}
		if err = tx.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
	} else {
		err = errors.New("parameter 'type' was not specified or is invalid")
	}
	if err != nil {
		return nil, err
	}

	return api.apiCommon.PostMempoolInsert(tx)
}

func CreateAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain) *API {

	api := API{
		chain:     chain,
		apiStore:  apiStore,
		apiCommon: apiCommon,
	}

	api.GetMap = map[string]func(values *url.Values) (interface{}, error){
		"":                   api.getInfo,
		"chain":              api.getBlockchain,
		"sync":               api.getBlockchainSync,
		"ping":               api.getPing,
		"block":              api.getBlock,
		"block-info":         api.getBlockInfo,
		"block-hash":         api.getBlockHash,
		"block-complete":     api.getBlockComplete,
		"tx":                 api.getTx,
		"tx-hash":            api.getTxHash,
		"account":            api.getAccount,
		"token":              api.getToken,
		"mempool":            api.getMempool,
		"mem-pool/tx-exists": api.getMempoolExists,
		"mem-pool/new-tx":    api.postMempoolInsert,
	}

	return &api
}
