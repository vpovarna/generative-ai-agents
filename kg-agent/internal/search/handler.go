package search

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/rs/zerolog/log"
)

type SearchHandler struct {
	service *Service
}

func NewSearchHandler(search *Service) *SearchHandler {
	return &SearchHandler{
		service: search,
	}
}

// Query handles POST /api/v1/semantic
func (h *SearchHandler) SemanticSearch(req *restful.Request, resp *restful.Response) {
	var searchRequest SearchRequest

	if err := req.ReadEntity(&searchRequest); err != nil {
		resp.WriteError(http.StatusBadGateway, err)
		return
	}

	// Set default search request limit if it's not set by the user
	if searchRequest.Limit == 0 {
		searchRequest.Limit = 10
	}

	ctx := req.Request.Context()

	searchResults, err := h.service.SematicSearch(ctx, searchRequest.Query, searchRequest.Limit)
	if err != nil {
		log.Error().Err(err).Msg("Semantic Search failed")
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	response := SearchResponse{
		Query:  searchRequest.Query,
		Result: searchResults,
		Count:  len(searchResults),
		Method: "sematic",
	}

	resp.WriteEntity(response)
}

// Query handles POST /api/v1/keyword
func (h *SearchHandler) KeywordSearch(req *restful.Request, resp *restful.Response) {
	var searchReq SearchRequest
	if err := req.ReadEntity(&searchReq); err != nil {
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	if searchReq.Limit == 0 {
		searchReq.Limit = 10
	}

	ctx := req.Request.Context()
	results, err := h.service.KeywordSearch(ctx, searchReq.Query, searchReq.Limit)
	if err != nil {
		log.Error().Err(err).Msg("Keyword search failed")
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	response := SearchResponse{
		Query:  searchReq.Query,
		Result: results,
		Count:  len(results),
		Method: "keyword",
	}

	resp.WriteEntity(response)
}
