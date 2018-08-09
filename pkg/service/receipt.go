package service

import (
	"github.com/bookerzzz/grok"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"log"
)

// DetectReceivedCoin Wallet内アカウントに入金があれば、そこから、未署名のトランザクションを返す？？
func (w *Wallet) DetectReceivedCoin() (*wire.MsgTx, error) {
	log.Println("[DetectReceivedCoin]")
	//TODO:このロジックを連続で走らせた場合、現在処理中のものが、タイミングによってはまた取得できてしまうので、そこを考慮しないといけない

	//TODO:LockUnspent
	//TODO:ここはgoroutineで並列化されたタスク内で、ロックされたtxidを監視し、confirmationが6になったら、解除するようにしたほうがいいかも。
	if err := w.Btc.UnlockAllUnspentTransaction(); err != nil {
		return nil, err
	}

	//1. アカウント一覧からまとめて残高を取得
	//[]btcjson.ListUnspentResult
	// ListUnspentResult models a successful response from the listunspent request.
	//type ListUnspentResult struct {
	//	TxID          string  `json:"txid"`
	//	Vout          uint32  `json:"vout"`
	//	Address       string  `json:"address"`
	//	Account       string  `json:"account"`
	//	ScriptPubKey  string  `json:"scriptPubKey"`
	//	RedeemScript  string  `json:"redeemScript,omitempty"`
	//	Amount        float64 `json:"amount"`
	//	Confirmations int64   `json:"confirmations"`
	//	Spendable     bool    `json:"spendable"`
	//}
	//TODO:とりあえず、ListUnspentを使っているが、全ユーザーにGetUnspentByAddress()を使わないといけないかも()
	list, err := w.Btc.Client().ListUnspent()
	if err != nil {
		return nil, errors.Errorf("ListUnspent(): error: %v", err)
	}
	log.Printf("List Unspent: %v\n", list)
	grok.Value(list)

	if len(list) == 0 {
		return nil, nil
	}

	//
	var total btcutil.Amount
	var inputs []btcjson.TransactionInput
	//CreateRawTransaction()は外で実行する
	//Loop内ではパラメータを作成するのみ
	for _, tx := range list {
		//FIXME: pendableは実環境では使えない。とりあえず、confirmation数でチェックにしておく
		//https://bitcoin.stackexchange.com/questions/63198/why-outputs-spendable-and-solvable-are-false
		if tx.Confirmations < 6 {
			//if tx.Spendable == false {
			continue
		}

		// Transaction詳細を取得(必要な情報があるかどうか不明)
		tran, err := w.Btc.GetTransactionByTxID(tx.TxID)
		if err != nil {
			//txIDがおかしいはず
			continue
		}
		//log.Printf("Transactions: %v\n", tran)
		//grok.Value(tran)

		//除外するアカウント
		if tran.Details[0].Account == "hokan" || tran.Details[0].Account == "" {
			continue
		}

		// Amount
		// Satoshiに変換しないといけない
		// 1Satoshi＝0.00000001BTC
		// TODO:ここで変換は必要ないはず、と思ったがfloatの計算っておかしいんだっけ？
		amt, err := btcutil.NewAmount(tx.Amount)
		if err != nil {
			//TODO:このタイミングでエラーはおきないはず
			log.Println(err)
			continue
		}
		//TODO:全額は送金できないので、このタイミングで手数料を差し引かねばならないが、ここでいいんだっけ？total算出後でもいい？
		total += amt //合計

		//TODO:lockunspent
		if w.Btc.LockUnspent(tx) != nil {
			continue
		}
		//TODO:このトランザクションIDはDBに保存が必要
		//func (d *LevelDB) Put(prefix, key string, val []byte) error{
		err = w.Db.Put("unspent", tx.TxID+string(tx.Vout), nil)
		if err != nil {
			//TODO:このタイミングでエラーはおきないはず
			log.Println(err)
			continue
		}

		// inputs
		inputs = append(inputs, btcjson.TransactionInput{
			Txid: tx.TxID,
			Vout: tx.Vout,
		})
	}
	log.Printf("Total Coin to send:%d(Satoshi), Length: %d", total, len(inputs))
	if len(inputs) == 0 {
		return nil, nil
	}

	//TODO: このタイミングで手数料を算出して、totalから差し引く？？
	//total = 18500000 //桁を間違えた。。。0.185 BTC
	//total = 195000000

	// CreateRawTransaction
	//FIXME:おつりがあるのであれば、おつりのトランザクションも作らないといけないし、このfuncのインターフェースを見直す必要がある
	msgTx, err := w.Btc.CreateRawTransaction(HokanAddress, total, inputs) //hokanのアドレス
	if err != nil {
		return nil, errors.Errorf("CreateRawTransaction(): error: %v", err)
	}
	log.Printf("CreateRawTransaction: %v\n", msgTx)
	//grok.Value(msgTx)

	//TODO:fundrawtransactionによって手数料を算出したほうがいい。
	//https://bitcoincore.org/en/doc/0.16.2/rpc/rawtransactions/fundrawtransaction/

	//TODO:本来、これをDumpして、どっかに保存する必要があるはず、それをUSBに入れてコールドウォレットに移動しなくてはいけない
	//Feeもこのタイミングで取得する？？

	return msgTx, nil
}
