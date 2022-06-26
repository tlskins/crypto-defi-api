package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/my_projects/sol-arb-api/api"
)

func main() {
	var err error
	// params := map[string]string{
	// 	"offset":    "0",
	// 	"page_size": "24",
	// 	"ranking":   "price_asc",
	// 	"symbol":    "primates",
	// }
	out := map[string]interface{}{}
	outErr := map[string]interface{}{}
	client := api.NewHttpClient()
	if err = api.HttpRequest(client, "POST", "https://api.coralcube.io/v1/getItems?offset=0&page_size=24&ranking=price_asc&symbol=primates", nil, nil, &out, &outErr); err != nil {
		log.Fatal(err)
	}
	spew.Dump(out)

	log.Println("Finished...")
}
