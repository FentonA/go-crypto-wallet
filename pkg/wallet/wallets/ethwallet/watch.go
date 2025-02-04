package ethwallet

import (
	"database/sql"

	"go.uber.org/zap"

	"github.com/hiromaily/go-crypto-wallet/pkg/account"
	wtype "github.com/hiromaily/go-crypto-wallet/pkg/wallet"
	"github.com/hiromaily/go-crypto-wallet/pkg/wallet/api/ethgrp"
	"github.com/hiromaily/go-crypto-wallet/pkg/wallet/coin"
	"github.com/hiromaily/go-crypto-wallet/pkg/wallet/service"
	ethsrv "github.com/hiromaily/go-crypto-wallet/pkg/wallet/service/eth/watchsrv"
	"github.com/hiromaily/go-crypto-wallet/pkg/wallet/service/watchsrv"
)

// ETHWatch watch only wallet object
type ETHWatch struct {
	ETH    ethgrp.Ethereumer
	dbConn *sql.DB
	logger *zap.Logger
	wtype  wtype.WalletType
	watchsrv.AddressImporter
	ethsrv.TxCreator
	service.TxSender
	service.TxMonitorer
	service.PaymentRequestCreator
}

// NewETHWatch returns ETHWatch object
func NewETHWatch(
	eth ethgrp.Ethereumer,
	dbConn *sql.DB,
	logger *zap.Logger,
	addrImporter watchsrv.AddressImporter,
	txCreator ethsrv.TxCreator,
	txSender service.TxSender,
	txMonitorer service.TxMonitorer,
	paymentRequestCreator service.PaymentRequestCreator,
	wtype wtype.WalletType,
) *ETHWatch {
	return &ETHWatch{
		ETH:                   eth,
		logger:                logger,
		dbConn:                dbConn,
		wtype:                 wtype,
		AddressImporter:       addrImporter,
		TxCreator:             txCreator,
		TxSender:              txSender,
		TxMonitorer:           txMonitorer,
		PaymentRequestCreator: paymentRequestCreator,
	}
}

// ImportAddress imports address
func (w *ETHWatch) ImportAddress(fileName string, _ bool) error {
	return w.AddressImporter.ImportAddress(fileName)
}

// CreateDepositTx creates deposit unsigned transaction
func (w *ETHWatch) CreateDepositTx(_ float64) (string, string, error) {
	return w.TxCreator.CreateDepositTx()
}

// CreatePaymentTx creates payment unsigned transaction
func (w *ETHWatch) CreatePaymentTx(_ float64) (string, string, error) {
	return w.TxCreator.CreatePaymentTx()
}

// CreateTransferTx creates transfer unsigned transaction
func (w *ETHWatch) CreateTransferTx(sender, receiver account.AccountType, floatAmount, _ float64) (string, string, error) {
	return w.TxCreator.CreateTransferTx(sender, receiver, floatAmount)
}

// UpdateTxStatus updates transaction status
func (w *ETHWatch) UpdateTxStatus() error {
	return w.TxMonitorer.UpdateTxStatus()
}

// MonitorBalance monitors balance
func (w *ETHWatch) MonitorBalance(confirmationNum uint64) error {
	return w.TxMonitorer.MonitorBalance(confirmationNum)
}

// SendTx sends signed transaction
func (w *ETHWatch) SendTx(filePath string) (string, error) {
	return w.TxSender.SendTx(filePath)
}

// CreatePaymentRequest creates payment_request dummy data for development
func (w *ETHWatch) CreatePaymentRequest() error {
	amtList := []float64{
		0.001,
		0.002,
		0.0025,
		0.0015,
		0.003,
	}
	return w.PaymentRequestCreator.CreatePaymentRequest(amtList)
}

// Done should be called before exit
func (w *ETHWatch) Done() {
	w.dbConn.Close()
	w.ETH.Close()
}

// CoinTypeCode returns coin.CoinTypeCode
func (w *ETHWatch) CoinTypeCode() coin.CoinTypeCode {
	return w.ETH.CoinTypeCode()
}
