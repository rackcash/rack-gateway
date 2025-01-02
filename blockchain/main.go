package main

import (
	"infra/blockchain/rack/config"
	"infra/blockchain/rack/currencies"
	"infra/blockchain/rack/currencies/eth"
	"infra/blockchain/rack/currencies/sol"
	"infra/blockchain/rack/currencies/ton"
	"infra/blockchain/rack/nats"
	"infra/pkg/dlog"
	"sync/atomic"
)

func main() {

	dlog := dlog.Init()

	var curs = new(atomic.Int32)

	config := config.ReadConfig()
	ns, c := nats.Init(config)

	go currencies.UpdateRates(config)

	//  TON INIT
	tonClient, blockid := ton.Init(config, curs)

	tonConfig := nats.Ton{Client: tonClient}
	tonConfig.BlockID.Store(blockid)

	go ton.RunUpdateBlock(tonClient, &tonConfig.BlockID)

	// SOLANA INIT

	solClient, solWs := sol.Init(config, curs)
	solConfig := nats.Sol{Client: solClient}
	solConfig.Ws.Store(solWs)

	// go sol.RunUpdateWs(&solConfig.Ws)

	// ETHEREUM

	ethClient := eth.Init(config, curs)

	app := nats.App{
		Ton:       &tonConfig,
		EthClient: ethClient,
		Config:    config,
		Sol:       &solConfig,
		// SolClient: solClient,
		// SolWs:     solWs,
		Ns:   ns,
		C:    c,
		Dlog: dlog,
	}

	app.Run(config, ns)
}
