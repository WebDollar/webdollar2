package mempool_sync

import (
	"pandora-pay/config"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
)

type MempoolSync struct {
	websockets *websocks.Websockets
}

func (self *MempoolSync) DownloadMempool(conn *connection.AdvancedConnection) (err error) {

	cb := self.websockets.ApiWebsockets.GetMap["mempool/new-tx-id"]

	index, page := 0, 0
	count := config.API_MEMPOOL_MAX_TRANSACTIONS

	var chainHash []byte

	//times is used to avoid infinite loops
	for {

		var data *api_common.APIMempoolReply
		if data, err = connection.SendJSONAwaitAnswer[api_common.APIMempoolReply](conn, []byte("mempool"), &api_common.APIMempoolRequest{chainHash, page, 0}, nil, 0); err != nil {
			return
		}

		if len(data.Hashes) == 0 || index >= data.Count {
			break
		}

		if chainHash == nil {
			chainHash = data.ChainHash
		}

		for _, tx := range data.Hashes {
			cb(conn, tx)
		}

		index += len(data.Hashes)
		page++

		if page > 20 || len(data.Hashes) != count { //done
			break
		}

	}

	return
}

func NewMempoolSync(websockets *websocks.Websockets) *MempoolSync {
	return &MempoolSync{
		websockets: websockets,
	}
}
