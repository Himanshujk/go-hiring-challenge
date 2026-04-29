package categories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

type mockCategoriesRepo struct {
	categories []models.Category
	repoErr    error
}

func (m *mockCategoriesRepo) GetAllCategories(_ context.Context) ([]models.Category, error) {
	return m.categories, m.repoErr
}

func (m *mockCategoriesRepo) CreateCategory(_ context.Context, c *models.Category) error {
	return m.repoErr
}

var sampleCategories = []models.Category{
	{Code: "clothing", Name: "Clothing"},
	{Code: "shoes", Name: "Shoes"},
	{Code: "accessories", Name: "Accessories"},
}

func TestHandleGet(t *testing.T) {
	t.Run("returns all categories", func(t *testing.T) {
		t.Parallel()
		handler := NewCategoriesHandler(&mockCategoriesRepo{categories: sampleCategories})

		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp listResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Len(t, resp.Categories, 3)
		assert.Equal(t, "clothing", resp.Categories[0].Code)
		assert.Equal(t, "Clothing", resp.Categories[0].Name)
	})

	t.Run("returns empty list when no categories", func(t *testing.T) {
		t.Parallel()
		handler := NewCategoriesHandler(&mockCategoriesRepo{categories: []models.Category{}})

		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp listResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Empty(t, resp.Categories)
	})

	t.Run("repository error returns 500", func(t *testing.T) {
		t.Parallel()
		handler := NewCategoriesHandler(&mockCategoriesRepo{repoErr: errors.New("db error")})

		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()
		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandlePost(t *testing.T) {
	t.Run("creates category and returns 201", func(t *testing.T) {
		t.Parallel()
		handler := NewCategoriesHandler(&mockCategoriesRepo{})

		body, _ := json.Marshal(map[string]string{"code": "electronics", "name": "Electronics"})
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp categoryResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "electronics", resp.Code)
		assert.Equal(t, "Electronics", resp.Name)
	})

	t.Run("missing code returns 400", func(t *testing.T) {
		t.Parallel()
		handler := NewCategoriesHandler(&mockCategoriesRepo{})

		body, _ := json.Marshal(map[string]string{"name": "Electronics"})
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing name returns 400", func(t *testing.T) {
		t.Parallel()
		handler := NewCategoriesHandler(&mockCategoriesRepo{})

		body, _ := json.Marshal(map[string]string{"code": "electronics"})
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		t.Parallel()
		handler := NewCategoriesHandler(&mockCategoriesRepo{})

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader([]byte("not json")))
		w := httptest.NewRecorder()
		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("repository error returns 500", func(t *testing.T) {
		t.Parallel()
		handler := NewCategoriesHandler(&mockCategoriesRepo{repoErr: errors.New("db error")})

		body, _ := json.Marshal(map[string]string{"code": "electronics", "name": "Electronics"})
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.HandlePost(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
