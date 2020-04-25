package repository

import (
	"context"
	"database/sql"
	models "github.com/hiromaily/go-bitcoin/pkg/models/rdb"
	"github.com/hiromaily/go-bitcoin/pkg/wallet/coin"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"go.uber.org/zap"
)

type PaymentRequestRepository interface {
	GetAll() ([]*models.PaymentRequest, error)
	GetAllByPaymentID(paymentID int64) ([]*models.PaymentRequest, error)
	InsertBulk(items []*models.PaymentRequest) error
	UpdatePaymentID(paymentID int64, ids []int64) (int64, error)
	UpdateIsDone(paymentID int64) (int64, error)
}

type paymentRequestRepository struct {
	dbConn       *sql.DB
	tableName    string
	coinTypeCode coin.CoinTypeCode
	logger       *zap.Logger
}

// NewPaymentRequestRepository returns NewPaymentRequestRepository interface
func NewPaymentRequestRepository(dbConn *sql.DB, coinTypeCode coin.CoinTypeCode, logger *zap.Logger) PaymentRequestRepository {
	return &paymentRequestRepository{
		dbConn:       dbConn,
		tableName:    "payment_request",
		coinTypeCode: coinTypeCode,
		logger:       logger,
	}
}

// GetAll returns all records whose payment_id is null
// - replaced from GetPaymentRequestAll
func (r *paymentRequestRepository) GetAll() ([]*models.PaymentRequest, error) {
	//sql := "SELECT * FROM payment_request WHERE payment_id IS NULL"
	ctx := context.Background()

	prItems, err := models.PaymentRequests(
		qm.Where("coin=?", r.coinTypeCode.String()),
		qm.And("payment_id IS NULL"),
	).All(ctx, r.dbConn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call models.PaymentRequests().All()")
	}
	return prItems, nil
}

// GetAllByPaymentID returns all records searched by payment_id
// - replaced from GetPaymentRequestByPaymentID
func (r *paymentRequestRepository) GetAllByPaymentID(paymentID int64) ([]*models.PaymentRequest, error) {
	//sql := "SELECT * FROM payment_request WHERE payment_id=?"
	ctx := context.Background()

	prItems, err := models.PaymentRequests(
		qm.Where("coin=?", r.coinTypeCode.String()),
		qm.And("payment_id=?", paymentID),
	).All(ctx, r.dbConn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call models.PaymentRequests().All()")
	}
	return prItems, nil
}

// Insert inserts multiple records
// - replaced from InsertPaymentRequest()
func (r *paymentRequestRepository) InsertBulk(items []*models.PaymentRequest) error {
	ctx := context.Background()
	return models.PaymentRequestSlice(items).InsertAll(ctx, r.dbConn, boil.Infer())
}

// UpdatePaymentID updates isDone
// - replaced from UpdatePaymentIDOnPaymentRequest
func (r *paymentRequestRepository) UpdatePaymentID(paymentID int64, ids []int64) (int64, error) {
	//sql := `UPDATE payment_request SET payment_id=? WHERE id IN (?)`
	ctx := context.Background()

	// Set updating columns
	updCols := map[string]interface{}{
		models.PaymentRequestColumns.PaymentID: paymentID,
	}

	// change []int64 to []interface
	targetIDs := make([]interface{}, len(ids))
	for i, v := range ids {
		targetIDs[i] = v
	}

	return models.PaymentRequests(
		qm.WhereIn("id IN ?", targetIDs...),
	).UpdateAll(ctx, r.dbConn, updCols)
}

// UpdateIsDone updates isDone
// - replaced from UpdateTxTypeNotifiedByID
func (r *paymentRequestRepository) UpdateIsDone(paymentID int64) (int64, error) {
	//sql := `UPDATE payment_request SET is_done=true WHERE payment_id=?`
	ctx := context.Background()

	// Set updating columns
	updCols := map[string]interface{}{
		models.PaymentRequestColumns.IsDone: true,
	}
	return models.PaymentRequests(
		qm.Where("payment_id=?", paymentID),
	).UpdateAll(ctx, r.dbConn, updCols)
}
