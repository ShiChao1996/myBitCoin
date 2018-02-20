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
 *     Initial: 2018/01/30        shichao
 */

package block

import (
	"time"
	"encoding/gob"
	"bytes"
	"myBitCoin/transaction"
	"myBitCoin/merkle"
)

type Block struct {
	TimeStamp int64
	PrevHash  []byte
	//Data      []byte
	Transactions []*transaction.Transaction
	Hash         []byte
	Nonce        int
}

/*
func (b *Block) SetHash() {
	timeStamp := []byte(strconv.FormatInt(b.TimeStamp, 10))
	headers := bytes.Join([][]byte{timeStamp, b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}*/

func NewBlock(transactions []*transaction.Transaction, prevHash []byte) *Block {
	b := &Block{time.Now().Unix(), prevHash, transactions, []byte{}, 0}
	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()
	b.Hash = hash
	b.Nonce = nonce
	return b
}

func NewGenesisBlock(coinbase *transaction.Transaction) *Block {
	return NewBlock([]*transaction.Transaction{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	encoder.Encode(b)
	return res.Bytes()
}

func DeSerialize(b []byte) *Block {
	var block *Block
	decoder := gob.NewDecoder(bytes.NewReader(b))
	decoder.Decode(&block)
	return block
}

func (b *Block) HashTransactions() []byte {
	var tr [][]byte

	for _, t := range b.Transactions {
		tr = append(tr, t.Serialize())
	}
	hashed := merkle.NewMerkleTree(tr)
	return hashed.RootNode.Data
}
