# go-pandora-pay
PandoraPay blockchain in go

The main design pattern that has been taken in consideration is to be **dead-simple**. A source code that is simple is bug free and easy to be developed and improved over time.

Status of Blockchain implementation:

- [x] Simple GUI
- [x] CLI commands
- [x] ECDSA
    - [x] Private Key
    - [x] Public Address (amount and paymentId)
    - [x] HD Wallet
- [x] Commit/Rollback Database
    - [x] Wallet
        - [x] Save and Load
        - [x] Export and import JSON        
    - [ ] Wallet Encryption
- [x] Merkle Tree
- [x] Block
    - [x] Serialization
    - [x] Deserialization
    - [x] Hashing
    - [x] Kernel Hash
    - [x] Forger signing  
- [x] Blockchain
    - [x] Saving state
    - [x] Locking mechanism
    - [x] Difficulty Adjustment
    - [x] Timestamp maximum drift    
- [x] Forging
    - [x] Forging with wallets Multithreading    
    - [ ] Forging with delegated stakes
        - [ ] Accepting to delegate stakes from network  
- [ ] Balance
    - [ ] Balance Update
    - [ ] Tokens
    - [ ] Patricia Trie ?
- [ ] Transactions
    - [ ] Transparent Transactions
    - [ ] Transaction Builder
    - [ ] Zether Deposit Transactions
    - [ ] Zether Withdraw Transactions
    - [ ] Zether Transfer Transactions
    - [ ] Multi Threading signature verification
- [ ] Mem Pool
    - [ ] Saving/Loading
    - [ ] Inserting Txs
    - [ ] Network propagation
- [ ] Network
    - [ ] HTTP server
    - [ ] HTTP websocket
    - [ ] TOR Integration