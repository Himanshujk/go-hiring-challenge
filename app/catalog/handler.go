package catalog

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type productRepository interface {
	GetProducts(ctx context.Context, filter models.ProductFilter, offset, limit int) ([]models.Product, int64, error)
	GetProductByCode(ctx context.Context, code string) (*models.Product, error)
}

type CatalogHandler struct {
	repo productRepository
}

func NewCatalogHandler(r productRepository) *CatalogHandler {
	return &CatalogHandler{repo: r}
}

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type variantResponse struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type productResponse struct {
	Code     string            `json:"code"`
	Price    float64           `json:"price"`
	Category categoryResponse  `json:"category"`
	Variants []variantResponse `json:"variants,omitempty"`
}

type listResponse struct {
	Products []productResponse `json:"products"`
	Total    int64             `json:"total"`
	Offset   int               `json:"offset"`
	Limit    int               `json:"limit"`
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	offset, limit, err := parsePagination(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := models.ProductFilter{
		CategoryCode: r.URL.Query().Get("category"),
	}

	if priceLt := r.URL.Query().Get("price_lt"); priceLt != "" {
		d, err := decimal.NewFromString(priceLt)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "invalid price_lt value")
			return
		}
		filter.PriceLessThan = &d
	}

	products, total, err := h.repo.GetProducts(r.Context(), filter, offset, limit)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.OKResponse(w, listResponse{
		Products: mapProducts(products),
		Total:    total,
		Offset:   offset,
		Limit:    limit,
	})
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	product, err := h.repo.GetProductByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "product not found")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.OKResponse(w, mapProduct(*product))
}

func parsePagination(r *http.Request) (offset, limit int, err error) {
	offset = 0
	limit = 10

	if v := r.URL.Query().Get("offset"); v != "" {
		parsed, parseErr := strconv.Atoi(v)
		if parseErr != nil || parsed < 0 {
			return 0, 0, errors.New("invalid offset value")
		}
		offset = parsed
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, parseErr := strconv.Atoi(v)
		if parseErr != nil {
			return 0, 0, errors.New("invalid limit value")
		}
		// clamp rather than reject — no reason to error on limit=200
		if parsed < 1 {
			parsed = 1
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	return offset, limit, nil
}

func mapProducts(products []models.Product) []productResponse {
	result := make([]productResponse, len(products))
	for i, p := range products {
		result[i] = mapProduct(p)
	}
	return result
}

func mapProduct(p models.Product) productResponse {
	variants := make([]variantResponse, len(p.Variants))
	for i, v := range p.Variants {
		price := p.Price
		// variant price is NULL in DB; GORM scans NULL as zero — zero means inherit from product
		if !v.Price.IsZero() {
			price = v.Price
		}
		variants[i] = variantResponse{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price.InexactFloat64(),
		}
	}
	return productResponse{
		Code:  p.Code,
		Price: p.Price.InexactFloat64(),
		Category: categoryResponse{
			Code: p.Category.Code,
			Name: p.Category.Name,
		},
		Variants: variants,
	}
}
