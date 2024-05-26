package catalogueGrpc

import (
	"catalogue-service/internal/data/models"
	"context"
	"errors"
	cataloguev1 "github.com/bxiit/protos/gen/go/catalogue"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

var ErrInvalidCredentials = errors.New("item not found")

type Catalogue interface {
	CreateItem(
		ctx context.Context,
		item *models.Item,
	) (int32, error)
	ListItems(
		context.Context,
	) ([]*models.Item, error)
	GetItem(
		context.Context,
		int,
	) (*models.Item, error)
}

type catalogueService struct {
	cataloguev1.UnimplementedCatalogueServiceServer
	catalogue Catalogue
}

// Register - for registering gRPC server
func Register(gRPCServer *grpc.Server, catalogue Catalogue) {
	cataloguev1.RegisterCatalogueServiceServer(gRPCServer, &catalogueService{catalogue: catalogue})
}

func (cs *catalogueService) CreateItem(ctx context.Context, req *cataloguev1.CreateItemRequest) (*cataloguev1.CreateItemResponse, error) {
	if req.Item.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Item.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}
	if req.Item.Quantity < 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity cannot be negative")
	}
	if req.Item.Price < 0 {
		return nil, status.Error(codes.InvalidArgument, "price cannot be negative")
	}

	// Create a new item
	var item models.Item
	err := copier.Copy(&item, req.Item)
	if err != nil {
		log.Fatalf("failed to copy %v", err)
		return nil, err
	}

	id, err := cs.catalogue.CreateItem(ctx, &item)
	if err != nil {
		return nil, status.Error(codes.Internal, "error with create item")
	}

	req.Item.Id = id

	// Return the response
	return &cataloguev1.CreateItemResponse{Item: req.Item}, nil
}

func (cs *catalogueService) ListItems(ctx context.Context, req *cataloguev1.ListItemsRequest) (*cataloguev1.ListItemsResponse, error) {
	var responseItems []*cataloguev1.Item

	items, err := cs.catalogue.ListItems(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		var it cataloguev1.Item
		err := copier.Copy(&it, &item)
		if err != nil {
			log.Fatalf("failed to copy %v", err)
			return nil, err
		}

		responseItems = append(responseItems, &it)
	}

	return &cataloguev1.ListItemsResponse{Items: responseItems}, nil
}

func (cs *catalogueService) GetItem(ctx context.Context, req *cataloguev1.GetItemRequest) (*cataloguev1.GetItemResponse, error) {
	id, err := strconv.Atoi(req.Id)
	if err != nil {
		return nil, err
	}

	item, err := cs.catalogue.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}

	itemResponse := &cataloguev1.Item{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Quantity:    item.Quantity,
	}
	return &cataloguev1.GetItemResponse{Item: itemResponse}, nil
}
