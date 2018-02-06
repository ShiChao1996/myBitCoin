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
 *     Initial: 2018/01/30        ShiChao
 */

package block

import (
	"github.com/boltdb/bolt"
	"myBitCoin/transaction"
	"fmt"
	"os"
	"encoding/hex"
	"log"
)

const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

type BlockChain struct {
	tip []byte
	DB  *bolt.DB
}

// 创建一个有创世块的新链
func NewBlockChain(address string) *BlockChain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db}

	return &bc
}

func CreateBlockChain(addr string) *BlockChain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			cbtx := transaction.NewCoinbaseTx(addr, genesisCoinbaseData)
			genesis := NewGenesisBlock(cbtx)

			b, _ = tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	bc := &BlockChain{tip, db}
	return bc
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func (c *BlockChain) AddBlock(transactions []*transaction.Transaction) {
	var lastHash []byte
	c.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	newBlock := NewBlock(transactions, lastHash)

	c.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		b.Put(newBlock.Hash, newBlock.Serialize())
		b.Put([]byte("l"), newBlock.Hash)
		c.tip = newBlock.Hash
		return nil
	})
}

func (c *BlockChain) FindUnspentTransactions(address string) []transaction.Transaction {
	var unspentTXs []transaction.Transaction
	spentTXOS := make(map[string][]int)
	bci := c.Iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Flag:
			for outIdx, out := range tx.Vout {
				if spentTXOS[txID] != nil {
					for _, spentOut := range spentTXOS[txID] {
						if spentOut == outIdx {
							continue Flag
						}
					}
				}
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.TxID)
						spentTXOS[inTxID] = append(spentTXOS[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (c *BlockChain) FindUTXO(address string) []transaction.TxOutput {
	unspentTxs := c.FindUnspentTransactions(address)
	var outputs []transaction.TxOutput
	for _, tx := range unspentTxs {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				outputs = append(outputs, out)
			}
		}
	}
	return outputs
}

func (c *BlockChain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{c.tip, c.DB}
}

// iterator the chain
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeSerialize(encodedBlock)
		return nil
	})
	i.currentHash = block.PrevHash
	return block
}
