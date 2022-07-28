package main

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/my_projects/sol-arb-api/api"
	h "github.com/my_projects/sol-arb-api/entrypoints"
	c "github.com/my_projects/sol-arb-api/services/token_tracker"
	s "github.com/my_projects/sol-arb-api/store"
)

func Handle(ctx context.Context, r *api.LambdaReq) (resp api.LambdaResp, err error) {
	resp = api.LambdaResp{}
	var store *s.Store
	store, err = h.InitStore()
	api.CheckError(http.StatusInternalServerError, err)
	defer store.Close()
	discClient := h.InitDiscordClient()
	client := api.NewHttpClient()
	err = c.CheckTokens(store, client, discClient, true, true)
	api.CheckError(http.StatusInternalServerError, err)

	resp, err = api.BuildLambdaResponse(
		map[string]interface{}{"message": "success"},
		os.Getenv("ALLOWED_ORIGIN"),
	)
	return
}

func main() {
	r := api.Responder{OriginStr: os.Getenv("ALLOWED_ORIGIN")}
	lambda.Start(r.HandleRequest(Handle))
}
