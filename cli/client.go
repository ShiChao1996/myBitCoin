/*
 * MIT License
 *
 * Copyright (c) 2017 SmartestEE Co., Ltd.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/*
 * Revision History:
 *     Initial: 2018/02/02        ShiChao
 */

package cli

import (
	"flag"
	"os"
	"fmt"
	"log"
	"strconv"
	blk "myBitCoin/block"
	"myBitCoin/transaction"
	"myBitCoin/wallet"
)

const usage = `
Usage:
  addblock -data BLOCK_DATA    add a block to the blockchain
  printchain                   print all the blocks of the blockchain
`

type Client struct {
	Bc *blk.BlockChain
}

func (cli *Client) printUsage() {
	fmt.Println(usage)
}

func (cli *Client) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *Client) createBlockchain(address, nodeID string) {
	fmt.Println("creating blockchain...")
	bc := blk.CreateBlockChain(address, nodeID)
	bc.DB.Close()
	fmt.Println("Done!")
}

/*
func (cli *Client) Run() {
	cli.validateArgs()

	var err error
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		err = addBlockCmd.Parse(os.Args[2:])
	case "printchain":
		err = printChainCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if err != nil {
		os.Exit(1)
		return
	}
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

}*/

func (cli *Client) Run() {
	cli.validateArgs()

	//nodeID := os.Getenv("NODE_ID")
	nodeID := os.Getenv("GOPATH")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!")
		os.Exit(1)
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.Send(*sendFrom, *sendTo, nodeID, *sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}
}

func (cli *Client) addBlock(data string) {
	//cli.Bc.AddBlock(data)
	fmt.Println("Success!")
}

func (cli *Client) printChain(nodeID string) {
	bc := blk.NewBlockChain(nodeID)
	defer bc.DB.Close()
	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		//fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blk.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *Client) getBalance(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Error: Address is not valid !")
	}
	bc := blk.NewBlockChain(nodeID)
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1: len(pubKeyHash)-4]
	UTXOs := bc.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *Client) Send(from, to, nodeID string, amount int) {
	if !wallet.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blk.NewBlockChain(nodeID)
	defer bc.DB.Close()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wlt := wallets.GetWallet(from)
	tx := bc.NewUTXOTransaction(wlt, to, amount)

	bc.AddBlock([]*transaction.Transaction{tx})
	fmt.Println("Success!")
}

func (cli *Client) createWallet(nodeID string) {
	wallets, _ := wallet.NewWallets(nodeID)
	address := wallets.CreateWallet()
	fmt.Println("nodeID: " + nodeID)
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}
