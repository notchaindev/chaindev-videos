package main

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"sync"
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

	FilterLogs(client)
}

func FilterLogs(client ethclient.Client) {
	// TraderJoe factory
	tjFactory := common.HexToAddress("0x9Ad6C38BE94206cA50bb0d90783181662f0Cfa10")
	log.Debugf("TraderJOE Factory: %v", tjFactory)

	// createPair[signature]
	createPair := common.BytesToHash(signature2ID("PairCreated(address,address,address,uint256)"))
	log.Debugf("TJ Factory PairCreated: %v", createPair)

	// FromBlock 2486392
	startBlock := int64(2486392)

	headerNumber, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	latestBlock := headerNumber.Number.Int64()

	totalBlock := latestBlock - startBlock

	currentBlock := startBlock
	var toBlock int64
	const PAGINATION = 2048
	eventNumber := 0
	wg := sync.WaitGroup{}
	for currentBlock < latestBlock {
		if currentBlock+PAGINATION > latestBlock {
			toBlock = latestBlock
		} else {
			toBlock = currentBlock + PAGINATION
		}

		query := interfaces.FilterQuery{
			Addresses: []common.Address{tjFactory},
			Topics: [][]common.Hash{
				{createPair},
			},
			FromBlock: big.NewInt(currentBlock),
			ToBlock:   big.NewInt(toBlock),
		}
		log.Debugf("query: %v", query)

		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			log.Fatal(err)
		}

		for _, vLog := range logs {
			wg.Add(1)
			eventNumber++
			go func(vlog types.Log) {
				defer wg.Done()
				log.Infof("Log: %v", vlog)
			}(vLog)
		}
		log.Infof("ETA: %.3f %%", float64(toBlock-startBlock)*100/float64(totalBlock))
		currentBlock += PAGINATION
	}
	wg.Wait()
	log.Infof("eventNumber: %d", eventNumber)
}

func signature2ID(sig string) []byte {
	signature := []byte(sig)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(signature)
	return hash.Sum(nil)
}
