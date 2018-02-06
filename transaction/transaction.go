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

package transaction

import (
	"fmt"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"bytes"
)

const subsidy = 10

type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

type TxOutput struct {
	Value        int
	ScriptPubKey string
}

type TxInput struct {
	TxID      []byte
	Vout      int
	ScriptSig string
}

func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txIn := TxInput{[]byte{}, -1, data}
	txOut := TxOutput{subsidy, to}
	tx := Transaction{nil, []TxInput{txIn}, []TxOutput{txOut}}
	tx.SetID()

	return &tx
}

func (tx *Transaction) SetID() {
	var (
		encode bytes.Buffer
		hash   [32]byte
	)
	encoder := gob.NewEncoder(&encode)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encode.Bytes())
	tx.ID = hash[:]
}

// IsCoinbase 判断是否是 coinbase 交易
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TxID) == 0 && tx.Vin[0].Vout == -1
}

func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
