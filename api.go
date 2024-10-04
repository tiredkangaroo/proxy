package main

import (
	"fmt"

	"github.com/nikumar1206/puff"
	"github.com/nikumar1206/puff/middleware"
)

type ProxyRequestIDInPath struct {
	ID string `kind:"path"`
}

func startAPI() {
	app := puff.DefaultApp("dashboard")

	if env.DEBUG == true {
		app.Use(middleware.CORS())
	}

	api := puff.NewRouter("api", "/api")
	// Retrieve all proxy requests.
	api.Get("/proxy-requests", nil, func(c *puff.Context) {
		data, err := getProxyRequests()
		if err != nil {
			env.Logger.Error("API", "request-id", c.GetRequestID(), "error", err.Error())
			c.InternalServerError("[%s] an error occured while getting proxy requests", c.GetRequestID())
			return
		}
		c.SendResponse(puff.JSONResponse{
			Content: map[string]any{"error": nil, "data": data},
		})
	})

	// Delete all proxy requests.
	api.Delete("/proxy-requests", nil, func(c *puff.Context) {
		err := deleteAllProxyRequests()
		if err != nil {
			env.Logger.Error("API", "request-id", c.GetRequestID(), "error", err.Error())
			c.SendResponse(puff.JSONResponse{
				Content: map[string]any{
					"error": fmt.Sprintf("[%s] an error occured while deleting all proxy requests", c.GetRequestID()),
				},
			})
			return
		}
		c.SendResponse(puff.JSONResponse{
			Content: map[string]any{
				"error": nil,
			},
		})
	})

	pg := new(ProxyRequestIDInPath)
	// Retrieve a proxy request.
	api.Get("/proxy-requests/{id}", pg, func(c *puff.Context) {
		pr, err := getProxyRequestByID(pg.ID)
		if err != nil {
			env.Logger.Error("API", "request-id", c.GetRequestID(), "error", err.Error())
			c.InternalServerError("[%s] an error occured while getting proxy request", c.GetRequestID())
			return
		}
		c.SendResponse(puff.JSONResponse{
			Content: map[string]any{"error": nil, "data": pr},
		})
	})

	p := new(ProxyRequestIDInPath)
	// Delete a proxy request.
	api.Delete("/proxy-requests/{id}", p, func(c *puff.Context) {
		err := deleteProxyRequest(p.ID)
		if err != nil {
			env.Logger.Error("API", "request-id", c.GetRequestID(), "error", err.Error())
			c.SendResponse(puff.JSONResponse{
				Content: map[string]any{
					"error": fmt.Sprintf("[%s] an error occured while deleting a proxy requests", c.GetRequestID()),
				},
			})
			return
		}
		c.SendResponse(puff.JSONResponse{
			Content: map[string]any{
				"error": nil,
			},
		})
	})

	app.IncludeRouter(api)
	app.ListenAndServe(":1212")
}
