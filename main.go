package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

var rpcFlag = flag.String("rpc", "", "ethereum rpc URL")
var gweiFlag = flag.Float64("gwei", 1.0, "gas price threshold to execute command at (default 1gwei)")
var intervalFlag = flag.Int64("interval", 60, "interval between price checks in seconds")

func main() {

	// parse flags
	flag.Parse()
	cmd := strings.Join(flag.Args(), " ")
	if cmd == "" {
		fmt.Printf("Usage: sublimate [flags] cmd\n")
		os.Exit(1)
	}
	rpcURL := os.Getenv("RPC_URL")
	if *rpcFlag != "" {
		rpcURL = *rpcFlag
	}
	if rpcURL == "" {
		fmt.Printf("Usage: sublimate [cmd]\n")
		os.Exit(1)
	}

	// execute command once gas price goes below target threshold
	cmdOutput, err := run(cmd, rpcURL, *gweiFlag, *intervalFlag)
	if err != nil {
		panic(err)
	}

	// send command output over Stdout so this works with pipes
	fmt.Fprintf(os.Stdout, "%s", string(cmdOutput))
}

func run(cmd string, rpcURL string, gasTargetGwei float64, interval int64) ([]byte, error) {

	var (
		rpcClient *ethclient.Client
		rpcError  error
	)
	rpcClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dialing RPC: %w", err)
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for ; true; <-ticker.C { // do-while esque trick with tickers

		// attempt to reconnect rpc if connection was severed
		if rpcError != nil {
			rpcClient, err = ethclient.Dial(rpcURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to dial RPC. Retrying at next tick...: %v\n", err)
				rpcError = err
				continue
			}
		}
		rpcError = nil

		// fetch the current gas price
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		gasPrice, err := rpcClient.SuggestGasPrice(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to fetch gas price. Retrying at next tick...: %v\n", err)
			rpcError = err
			continue
		}
		priceGweiF, _ := WeiToGwei(gasPrice).Float64()

		// wait till next tick if too expensive
		if priceGweiF > gasTargetGwei {
			fmt.Fprintf(os.Stderr, "currentGasPrice: %.02f gwei, target: %.02f gwei, price too high waiting...\n", priceGweiF, gasTargetGwei)
			continue
		}

		// TODO: deal with stderr output from the run command
		// execute command
		fmt.Fprintf(os.Stderr, "gas threshold met, executing cmd...\n")
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			return out, fmt.Errorf("Failed to execute command: %s", cmd)
		}
		return out, nil
	}

	return nil, nil
}

func WeiToGwei(val *big.Int) *big.Float {
	return new(big.Float).Quo(
		new(big.Float).SetInt(val),
		big.NewFloat(params.GWei),
	)
}
