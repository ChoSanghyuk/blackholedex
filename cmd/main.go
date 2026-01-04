package main

import (
	blackholedex "blackholego"
	"blackholego/configs"
	"blackholego/pkg/txlistener"
	"blackholego/pkg/util"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {

	// Get private key
	encryptedPk := os.Getenv("ENC_PK")
	if encryptedPk == "" {
		panic("PK not set")
	}

	key := os.Getenv("KEY")
	if key == "" {
		panic("PK not set")
	}

	pk, err := util.Decrypt([]byte(key), encryptedPk)
	if err != nil {
		panic(err)
	}

	conf, err := configs.LoadConfig("configs/config.yml")
	if err != nil {
		panic(err)
	}

	client, err := ethclient.Dial(conf.RPC)
	if err != nil {
		panic(err)
	}

	listener := txlistener.NewTxListener(
		client,
		txlistener.WithPollInterval(2*time.Second),
		txlistener.WithTimeout(5*time.Minute),
	)

	blackholeConf := conf.ToBlackholeConfigs()
	blackholeConf.Pk = pk
	blackhole, err := blackholedex.NewBlackhole(
		client,
		blackholeConf,
		listener,
	)
	if err != nil {
		panic(err)
	}

	strategyConf := conf.ToStrategyConfig()
	reportChan := make(chan string)
	go func() {
		err := blackhole.RunStrategy1(
			context.Background(),
			reportChan,
			strategyConf,
		)
		fmt.Printf("RunStrategy1 오류 발생. %s", err)
	}()

	for update := range reportChan {
		println(update)
	}

}
