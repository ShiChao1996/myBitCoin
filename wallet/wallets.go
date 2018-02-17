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
 *     Initial: 2018/02/13        ShiChao
 */

package wallet

import (
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"encoding/gob"
	"crypto/elliptic"
	"bytes"
)

const walletFile = "%s/wallet_.dat"

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallets(nodeId string) (*Wallets, error) {
	wallets := &Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile(nodeId)
	return wallets, err
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := string(wallet.GetAddress())
	ws.Wallets[address] = wallet
	return address
}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws *Wallets) GetWallet(address string) *Wallet {
	return ws.Wallets[address]
}

func (ws *Wallets) LoadFromFile(nodeId string) error {
	file := fmt.Sprintf(walletFile, nodeId)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return err
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets
	return nil
}

func (ws *Wallets) SaveToFile(nodeId string) {
	file := fmt.Sprintf(walletFile, nodeId)
	fmt.Println("file: " + file)
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(file, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
