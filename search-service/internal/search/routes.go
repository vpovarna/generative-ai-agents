package search

import (
	"github.com/emicklei/go-restful/v3"
)

func RegisterRoutes(container *restful.Container, handler *SearchHandler) {
	ws := new(restful.WebService)
	ws.
		Path("/search/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// Semantic search endpoint
	ws.Route(ws.POST("/semantic").
		To(handler.SemanticSearch).
		Doc("Vector similarity search").
		Reads(SearchRequest{}).
		Writes(SearchResponse{}))

	// Keyword search endpoint
	ws.Route(ws.POST("/keyword").
		To(handler.KeywordSearch).
		Doc("Full-text keyword search").
		Reads(SearchRequest{}).
		Writes(SearchResponse{}))

	// Hybrid search endpoint
	ws.Route(ws.POST("/hybrid").
		To(handler.HybridSearch).
		Doc("Hybrid search with RRF").
		Reads(SearchRequest{}).
		Writes(SearchResponse{}))

	container.Add(ws)
}
