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
	"crypto/rand"
	"myBitCoin/wallet"
	"crypto/ecdsa"
	"encoding/hex"
)

const subsidy = 10

type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TxOutput) Lock(addr []byte) {
	hashPubKey := wallet.Base58Decode(addr)
	hashPubKey = hashPubKey[1:len(hashPubKey)-4]
	out.PubKeyHash = hashPubKey
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTxOut(value int, address string) TxOutput {
	txo := TxOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

type TxInput struct {
	TxID      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	fmt.Println(data)
	txIn := TxInput{[]byte{}, -1, nil, []byte(data)}
	txOut := NewTxOut(subsidy, to) //TxOutput{subsidy, to}
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

/*func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}*/

func (tr *Transaction) Sign(privKey ecdsa.PrivateKey, txs map[string]Transaction) {
	if tr.IsCoinbase() {
		return
	}

	for _, in := range tr.Vin {
		if txs[hex.EncodeToString(in.TxID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tr.TrimmedCopy()
	for inID,in := range txCopy.Vin{
		prevTx := txs[hex.EncodeToString(in.TxID)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[in.Vout].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tr.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil
	}
}

func (tr *Transaction) TrimmedCopy() Transaction {
	var (
		inputs  []TxInput
		outputs []TxOutput
	)

	for _, in := range tr.Vin {
		inputs = append(inputs, TxInput{in.TxID, in.Vout, nil, nil})
	}

	for _, out := range tr.Vout {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	return Transaction{tr.ID, inputs, outputs}
}
