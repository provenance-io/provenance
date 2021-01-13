package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/provenance-io/provenance/x/attribute/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/gorilla/mux"
)

// RegisterRoutes defines routes that get registered by the main application.
func registerQueryRoutes(cliCtx client.Context, r *mux.Router) {
	// Register handler for getting all account attributes
	r.HandleFunc(
		fmt.Sprintf("/%s/{address}/attributes", types.StoreKey),
		getAllAccountAttributes(cliCtx),
	).Methods("GET")
	// Register handler for getting account attributes by name
	r.HandleFunc(
		fmt.Sprintf("/%s/{address}/attributes/{name}", types.StoreKey),
		getAccountAttributes(cliCtx),
	).Methods("GET")
	// Register handler for scanning account attributes by name suffix
	r.HandleFunc(
		fmt.Sprintf("/%s/{address}/scan/{suffix}", types.StoreKey),
		scanAccountAttributes(cliCtx),
	).Methods("GET")
}

// Get all account attributes.
func getAllAccountAttributes(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		addr := strings.TrimSpace(vars["address"])
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/attributes/%s", types.StoreKey, addr), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// Get account attributes by name.
func getAccountAttributes(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		addr := strings.TrimSpace(vars["address"])
		name := strings.TrimSpace(vars["name"])
		path := fmt.Sprintf("custom/%s/attribute/%s/%s", types.StoreKey, addr, name)
		res, _, err := cliCtx.QueryWithData(path, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// Scan for account attributes with the given name suffix.
func scanAccountAttributes(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		addr := strings.TrimSpace(vars["address"])
		suffix := strings.TrimSpace(vars["suffix"])
		path := fmt.Sprintf("custom/%s/scan/%s/%s", types.StoreKey, addr, suffix)
		res, _, err := cliCtx.QueryWithData(path, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
