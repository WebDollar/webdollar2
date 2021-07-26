package mempool

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"sync/atomic"
)

type mempoolWork struct {
	chainHash    []byte         `json:"-"` //32 byte
	chainHeight  uint64         `json:"-"`
	result       *MempoolResult `json:"-"`
	waitAnswerCn chan struct{}
}

type mempoolWorker struct {
	dbTx store_db_interface.StoreDBTransactionInterface `json:"-"`
}

type MempoolWorkerAddTx struct {
	Tx     *mempoolTx
	Result chan<- error
}

type MempoolWorkerRemoveTxs struct {
	Txs    []*transaction.Transaction
	Result chan<- bool
}

//process the worker for transactions to prepare the transactions to the forger
func (worker *mempoolWorker) processing(
	newWorkCn <-chan *mempoolWork,
	suspendProcessingCn <-chan struct{},
	continueProcessingCn <-chan bool,
	addTransactionCn <-chan *MempoolWorkerAddTx,
	removeTransactionsCn <-chan *MempoolWorkerRemoveTxs,
	txs *MempoolTxs,
) {

	var work *mempoolWork

	txsList := []*mempoolTx{}
	txsMap := make(map[string]bool)
	listIndex := 0
	readyListSent := true

	var accs *accounts.Accounts
	var toks *tokens.Tokens

	includedTotalSize := uint64(0)
	includedTxs := []*mempoolTx{}
	sendReadyListCn := make(chan struct{})

	txs.clearList()

	resetNow := func(newWork *mempoolWork) {

		if newWork.chainHash != nil {
			txs.clearList()
			if readyListSent {
				close(sendReadyListCn)
				readyListSent = false
			}
		}
		close(newWork.waitAnswerCn)

		if newWork.chainHash != nil {
			accs = nil
			toks = nil
			work = newWork
			includedTotalSize = uint64(0)
			includedTxs = []*mempoolTx{}
			listIndex = 0
			txsMap = make(map[string]bool)
			if len(txsList) > 1 {
				sortTxs(txsList)
			}
		}
	}

	removeTxs := func(data *MempoolWorkerRemoveTxs) {
		result := false

		removedTxsMap := make(map[string]bool)
		for _, tx := range data.Txs {
			if txsMap[tx.Bloom.HashStr] {
				removedTxsMap[tx.Bloom.HashStr] = true
				delete(txsMap, tx.Bloom.HashStr)
				result = true
			}
		}
		if len(removedTxsMap) > 0 {
			newList := make([]*mempoolTx, len(txsList)-len(removedTxsMap))
			c := 0
			for i, tx := range txsList {
				if !removedTxsMap[tx.Tx.Bloom.HashStr] {
					newList[c] = tx
					c += 1
				} else {
					if listIndex > i {
						listIndex -= 1
					}
				}
			}
		}

		data.Result <- result
	}

	suspended := false
	for {

		select {
		case <-suspendProcessingCn:
			suspended = true
			continue
		case newWork := <-newWorkCn:
			resetNow(newWork)
		case data := <-removeTransactionsCn:
			removeTxs(data)
		case noError := <-continueProcessingCn:
			suspended = false
			if noError {
				work = nil //it needs a new work
			} else {
				accs = nil
				toks = nil
			}
		}

		if work == nil || suspended { //if no work was sent, just loop again
			continue
		}

		//let's check hf the work has been changed
		store.StoreBlockchain.DB.View(func(dbTx store_db_interface.StoreDBTransactionInterface) (err error) {

			if accs != nil {
				accs.Tx = dbTx
				toks.Tx = dbTx
			}

			for {

				if accs == nil {
					accs = accounts.NewAccounts(dbTx)
					toks = tokens.NewTokens(dbTx)
				}

				select {
				case <-suspendProcessingCn:
					suspended = true
					return
				case newWork := <-newWorkCn:
					resetNow(newWork)
				case data := <-removeTransactionsCn:
					removeTxs(data)
				default:

					var tx *mempoolTx
					var newAddTx *MempoolWorkerAddTx

					if listIndex == len(txsList) {

						select {
						case newWork := <-newWorkCn:
							resetNow(newWork)
							continue
						case <-suspendProcessingCn:
							suspended = true
							return
						case data := <-removeTransactionsCn:
							removeTxs(data)
						case newAddTx = <-addTransactionCn:
							tx = newAddTx.Tx
						case <-sendReadyListCn:
							//sending readyList only in case there is no transaction in the add channel
							sendReadyListCn = make(chan struct{})
							txs.readyList()
							readyListSent = true
						}

					} else {
						tx = txsList[listIndex]
						listIndex += 1
					}

					var finalErr error

					if tx != nil && !txsMap[tx.Tx.Bloom.HashStr] {

						txsMap[tx.Tx.Bloom.HashStr] = true

						if err = tx.Tx.IncludeTransaction(work.chainHeight, accs, toks); err != nil {

							finalErr = err

							accs.Rollback()
							toks.Rollback()

							if newAddTx == nil {
								//removing
								//this is done because listIndex was incremented already before
								txsList = append(txsList[:listIndex-1], txsList[listIndex:]...)
								listIndex--
								delete(txsMap, tx.Tx.Bloom.HashStr)
							}

							txs.txs.Delete(tx.Tx.Bloom.HashStr)

						} else {

							if includedTotalSize+tx.Tx.Bloom.Size < config.BLOCK_MAX_SIZE {

								includedTotalSize += tx.Tx.Bloom.Size
								includedTxs = append(includedTxs, tx)

								atomic.StoreUint64(&work.result.totalSize, includedTotalSize)
								work.result.txs.Store(includedTxs)

								accs.CommitChanges()
								toks.CommitChanges()
							}

							if newAddTx != nil {
								txsList = append(txsList, newAddTx.Tx)
								listIndex += 1
							}

							txs.addToList(tx)

						}

					}

					if newAddTx != nil && newAddTx.Result != nil {
						newAddTx.Result <- finalErr
					}

				}
			}

		})

	}
}
