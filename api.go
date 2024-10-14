package main

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/nikumar1206/puff"
	"github.com/nikumar1206/puff/middleware"
)

// ProxyRequestIDInPath represents the fields for a request in which the
// ProxyRequest ID is in the path.
type ProxyRequestIDInPath struct {
	ID string `kind:"path" description:"ID of the ProxyRequest"`
}

// SiteRegexInPath represents the fields for a request in which a
// regex is in the path.
type SiteRegexInPath struct {
	R string `kind:"path" description:"Regex of the site"`
}

// SiteInPath represents the fields for a request in which the
// Site URL is in the path.
type SiteInPath struct {
	R string `kind:"path" description:"URL of the site"`
}

// startAPI starts the API server.
func startAPI() {
	app := puff.DefaultApp("dashboard")
	app.Use(middleware.PanicWithConfig(middleware.PanicConfig{
		Skip: func(*puff.Context) bool { return false },
		FormatErrorResponse: func(c puff.Context, _ any) puff.Response {
			return puff.JSONResponse{
				StatusCode: 500,
				Content: map[string]any{
					"error": fmt.Sprintf("[%s] an unknown internal server error occured", c.GetRequestID()),
				},
			}
		},
	}))

	// set puff app logger and proxy slog logger
	loggerConfig := puff.LoggerConfig{}
	if env.DEBUG == true {
		app.Use(middleware.CORS())
		loggerConfig.Colorize = true
		loggerConfig.Level = slog.LevelDebug
	}
	logger := puff.NewLogger(loggerConfig)
	app.Logger = logger
	env.Logger = logger

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

	// provides the ability to refresh the information the proxy
	// works on. it does not provide an error in the case a refresh
	// fails.
	api.Patch("/refresh", nil, func(ctx *puff.Context) {
		fetchBlockedSites()
	})

	thisBlockedSite := new(SiteInPath)
	// checks if the site provided in the path is blocked.
	api.Get("/blockedsites/{r}", thisBlockedSite, func(ctx *puff.Context) {
		blocked := anyRegexMatch(env.BlockedSites, []byte(thisBlockedSite.R))
		ctx.SendResponse(puff.JSONResponse{
			Content: map[string]any{
				"error": nil,
				"data":  blocked,
			},
		})
	})

	// retrieves all blocked sites.
	api.Get("/blockedsites", nil, func(ctx *puff.Context) {
		sites := []string{}
		for _, site := range env.BlockedSites {
			sites = append(sites, site.String())
		}
		ctx.SendResponse(puff.JSONResponse{
			Content: map[string]any{
				"error": nil,
				"data":  sites,
			},
		})
	})

	blocksiteregex := new(SiteRegexInPath)
	// blocks all sites that match regex provided in path.
	api.Post("/blockedsites/{r}", blocksiteregex, func(ctx *puff.Context) {
		r, err := regexp.Compile(blocksiteregex.R)
		if err != nil {
			ctx.BadRequest("[%s] bad regex: %s", ctx.GetRequestID(), err.Error())
			return
		}
		err = saveBlockedSite(r)
		if err != nil {
			slog.Error("an error occured while attempting to save blocked site", "error", err.Error(), "request-id", ctx.GetRequestID())
			ctx.InternalServerError("[%s] an unknown error occured", ctx.GetRequestID())
			return
		}
		fetchBlockedSites()
	})

	blockedsiteregex := new(SiteRegexInPath)
	// unblocks all sites that match regex provided in path.
	api.Delete("/blockedsites/{r}", blockedsiteregex, func(ctx *puff.Context) {
		r, err := regexp.Compile(blockedsiteregex.R)
		if err != nil {
			ctx.BadRequest("[%s] bad regex: %s", ctx.GetRequestID(), err.Error())
			return
		}
		err = deleteBlockedSite(r)
		if err != nil {
			slog.Error("an error occured while attempting to delete blocked site", "error", err.Error(), "request-id", ctx.GetRequestID())
			ctx.InternalServerError("[%s] an unknown error occured", ctx.GetRequestID())
			return
		}
		ctx.SendResponse(puff.JSONResponse{
			Content: map[string]any{
				"error": nil,
			},
		})
		fetchBlockedSites()
	})

	app.IncludeRouter(api)
	app.ListenAndServe(":1212")
}
