package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/ethereum/go-ethereum/common"

	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(colorable.NewColorableStdout())
	log.SetFormatter(&log.TextFormatter{
		DisableColors:          false,
		DisableLevelTruncation: true,
		ForceColors:            true,
		FullTimestamp:          true,
		PadLevelText:           true,
		TimestampFormat:        "2006-01-02 15:04:05.00000",
	})
	log.SetLevel(log.DebugLevel)
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigs
		log.Warnln("Test interrupted by the user")
		log.Exit(0)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := ethclient.DialContext(ctx, "ws://localhost:9650/ext/bc/C/ws")
	// client, err := ethclient.DialContext(ctx, "wss://api.avax.network/ext/bc/C/ws")
	// client, err := ethclient.DialContext(ctx, "https://api.avax.network/ext/bc/C/rpc")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	Mempool(client)
}

func Mempool(client ethclient.Client) {
	// TraderJoe router
	tjRouter := common.HexToAddress("0x60aE616a2155Ee3d9A68541Ba4544862310933d4")
	log.Debugf("🔵 TraderJOE Router: %+v", tjRouter)

	mpChan := make(chan *common.Hash)
	sub, err := client.SubscribeNewPendingTransactions(context.Background(), mpChan)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vMempool := <-mpChan:
			tx, pending, err := client.TransactionByHash(context.Background(), *vMempool)
			if err != nil {
				log.Errorf("⛔️ TransactionByHash: %v", err)
			} else {
				msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), nil)
				if err != nil {
					log.Errorf("⛔️ AsMessage: %v", err)
					continue
				}
				if *msg.To() == tjRouter {
					log.Infof("✅ [TX to TJ] pending:%t - TxType:%d - From:%v - To:%v - TxValue:%d - Data:%v", pending, tx.Type(), msg.From(), msg.To(), msg.Value(), msg.Data())
				}
				// if msg.From() == tjRouter {
				// 	log.Infof("☑️ [TX from TJ] pending:%t - TxType:%d - From:%v - To:%v - TxValue:%d - Data:%v", pending, tx.Type(), msg.From(), msg.To(), msg.Value(), msg.Data())
				// }
			}
		}
	}
}
