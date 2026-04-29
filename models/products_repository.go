package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// ProductFilter holds optional filters for product queries.
type ProductFilter struct {
	CategoryCode  string
	PriceLessThan *decimal.Decimal
}

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

// GetProducts returns a page of products matching the filter along with the total count.
func (r *ProductsRepository) GetProducts(ctx context.Context, filter ProductFilter, offset, limit int) ([]Product, int64, error) {
	query := r.db.WithContext(ctx).Model(&Product{})

	if filter.CategoryCode != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", filter.CategoryCode)
	}

	if filter.PriceLessThan != nil {
		query = query.Where("products.price < ?", filter.PriceLessThan)
	}

	// count before applying pagination so the caller knows the full result set size
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting products: %w", err)
	}

	var products []Product
	if err := query.Preload("Category").Preload("Variants").
		Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("fetching products: %w", err)
	}

	return products, total, nil
}

// GetProductByCode returns a single product by its code, with category and variants preloaded.
func (r *ProductsRepository) GetProductByCode(ctx context.Context, code string) (*Product, error) {
	var product Product
	err := r.db.WithContext(ctx).Preload("Category").Preload("Variants").
		Where("code = ?", code).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("fetching product by code: %w", err)
	}
	return &product, nil
}
