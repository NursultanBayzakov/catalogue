package catalogue

import (
	"catalogue-service/internal/data/models"
	"catalogue-service/internal/sl"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Catalogue struct {
	log               *slog.Logger
	catalogueProvider CatalogueProvider
	tokenTTL          time.Duration
}

func New(
	log *slog.Logger,
	catalogueProvider CatalogueProvider,
	tokenTtl time.Duration,
) *Catalogue {
	return &Catalogue{
		log:               log,
		catalogueProvider: catalogueProvider,
		tokenTTL:          tokenTtl,
	}
}

type CatalogueProvider interface {
	SaveItem(
		ctx context.Context,
		item *models.Item,
	) (int32, error)
	GetAllItems(
		context.Context,
	) ([]*models.Item, error)
	GetItemById(
		context.Context,
		int,
	) (*models.Item, error)
}

func (c *Catalogue) CreateItem(ctx context.Context, item *models.Item) (int32, error) {
	const op = "Catalogue.CreateItem"

	log := c.log.With(
		slog.String("op", op),
		slog.String("item name", item.Name),
	)

	log.Info("attempting to create item")

	id, err := c.catalogueProvider.SaveItem(ctx, item)
	if err != nil {
		c.log.Warn("failed to save item", sl.Err(err))
		return 0, fmt.Errorf("%s", op)
	}

	return id, nil
}

func (c *Catalogue) ListItems(ctx context.Context) ([]*models.Item, error) {
	const op = "Catalogue.ListItems"
	log := c.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to get all items")

	items, err := c.catalogueProvider.GetAllItems(ctx)
	if err != nil {
		c.log.Warn("failed to get all items", sl.Err(err))
		return nil, err
	}

	return items, nil
}

func (c *Catalogue) GetItem(ctx context.Context, id int) (*models.Item, error) {
	const op = "Catalogue.GetItem"
	log := c.log.With(
		slog.String("op", op),
		slog.Int("item id", id),
	)

	log.Info("attempting to get item")
	item, err := c.catalogueProvider.GetItemById(ctx, id)
	if err != nil {
		c.log.Warn("failed to get item", sl.Err(err))
		return nil, err
	}

	return item, nil
}
