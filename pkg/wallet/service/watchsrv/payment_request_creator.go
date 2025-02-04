package watchsrv

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"go.uber.org/zap"

	"github.com/hiromaily/go-crypto-wallet/pkg/account"
	"github.com/hiromaily/go-crypto-wallet/pkg/converter"
	models "github.com/hiromaily/go-crypto-wallet/pkg/models/rdb"
	"github.com/hiromaily/go-crypto-wallet/pkg/repository/watchrepo"
	"github.com/hiromaily/go-crypto-wallet/pkg/wallet"
	"github.com/hiromaily/go-crypto-wallet/pkg/wallet/coin"
)

// PaymentRequestCreate type
type PaymentRequestCreate struct {
	converter    converter.Converter
	logger       *zap.Logger
	dbConn       *sql.DB
	addrRepo     watchrepo.AddressRepositorier
	payReqRepo   watchrepo.PaymentRequestRepositorier
	coinTypeCode coin.CoinTypeCode
	wtype        wallet.WalletType
}

// NewPaymentRequestCreate returns PaymentRequestCreate object
func NewPaymentRequestCreate(
	converter converter.Converter,
	logger *zap.Logger,
	dbConn *sql.DB,
	addrRepo watchrepo.AddressRepositorier,
	payReqRepo watchrepo.PaymentRequestRepositorier,
	coinTypeCode coin.CoinTypeCode,
	wtype wallet.WalletType,
) *PaymentRequestCreate {
	return &PaymentRequestCreate{
		converter:    converter,
		logger:       logger,
		dbConn:       dbConn,
		addrRepo:     addrRepo,
		payReqRepo:   payReqRepo,
		coinTypeCode: coinTypeCode,
		wtype:        wtype,
	}
}

// CreatePaymentRequest creates payment_request dummy data for development
func (p *PaymentRequestCreate) CreatePaymentRequest(amtList []float64) error {
	// get client pubkeys
	pubkeyItems, err := p.addrRepo.GetAll(account.AccountTypeClient)
	if err != nil {
		return errors.Wrap(err, "fail to call addrRepo.GetAll()")
	}
	if len(pubkeyItems) < len(amtList)*2 {
		return errors.Errorf("pubkey for client should be created at least %d", len(amtList))
	}
	// start transaction
	dtx, err := p.dbConn.Begin()
	if err != nil {
		return errors.Wrap(err, "fail to start transaction")
	}
	defer func() {
		if err != nil {
			dtx.Rollback()
		} else {
			dtx.Commit()
		}
	}()

	// delete payment request
	_, err = p.payReqRepo.DeleteAll()
	if err != nil {
		return errors.Wrap(err, "fail to call payReqRepo.DeleteAll()")
	}
	// insert payment_request
	payReqItems := make([]*models.PaymentRequest, 0, len(amtList))
	var idx int
	for _, amt := range amtList {
		payReqItems = append(payReqItems, &models.PaymentRequest{
			Coin:            p.coinTypeCode.String(),
			PaymentID:       null.NewInt64(0, false),
			SenderAddress:   pubkeyItems[0+idx].WalletAddress,
			SenderAccount:   pubkeyItems[0+idx].Account,
			ReceiverAddress: pubkeyItems[len(amtList)+idx].WalletAddress,
			Amount:          p.converter.FloatToDecimal(amt),
			IsDone:          false,
			UpdatedAt:       null.TimeFrom(time.Now()),
		})
		idx++
	}
	if err = p.payReqRepo.InsertBulk(payReqItems); err != nil {
		return errors.Wrap(err, "fail to call payReqRepo.InsertBulk()")
	}
	return nil
}
