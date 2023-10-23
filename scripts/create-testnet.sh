if [ $# -eq 0 ]; then
  echo "argument missing"
  echo "mode=normal for starting in normal mode"
  echo "mode=race to enable the race detection"
  echo "continue to just open the instances"
  echo "--pprof to enable debugging using profiling"
  echo "--debug to enable debug info"
  echo "--light-computations to make the testnet use less CPU"
  echo "--tcp-server-address=\"domain:port\""
  echo "--tcp-server-port=\"16000\""
  echo "--tcp-server-auto-tls-certificate"
  exit 1
fi

SCRIPTPATH="$(
  cd -- "$(dirname "$0")" >/dev/null 2>&1
  pwd -P
)"

nodes=4
race=false
continue=false
extraArgs=""

killall -9 go

for arg in $@; do
  if [ $arg == "--pprof" ]; then
    extraArgs+=" $arg "
  fi
  if [ $arg == "mode=race" ]; then
    race=true
  fi
  if [ $arg == "--debug" ]; then
    extraArgs+=" $arg "
  fi
  if [ $arg == "--light-computations" ]; then
    extraArgs+=" --light-computations"
  fi
  if [[ $arg == *"--tcp-server-address="* ]]; then
    extraArgs+=" $arg "
  fi
  if [[ $arg == *"--tcp-server-port="* ]]; then
    extraArgs+=" $arg "
  fi
  if [ $arg == "--tcp-server-auto-tls-certificate" ]; then
      extraArgs+=" $arg "
  fi
  if [ $arg == "continue" ]; then
    continue=true
  fi
done

str="genesis.data,"

go build main.go

if [ $continue == false ]; then

  # Let's delete old blockchain and verify if all nodes still have a genesis file
  genesisExists=true
  for ((i = 0; i < $nodes; ++i)); do
    echo "deleting $i"
    rm -r ./_build/webd2/devnet_$i/DEV/logs 2>/dev/null
    rm ./_build/webd2/devnet_$i/DEV/store/blockchain_store.bolt 2>/dev/null
    rm ./_build/webd2/devnet_$i/DEV/store/mempool_store.bolt 2>/dev/null

    if [ ! -e /_build/webd2/devnet_$i/DEV/genesis.data ]; then
      genesisExists=false
    fi
  done

  sleep 0.2

  # In case the genesis file is not found, let's create new wallets and generate the staked addresses files
  if [ $genesisExists == false ]; then

    for ((i = 0; i < $nodes; ++i)); do

      echo "delete wallet $i"
      rm ./_build/webd2/devnet_$i/DEV/store/wallet_store.bolt 2>/dev/null

      echo "running $i"
      go run main.go --instance="devnet" --instance-id="$i" --network="devnet" --wallet-export-shared-staked-address="auto,0,staked.address" --exit
      mv ./_build/webd2/devnet_$i/DEV/staked.address ./_build/webd2/devnet_0/DEV/$i.stake
      echo "executed"

    done

  fi

  for ((i = 0; i < $nodes; ++i)); do
    str+="$i.stake"

    if [ $i != $((nodes - 1)) ]; then
      str+=","
    fi
  done

  # A new genesis file will be created to restart the timestamp
  echo "creating genesis $str"
  go run main.go --instance="devnet" --instance-id="0" --network="devnet" --create-new-genesis="$str" --exit

  sleep 0.1

  echo "let's copy the genesis file to each node"
  for ((i = 1; i < $nodes; ++i)); do
    echo "copying genesis $i"
    cp ./_build/webd2/devnet_0/DEV/genesis.data ./_build/webd2/devnet_$i/DEV/genesis.data
  done

  echo "let's delete again the blockchain to restart"
  for ((i = 0; i < $nodes; ++i)); do
    rm ./_build/webd2/devnet_$i/DEV/store/blockchain_store.bolt 2>/dev/null
  done

  extraArgs+=" --skip-init-sync "
fi

sleep 0.1

#cloudflare public ports
screen -S app0 -dm go run main.go --instance="devnet" --instance-id="0" --tcp-server-port="2053" --new-devnet --run-testnet-script --network="devnet" --set-genesis="file" --forging --hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true" --delegator-enabled="true"  $extraArgs &
screen -S app1 -dm go run main.go --instance="devnet" --instance-id="1" --tcp-server-port="2083" --new-devnet --run-testnet-script --network="devnet" --set-genesis="file" --forging --hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true" --delegator-enabled="true"  $extraArgs &
screen -S app2 -dm go run main.go --instance="devnet" --instance-id="2" --tcp-server-port="2087" --new-devnet --run-testnet-script --network="devnet" --set-genesis="file" --forging --hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true" --delegator-enabled="true"  $extraArgs &
screen -S app3 -dm go run main.go --instance="devnet" --instance-id="3" --tcp-server-port="2096" --new-devnet --run-testnet-script --network="devnet" --set-genesis="file" --forging --hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true" --delegator-enabled="true"  $extraArgs &

wait

echo "finished"
exit 1
