package data

import (
	"catalogue-service/internal/data/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"strings"
	"time"
)

type ItemRepo struct {
	DB *sql.DB
}

var (
	ErrRecordNotFound   = errors.New("record (row, entry) not found")
	ErrItemAlreadyExist = errors.New("item already exists")
)

func (ir *ItemRepo) SaveItem(ctx context.Context, item *models.Item) (int32, error) {
	const op = "data.SaveItem"
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}
	query := `INSERT INTO catalogue.item_info (name, price, description, quantity, image_url)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`
	args := []interface{}{
		item.Name,
		item.Price,
		item.Description,
		item.Quantity,
		item.ImageURL,
	}

	tx, err := ir.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fail(err)
	}
	defer tx.Rollback()

	err = ir.DB.QueryRowContext(ctx, query, args...).Scan(&item.ID)
	if err != nil || item.ID == 0 {
		var pqErr *pq.Error
		switch {
		case errors.As(err, &pqErr) && pqErr.Code == "23505" && strings.Contains(pqErr.Message, "item_info_name_key"):
			return 0, ErrItemAlreadyExist
		case errors.Is(err, sql.ErrNoRows):
			return 0, fail(err)
		default:
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fail(err)
	}
	return item.ID, nil
}

func (ir *ItemRepo) GetItemById(ctx context.Context, id int) (*models.Item, error) {
	const op = "data.GetItemById"
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}
	var item models.Item
	query := `SELECT * FROM catalogue.item_info
			WHERE id = $1`
	err := ir.DB.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Price,
		&item.Description,
		&item.Quantity,
		&item.ImageURL,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fail(ErrRecordNotFound)
		default:
			return nil, fail(err)
		}
	}
	return &item, nil
}

func (ir *ItemRepo) GetAllItems(ctx context.Context) ([]*models.Item, error) {
	const op = "data.GetAllItems"
	fail := func(e error) error {
		return fmt.Errorf("%s, %v", op, e)
	}
	query := `SELECT * FROM catalogue.item_info`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := ir.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fail(err)
	}

	var items []*models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Price,
			&item.Description,
			&item.Quantity,
			&item.ImageURL,
		)
		if err != nil {
			return nil, fail(err)
		}

		items = append(items, &item)
	}

	return items, nil
}

func (ir *ItemRepo) Delete(ctx context.Context, id int) error {
	const op = "data.Delete"
	fail := func(e error) error {
		return fmt.Errorf("%s: %v", op, e)
	}
	query := `
			DELETE FROM catalogue.item_info
			WHERE id = $1`
	tx, err := ir.DB.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return fail(err)
	}

	exec, err := ir.DB.Exec(query, id)
	if err != nil {
		return err
	}

	affected, err := exec.RowsAffected()
	if err != nil || affected == 0 {
		return fail(err)
	}

	if err = tx.Commit(); err != nil {
		return fail(err)
	}

	return nil
}
