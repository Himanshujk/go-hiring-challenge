package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// mockProductRepo implements productRepository for testing.
type mockProductRepo struct {
	products []models.Product
	total    int64
	repoErr  error
	product  *models.Product
	notFound bool
}

func (m *mockProductRepo) GetProducts(_ context.Context, _ models.ProductFilter, _, _ int) ([]models.Product, int64, error) {
	return m.products, m.total, m.repoErr
}

func (m *mockProductRepo) GetProductByCode(_ context.Context, _ string) (*models.Product, error) {
	if m.notFound {
		return nil, models.ErrNotFound
	}
	return m.product, m.repoErr
}

var clothingCategory = models.Category{Code: "clothing", Name: "Clothing"}

var sampleProducts = []models.Product{
	{
		Code:     "PROD001",
		Price:    decimal.NewFromFloat(10.99),
		Category: clothingCategory,
		Variants: []models.Variant{
			{Name: "Small", SKU: "PROD001-S", Price: decimal.NewFromFloat(0)},
			{Name: "Large", SKU: "PROD001-L", Price: decimal.NewFromFloat(11.99)},
		},
	},
	{
		Code:     "PROD004",
		Price:    decimal.NewFromFloat(15.00),
		Category: clothingCategory,
		Variants: []models.Variant{},
	},
}

func TestHandleGet(t *testing.T) {
	t.Run("returns products with pagination defaults", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{products: sampleProducts, total: 2})

		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp listResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, int64(2), resp.Total)
		assert.Equal(t, 0, resp.Offset)
		assert.Equal(t, 10, resp.Limit)
		assert.Len(t, resp.Products, 2)
	})

	t.Run("custom offset and limit", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{products: sampleProducts, total: 2})

		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=5&limit=20", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp listResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, 5, resp.Offset)
		assert.Equal(t, 20, resp.Limit)
	})

	t.Run("limit clamped to max 100", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{products: sampleProducts, total: 2})

		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=999", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp listResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, 100, resp.Limit)
	})

	t.Run("limit clamped to min 1", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{products: sampleProducts, total: 2})

		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=0", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp listResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, 1, resp.Limit)
	})

	t.Run("invalid offset returns 400", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{})

		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=abc", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("negative offset returns 400", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{})

		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=-1", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid limit returns 400", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{})

		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=abc", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid price_lt returns 400", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{})

		req := httptest.NewRequest(http.MethodGet, "/catalog?price_lt=notanumber", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("filter by category passes through to repo", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{products: sampleProducts[:1], total: 1})

		req := httptest.NewRequest(http.MethodGet, "/catalog?category=clothing", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp listResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Len(t, resp.Products, 1)
		assert.Equal(t, "PROD001", resp.Products[0].Code)
	})

	t.Run("repository error returns 500", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{repoErr: errors.New("db error")})

		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("product response includes category", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{products: sampleProducts[:1], total: 1})

		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		var resp listResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "clothing", resp.Products[0].Category.Code)
		assert.Equal(t, "Clothing", resp.Products[0].Category.Name)
	})
}

func TestHandleGetByCode(t *testing.T) {
	t.Run("returns product details", func(t *testing.T) {
		t.Parallel()
		prod := sampleProducts[0]
		handler := NewCatalogHandler(&mockProductRepo{product: &prod})

		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		w := httptest.NewRecorder()
		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp productResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "PROD001", resp.Code)
		assert.Equal(t, 10.99, resp.Price)
		assert.Equal(t, "clothing", resp.Category.Code)
		assert.Len(t, resp.Variants, 2)
	})

	t.Run("variant inherits product price when price is zero", func(t *testing.T) {
		t.Parallel()
		prod := sampleProducts[0]
		handler := NewCatalogHandler(&mockProductRepo{product: &prod})

		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		w := httptest.NewRecorder()
		handler.HandleGetByCode(w, req)

		var resp productResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		// Small variant has zero price — should inherit product price 10.99
		assert.Equal(t, 10.99, resp.Variants[0].Price)
	})

	t.Run("variant uses its own price when set", func(t *testing.T) {
		t.Parallel()
		prod := sampleProducts[0]
		handler := NewCatalogHandler(&mockProductRepo{product: &prod})

		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		w := httptest.NewRecorder()
		handler.HandleGetByCode(w, req)

		var resp productResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		// Large variant has price 11.99
		assert.Equal(t, 11.99, resp.Variants[1].Price)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{notFound: true})

		req := httptest.NewRequest(http.MethodGet, "/catalog/UNKNOWN", nil)
		req.SetPathValue("code", "UNKNOWN")
		w := httptest.NewRecorder()
		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("repository error returns 500", func(t *testing.T) {
		t.Parallel()
		handler := NewCatalogHandler(&mockProductRepo{repoErr: errors.New("db error")})

		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		w := httptest.NewRecorder()
		handler.HandleGetByCode(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
