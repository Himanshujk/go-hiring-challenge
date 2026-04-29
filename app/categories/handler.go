package categories

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type categoriesRepository interface {
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	CreateCategory(ctx context.Context, c *models.Category) error
}

type CategoriesHandler struct {
	repo categoriesRepository
}

func NewCategoriesHandler(r categoriesRepository) *CategoriesHandler {
	return &CategoriesHandler{repo: r}
}

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type listResponse struct {
	Categories []categoryResponse `json:"categories"`
}

func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.GetAllCategories(r.Context())
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := listResponse{Categories: make([]categoryResponse, len(cats))}
	for i, c := range cats {
		resp.Categories[i] = categoryResponse{Code: c.Code, Name: c.Name}
	}

	api.OKResponse(w, resp)
}

func (h *CategoriesHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code and name are required")
		return
	}

	category := &models.Category{Code: req.Code, Name: req.Name}
	if err := h.repo.CreateCategory(r.Context(), category); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(categoryResponse{Code: category.Code, Name: category.Name}); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
}
