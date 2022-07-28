package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/davecgh/go-spew/spew"

	"github.com/my_projects/sol-arb-api/api"
	h "github.com/my_projects/sol-arb-api/entrypoints"
	s "github.com/my_projects/sol-arb-api/store"
	t "github.com/my_projects/sol-arb-api/types"
)

func Handle(allowedOrigin string) api.LambdaHandlerFunc {
	return func(ctx context.Context, r *api.LambdaReq) (resp api.LambdaResp, err error) {
		resp = api.LambdaResp{}
		var store *s.Store
		store, err = h.InitStore()
		api.CheckError(http.StatusInternalServerError, err)
		defer store.Close()

		spew.Dump(r.Path, r.PathParameters)

		// search orders counts *** block must be before
		if r.HTTPMethod == "GET" && strings.Contains(r.Path, "/by-discord-id/") {
			discId := r.PathParameters["discordId"]
			log.Printf("Getting trackers by discord id %s\n", discId)
			var trackers []*t.TokenTracker
			trackers, err = store.GetTokenTrackersByDiscordId(discId)
			api.CheckError(http.StatusNotFound, err)
			resp, err = api.BuildLambdaResponse(trackers, allowedOrigin)
			return
		}

		return
	}
}

func main() {
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	r := api.Responder{OriginStr: allowedOrigin}
	lambda.Start(r.HandleRequest(Handle(allowedOrigin)))
}
