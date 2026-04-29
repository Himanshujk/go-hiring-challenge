package models

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

func (r *CategoriesRepository) GetAllCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("fetching categories: %w", err)
	}
	return categories, nil
}

func (r *CategoriesRepository) CreateCategory(ctx context.Context, c *Category) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("creating category: %w", err)
	}
	return nil
}
