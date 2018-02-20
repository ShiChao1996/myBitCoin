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
 *     Initial: 2018/02/18        ShiChao
 */

package utxo

import (
	"github.com/boltdb/bolt"
	"encoding/hex"

	blk "myBitCoin/block"
	"myBitCoin/transaction"
	"log"
)

const utxoBucket = "chainstate"

type UTXOSet struct {
	BlockChain *blk.BlockChain
}

func (utxo *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspendOutputs := make(map[string][]int)
	accumulation := 0
	db := utxo.BlockChain.DB

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := transaction.DeserializeOutPuts(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					accumulation += out.Value
					unspendOutputs[txID] = append(unspendOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulation, unspendOutputs
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []transaction.TxOutput {
	var UTXOs []transaction.TxOutput
	db := u.BlockChain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := transaction.DeserializeOutPuts(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

func (utxo *UTXOSet) Reindex() {
	db := utxo.BlockChain.DB
	bucketName := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXOs := utxo.BlockChain.FindUTXO()
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		for txID, out := range UTXOs {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = bucket.Put(key, out.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}
/*
func (utxo *UTXOSet) Update(block *blk.Block) {
	db := utxo.BlockChain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))

		for _, tr := range block.Transactions {
			if tr.IsCoinbase() == false {
				for _, in := range tr.Vin {
					updatedOutputs := transaction.TxOutPuts{}
					outputBytes := bucket.Get(in.TxID)
					outputs := transaction.DeserializeOutPuts(outputBytes)

					for outIdx, out := range outputs.Outputs {
						if outIdx != in.Vout {
							updatedOutputs.Outputs = append(updatedOutputs.Outputs, out)
						}
					}

					if len(updatedOutputs.Outputs) == 0 {
						err := bucket.Delete(in.TxID)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := bucket.Put(in.TxID, updatedOutputs.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}

			newOutputs := transaction.TxOutPuts{}
			for _, out := range tr.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := bucket.Put(tr.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}*/

func (u UTXOSet) Update(block *blk.Block) {
	db := u.BlockChain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tr := range block.Transactions {
			if tr.IsCoinbase() == false {
				for _, vin := range tr.Vin {
					updatedOuts := transaction.TxOutPuts{}
					outsBytes := b.Get(vin.TxID)
					outs := transaction.DeserializeOutPuts(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.TxID)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.TxID, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := transaction.TxOutPuts{}
			for _, out := range tr.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tr.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
