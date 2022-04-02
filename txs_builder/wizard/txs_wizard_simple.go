package wizard

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

func signSimpleTransaction(tx *transaction.Transaction, privateKey *addresses.PrivateKey, fee *WizardTransactionFee, statusCallback func(string)) (err error) {

	txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

	extraBytes := cryptography.SignatureSize
	txBase.Fee = setFee(tx, extraBytes, fee.Clone(), true)
	statusCallback("Transaction Fee set")

	statusCallback("Transaction Signing...")
	if txBase.Vin.Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return err
	}
	statusCallback("Transaction Signed")

	return
}

func CreateSimpleTx(nonce uint64, key []byte, chainHeight uint64, extra WizardTxSimpleExtra, data *WizardTransactionData, fee *WizardTransactionFee, feeVersion bool, validateTx bool, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	privateKey, err := addresses.NewPrivateKey(key)
	if err != nil {
		return nil, err
	}

	dataFinal, err := data.getData()
	if err != nil {
		return
	}

	spaceExtra := 0

	var txScript transaction_simple.ScriptType
	var extraFinal transaction_simple_extra.TransactionSimpleExtraInterface
	switch txExtra := extra.(type) {
	case *WizardTxSimpleExtraUpdateAssetFeeLiquidity:
		extraFinal = &transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity{
			Liquidities:     txExtra.Liquidities,
			CollectorHasNew: txExtra.CollectorHasNew,
			Collector:       txExtra.Collector,
		}
		txScript = transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY

		spaceExtra += 1 + len(txExtra.Collector) + 1
		for _, liquidity := range txExtra.Liquidities {
			if liquidity.Rate > 0 {
				spaceExtra += len(helpers.SerializeToBytes(liquidity))
			}
		}
	}

	txBase := &transaction_simple.TransactionSimple{
		TxScript:    txScript,
		DataVersion: data.getDataVersion(),
		Data:        dataFinal,
		Nonce:       nonce,
		Fee:         0,
		Extra:       extraFinal,
		Vin: &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: privateKey.GeneratePublicKey(),
		},
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_SIMPLE,
		SpaceExtra:               uint64(spaceExtra),
		TransactionBaseInterface: txBase,
	}
	statusCallback("Transaction Created")

	if err = signSimpleTransaction(tx, privateKey, fee, statusCallback); err != nil {
		return
	}
	if err = bloomAllTx(tx, statusCallback); err != nil {
		return
	}
	return tx, nil
}
