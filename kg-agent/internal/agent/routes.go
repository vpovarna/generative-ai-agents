package agent

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/middleware"
)

func RegisterRoutes(container *restful.Container, handler *Handler) {
	ws := new(restful.WebService)

	ws.
		Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// Health endpoint
	ws.
		Route(ws.GET("health").
			To(handler.Health).
			Doc("Health check").
			Metadata(restfulspec.KeyOpenAPITags, []string{"health"}).
			Writes(HealthResponse{}).
			Returns(200, "OK", HealthResponse{}))

	ws.
		Route(ws.POST("/query").
			To(handler.Query).
			Doc("Query Claude").
			Metadata(restfulspec.KeyOpenAPITags, []string{"query"}).
			Reads(QueryRequest{}).
			Writes(QueryResponse{}).
			Returns(200, "OK", QueryResponse{}).
			Returns(400, "Bad Request", middleware.ErrorResponse{}).
			Returns(500, "Internal Server Error", middleware.ErrorResponse{}))

	ws.
		Route(ws.POST("/query/stream").
			To(handler.QueryStream).
			Consumes(restful.MIME_JSON).
			Produces("text/event-stream").
			Doc("Stream Query Claude").
			Metadata(restfulspec.KeyOpenAPITags, []string{"query"}).
			Reads(QueryRequest{}).
			Returns(200, "OK", nil).
			Returns(400, "Bad Request", middleware.ErrorResponse{}).
			Returns(500, "Internal Server Error", middleware.ErrorResponse{}))

	// Admin: Clear cache endpoint
	ws.
		Route(ws.POST("/admin/cache/clear").
			To(handler.ClearCache).
			Doc("Clear search result cache").
			Metadata(restfulspec.KeyOpenAPITags, []string{"admin"}).
			Returns(200, "OK", map[string]string{}).
			Returns(500, "Internal Server Error", middleware.ErrorResponse{}))

	container.Add(ws)
}
