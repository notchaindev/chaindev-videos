package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"

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

	RealtimeEventLogs(client)
}

func RealtimeEventLogs(client ethclient.Client) {
	USDC := common.HexToAddress("0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E")
	log.Debugf("USDC: %v", USDC)

	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	log.Debugf("Transfer topic: %v", transferTopic)

	approvalTopic := common.BytesToHash(signature2ID("Approval(address,address,uint256)"))
	log.Debugf("Approval topic: %v", approvalTopic)

	query := interfaces.FilterQuery{
		Addresses: []common.Address{USDC},
		Topics: [][]common.Hash{
			{approvalTopic, transferTopic},
		},
	}
	log.Debugf("query: %v", query)
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			log.Infof("Log: %v", vLog)
		}
	}
}

func signature2ID(sig string) []byte {
	signature := []byte(sig)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(signature)
	return hash.Sum(nil)
}
