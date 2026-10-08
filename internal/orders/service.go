package orders

import (
	"context"
	"errors"
	"fmt"

	repo "ecom/internal/adapters/postgresql/sqlc"

	"github.com/jackc/pgx/v5"
)

var (
	ErrProductNotFound = errors.New("Product not found")
	ErrProductNoStock = errors.New("Product has not enought stock")
)

type OrderServices interface {
	PlaceOrder(ctx context.Context, orderData createOrderParams) (repo.Order, error)
	FindOrderByID(ctx context.Context, id int64) (repo.Order, error)
}

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewService(repo *repo.Queries, db *pgx.Conn) OrderServices {
	return &svc{repo: repo, db: db}
}

func (s *svc) PlaceOrder(ctx context.Context, orderData createOrderParams) (repo.Order, error) {
	if orderData.CustomerID == 0 {
		return repo.Order{}, fmt.Errorf("Cusomer ID is required")
	}
	if len(orderData.Items) == 0 {
		return repo.Order{}, fmt.Errorf("At least one Item is required")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return repo.Order{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.repo.WithTx(tx)

	order, err := qtx.CreateOrder(ctx, orderData.CustomerID)
	if err != nil {
		return repo.Order{}, err
	}

	for _, item := range orderData.Items {
		product, err := qtx.FindProductByID(ctx, item.ProductID)
		if err != nil {
			return repo.Order{}, ErrProductNotFound
		}
		if product.Quantity < item.Quantity {
			return repo.Order{}, ErrProductNoStock
		}

		_, err = qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			PriceCent: product.PriceInCenters,
		})
		if err != nil {
			return repo.Order{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return repo.Order{}, err
	}

	return order, nil
}

func (s *svc) FindOrderByID(ctx context.Context, id int64) (repo.Order, error) {
	return repo.Order{}, fmt.Errorf("not implemented")
}
