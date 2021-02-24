package forging

import (
	"encoding/binary"
	"pandora-pay/block"
	"pandora-pay/block/difficulty"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/wallet"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

func createNextBlock(height uint64) (*block.Block, error) {

	if height == 0 {
		return genesis.CreateGenesisBlock()
	} else {

		var blockHeader = block.BlockHeader{
			Version: 0,
			Height:  blockchain.Chain.Height,
		}
		var blk = block.Block{
			BlockHeader:    blockHeader,
			MerkleHash:     crypto.SHA3Hash([]byte{}),
			PrevHash:       blockchain.Chain.Hash,
			PrevKernelHash: blockchain.Chain.KernelHash,
			Timestamp:      0,
		}

		return &blk, nil
	}

}

//inside a thread
func forge(blk *block.Block, threads, threadIndex int, wg *sync.WaitGroup) {

	buf := make([]byte, binary.MaxVarintLen64)

	serialized := blk.SerializeBlock(false, false, false, false, false)
	now := time.Now()
	timestamp := uint64(now.Unix())

	addresses := wallet.GetAddresses()

	for forging {

		if timestamp > uint64(now.Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		//forge with my wallets
		for i := 0; i < len(addresses) && forging; i++ {

			if i%threads == threadIndex {

				n := binary.PutUvarint(buf, timestamp)
				serialized = append(serialized, buf[:n]...)

				serialized = append(serialized, addresses[i].PublicKey...)
				kernelHash := crypto.SHA3Hash(serialized)

				if difficulty.CheckKernelHashBig(kernelHash, blockchain.Chain.BigDifficulty) {

					mutex.Lock()

					copy(blk.Forger[:], addresses[i].PublicKey[:])
					blk.Timestamp = timestamp
					serializationForSigning := blk.SerializeForSigning()
					signature, _ := addresses[i].PrivateKey.Sign(&serializationForSigning)

					copy(blk.Signature[:], signature[:])

					var array []*block.Block
					array = append(array, blk)

					result, err := blockchain.Chain.AddBlocks(array)
					if err == nil && result {
						forging = false
					}

					mutex.Unlock()

				} else {
					serialized = serialized[:len(serialized)-n-33]
				}

			}

		}
		timestamp += 1

	}

	wg.Done()
}