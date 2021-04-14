package api_websockets

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/config"
	"pandora-pay/mempool"
	api_store "pandora-pay/network/api/api-store"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/settings"
)

type APIWebsockets struct {
	GetMap   map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error)
	chain    *blockchain.Blockchain
	mempool  *mempool.Mempool
	ApiStore *api_store.APIStore
}

func (api *APIWebsockets) ValidateHandshake(handshake *APIHandshake) error {
	handshake2 := *handshake
	if handshake2[2] != string(config.NETWORK_SELECTED) {
		return errors.New("Network is different")
	}
	return nil
}

func (api *APIWebsockets) getHandshake(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	handshake := APIHandshake{}
	if err := json.Unmarshal(values, &handshake); err != nil {
		return nil, err
	}
	if err := api.ValidateHandshake(&handshake); err != nil {
		return nil, err
	}
	return json.Marshal(&APIHandshake{config.NAME, config.VERSION, string(config.NETWORK_SELECTED)})
}

func (api *APIWebsockets) getHash(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	blockHeight := APIBlockHeight(0)
	if err := json.Unmarshal(values, &blockHeight); err != nil {
		return nil, err
	}
	return api.ApiStore.LoadBlockHash(blockHeight)
}

func (api *APIWebsockets) getBlock(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	blockHeight := APIBlockHeight(0)
	var blk *api_store.BlockWithTxs
	var err error

	if err := json.Unmarshal(values, &blockHeight); err != nil {
		return nil, err
	}
	if blk, err = api.ApiStore.LoadBlockWithTXsFromHeight(blockHeight); err != nil {
		return nil, err
	}
	return json.Marshal(blk)
}

func (api *APIWebsockets) getBlockComplete(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	blockHeight := APIBlockHeight(0)
	var blkComplete *block_complete.BlockComplete
	var err error

	if err = json.Unmarshal(values, &blockHeight); err != nil {
		return nil, err
	}
	if blkComplete, err = api.ApiStore.LoadBlockCompleteFromHeight(blockHeight); err != nil {
		return nil, err
	}

	return blkComplete.Serialize(), nil
}

func CreateWebsocketsAPI(apiStore *api_store.APIStore, chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) *APIWebsockets {

	api := APIWebsockets{
		chain:    chain,
		mempool:  mempool,
		ApiStore: apiStore,
	}

	api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error){
		"handshake":      api.getHandshake,
		"hash":           api.getHash,
		"block":          api.getBlock,
		"block-complete": api.getBlockComplete,
	}

	return &api
}