package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/provenance-io/provenance/x/name/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/gorilla/mux"
)

// RegisterRoutes defines routes that get registered by the main application.
func registerQueryRoutes(cliCtx client.Context, r *mux.Router) {
	// Register handler for resolving addresses for names.
	r.HandleFunc(
		fmt.Sprintf("/%s/{name}", types.StoreKey),
		resolveNameHandler(cliCtx),
	).Methods("GET")
	// Register handler for looking up names bound to an address.
	r.HandleFunc(
		fmt.Sprintf("/%s/{address}/names", types.StoreKey),
		lookupNamesHandler(cliCtx),
	).Methods("GET")
}

// Resolve the address for the given name.
func resolveNameHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := strings.ToLower(strings.TrimSpace(vars["name"]))
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/resolve/%s", types.StoreKey, name), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// Lookup names that resolve to the given address.
func lookupNamesHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		address := strings.ToLower(strings.TrimSpace(vars["address"]))
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/lookup/%s", types.StoreKey, address), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
