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
	"fmt"
	"os"
	"encoding/hex"
	"log"
	"bytes"
	"errors"
	"myBitCoin/transaction"
	"myBitCoin/wallet"
	"crypto/ecdsa"
)

const (
	dbFile              = "%s/blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

type BlockChain struct {
	tip []byte
	DB  *bolt.DB
}

// 创建一个有创世块的新链
func NewBlockChain(nodeID string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
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

func CreateBlockChain(addr, nodeID string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
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

func dbExists(dbFile string) bool {
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

// FindTransaction finds a transaction by its ID
func (bc *BlockChain) FindTransaction(ID []byte) (transaction.Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.New("Transaction is not found")
}

func (c *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []transaction.Transaction {
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
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
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

/*func (c *BlockChain) FindUTXO(pubKeyHash []byte) []transaction.TxOutput {
	unspentTxs := c.FindUnspentTransactions(pubKeyHash)
	var outputs []transaction.TxOutput
	for _, tx := range unspentTxs {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				outputs = append(outputs, out)
			}
		}
	}
	return outputs
}*/
// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *BlockChain) FindUTXO() map[string]transaction.TxOutPuts {
	UTXO := make(map[string]transaction.TxOutPuts)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.TxID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO
}

func (c *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspendOutputs := make(map[string][]int)
	accumulation := 0
	unspentTr := c.FindUnspentTransactions(pubKeyHash)

Find:
	for _, tx := range unspentTr {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulation < amount {
				unspendOutputs[txID] = append(unspendOutputs[txID], outIdx)
				accumulation += out.Value
			}
			if accumulation >= amount {
				break Find
			}
		}
	}

	return accumulation, unspendOutputs
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

func (c *BlockChain) NewUTXOTransaction(wlt *wallet.Wallet, to string, amount int) *transaction.Transaction {
	var (
		outputs []transaction.TxOutput
		inputs  []transaction.TxInput
	)

	pubKeyHash := wallet.HashPubKey(wlt.PublicKey)
	acc, validOuts := c.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for id, outs := range validOuts {
		txID, _ := hex.DecodeString(id)

		for _, out := range outs {
			input := transaction.TxInput{txID, out, nil, wlt.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := wlt.GetAddress()
	outputs = append(outputs, transaction.NewTxOut(amount,to))
	if acc > amount {
		delta := acc - amount
		outputs = append(outputs, transaction.NewTxOut(delta, string(from)))
	}

	tx := &transaction.Transaction{nil, inputs, outputs}
	tx.SetID()

	return tx
}

func (bc *BlockChain)SignTransactions(tx transaction.Transaction, private ecdsa.PrivateKey)  {
	prevTxs := make(map[string]transaction.Transaction)

	for _,in := range tx.Vin{
		prev,err := bc.FindTransaction(in.TxID)
		if err != nil{
			log.Panic(err)
		}
		id := hex.EncodeToString(in.TxID)
		prevTxs[id] = prev
	}

	tx.Sign(private,prevTxs)
}

func (bc *BlockChain) VerifyTransaction(tr *transaction.Transaction)bool {
	prevTxs := make(map[string]transaction.Transaction)

	for _,in := range tr.Vin{
		prev,err := bc.FindTransaction(in.TxID)
		if err != nil{
			log.Panic(err)
		}
		id := hex.EncodeToString(in.TxID)
		prevTxs[id] = prev
	}

	return tr.Verify(prevTxs)
}