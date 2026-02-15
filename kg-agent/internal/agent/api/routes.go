package agent

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/agent"
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
			Writes(agent.HealthResponse{}).
			Returns(200, "OK", agent.HealthResponse{}))

	ws.
		Route(ws.POST("/query").
			To(handler.Query).
			Doc("Query Claude").
			Metadata(restfulspec.KeyOpenAPITags, []string{"query"}).
			Reads(agent.QueryRequest{}).
			Writes(agent.QueryResponse{}).
			Returns(200, "OK", agent.QueryResponse{}).
			Returns(400, "Bad Request", middleware.ErrorResponse{}).
			Returns(500, "Internal Server Error", middleware.ErrorResponse{}))

	ws.
		Route(ws.POST("/query/stream").
			To(handler.QueryStream).
			Doc("Stream Query Claude").
			Metadata(restfulspec.KeyOpenAPITags, []string{"query"}).
			Reads(agent.QueryRequest{}).
			Writes(agent.QueryResponse{}).
			Returns(200, "OK", agent.QueryResponse{}).
			Returns(400, "Bad Request", middleware.ErrorResponse{}).
			Returns(500, "Internal Server Error", middleware.ErrorResponse{}))

	container.Add(ws)
}
