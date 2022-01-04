package api_delegator_node

import (
	"bytes"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_nodes"
	"pandora-pay/recovery"
	"pandora-pay/wallet/wallet_address"
	"sync/atomic"
	"time"
)

func (api *DelegatorNode) execute() {
	recovery.SafeGo(func() {

		updateNewChainUpdateListener := api.chain.UpdateNewChain.AddListener()
		defer api.chain.UpdateNewChain.RemoveChannel(updateNewChainUpdateListener)

		for {

			chainHeight, ok := <-updateNewChainUpdateListener
			if !ok {
				return
			}

			atomic.StoreUint64(&api.chainHeight, chainHeight)
		}
	})

	recovery.SafeGo(func() {

		lastHeight := uint64(0)
		api.ticker = time.NewTicker(10 * time.Second)

		for {
			if _, ok := <-api.ticker.C; !ok {
				return
			}

			chainHeight := atomic.LoadUint64(&api.chainHeight)
			if lastHeight != chainHeight {
				lastHeight = chainHeight

				api.pendingDelegatesStakesChanges.Range(func(key, value interface{}) bool {
					pendingDelegateStakeChange := value.(*pendingDelegateStakeChange)
					if chainHeight >= pendingDelegateStakeChange.blockHeight+10 {
						api.pendingDelegatesStakesChanges.Delete(key)
					}
					return true
				})
			}
		}
	})

	api.updateAccountsChanges()

}

func (api *DelegatorNode) updateAccountsChanges() {

	recovery.SafeGo(func() {

		updatePlainAccountsCn := api.chain.UpdatePlainAccounts.AddListener()
		defer api.chain.UpdatePlainAccounts.RemoveChannel(updatePlainAccountsCn)

		for {

			plainAccs, ok := <-updatePlainAccountsCn
			if !ok {
				return
			}

			for k, v := range plainAccs.HashMap.Committed {
				data, loaded := api.pendingDelegatesStakesChanges.Load(k)
				if loaded {

					pendingDelegatingStakeChange := data.(*pendingDelegateStakeChange)

					if v.Stored == "update" {
						plainAcc := v.Element.(*plain_account.PlainAccount)
						if plainAcc.DelegatedStake.HasDelegatedStake() && bytes.Equal(plainAcc.DelegatedStake.DelegatedStakePublicKey, pendingDelegatingStakeChange.delegateStakingPublicKey) {

							if plainAcc.DelegatedStake.DelegatedStakeFee < config_nodes.DELEGATOR_FEE {
								continue
							}

							addr, err := addresses.CreateAddr(pendingDelegatingStakeChange.publicKey, nil, nil, 0, nil)
							if err != nil {
								continue
							}

							_ = api.wallet.AddDelegateStakeAddress(&wallet_address.WalletAddress{
								wallet_address.VERSION_NORMAL,
								"Delegate Stake",
								0,
								false,
								nil,
								nil,
								pendingDelegatingStakeChange.publicKey,
								make(map[string]*wallet_address.WalletAddressBalanceDecoded),
								addr.EncodeAddr(),
								"",
								&wallet_address.WalletAddressDelegatedStake{
									&addresses.PrivateKey{Key: pendingDelegatingStakeChange.delegateStakingPrivateKey.Key},
									pendingDelegatingStakeChange.delegateStakingPublicKey,
									0,
								},
							}, true)
						}
					}

				}
			}
		}
	})

}
