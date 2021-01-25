package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/provenance-io/provenance/x/marker/types"
	// metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

const (
	markerID = "id"
)

func RegisterQueryRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc(
		fmt.Sprintf("/%s/%s", types.QuerierRoute, types.QueryMarkers),
		queryListAllMarkersHandlerFn(cliCtx),
	).Methods("GET")
	r.HandleFunc(
		fmt.Sprintf("/%s/%s/{%s}", types.QuerierRoute, types.QueryHolders, markerID),
		queryListMarkerHoldersHandlerFn(cliCtx),
	).Methods("GET")
	r.HandleFunc(
		fmt.Sprintf("/%s/%s/{%s}", types.QuerierRoute, types.QueryMarker, markerID),
		queryMarkerHandlerFn(cliCtx, types.QueryMarker),
	).Methods("GET")
	r.HandleFunc(
		fmt.Sprintf("/%s/%s/{%s}", types.QuerierRoute, types.QueryMarkerAccess, markerID),
		queryMarkerHandlerFn(cliCtx, types.QueryMarkerAccess),
	).Methods("GET")
	r.HandleFunc(
		fmt.Sprintf("/%s/%s/{%s}", types.QuerierRoute, types.QueryMarkerAssets, markerID),
		queryMarkerAssetsHandlerFn(cliCtx),
	).Methods("GET")
	r.HandleFunc(
		fmt.Sprintf("/%s/%s/{%s}", types.QuerierRoute, types.QueryMarkerEscrow, markerID),
		queryMarkerHandlerFn(cliCtx, types.QueryMarkerEscrow),
	).Methods("GET")
	r.HandleFunc(
		fmt.Sprintf("/%s/%s/{%s}", types.QuerierRoute, types.QueryMarkerSupply, markerID),
		queryMarkerHandlerFn(cliCtx, types.QueryMarkerSupply),
	).Methods("GET")
}

func queryListAllMarkersHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// This shadow is expected, function will conditionally modify cliCtx as required.
		// nolint: govet
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		status := r.FormValue("status")
		if len(status) > 0 {
			_, err = types.MarkerStatusFromString(status)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest,
					fmt.Sprintf("%s expected one of 'proposed,finalized,active,cancelled,destroyed'.", err.Error()))
				return
			}
		}
		params := types.NewQueryMarkersParams(page, limit, "", status)
		bz, err := cliCtx.LegacyAmino.MarshalJSON(params)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryMarkers)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarkerHandlerFn(cliCtx client.Context, fn string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars[markerID]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.ModuleName, fn, id), nil)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarkerAssetsHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// _, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 200)
		_, _, _, err := rest.ParseHTTPArgsWithLimit(r, 200)
		if rest.CheckBadRequestError(w, err) {
			return
		}
		vars := mux.Vars(r)
		id := vars[markerID]

		var addr sdk.AccAddress
		var addrErr error

		if addr, err = sdk.AccAddressFromBech32(id); err != nil {
			if addr, addrErr = types.MarkerAddress(id); addrErr != nil {
				rest.WriteErrorResponse(w, http.StatusNotFound,
					fmt.Errorf("invalid address: %s, %w", addrErr, err).Error())
				return
			}
		}

		// params := metadatatypes.OwnershipRequest{
		// 	Pagination: &query.PageRequest{
		// 		Limit:  uint64(limit),
		// 		Offset: uint64(page * limit),
		// 	},
		// 	Address: addr.String(),
		// }

		bz, err := []byte{}, fmt.Errorf("todo: import metadata module") //  cliCtx.LegacyAmino.MarshalJSON(params)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		// This shadow is expected, function will conditionally modify cliCtx as required.
		// nolint: govet
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// query is handled by the metadata module, ownership function.  Requires an address,
		// returns scope uuids linked to it.
		path := fmt.Sprintf("custom/%s/%s/%s", "metadata", "ownership", addr.String())
		res, _, err := cliCtx.QueryWithData(path, bz)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryListMarkerHoldersHandlerFn(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// This shadow is expected, function will conditionally modify cliCtx as required.
		// nolint: govet
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)

		params := types.NewQueryMarkersParams(page, limit, vars[markerID], "")
		bz, err := cliCtx.LegacyAmino.MarshalJSON(params)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryHolders)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
