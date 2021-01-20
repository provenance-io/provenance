package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/provenance-io/provenance/x/name/types"
)

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	// Register handler for binding names to addresses.
	r.HandleFunc(
		fmt.Sprintf("/%s", types.StoreKey),
		NewBindNameRequestHandlerFn(clientCtx),
	).Methods("POST")
}

// BindNameRequest type for binding names to addresses.
type BindNameRequest struct {
	BaseReq    rest.BaseReq `json:"base_req"`
	Name       string       `json:"name"`
	Address    string       `json:"address"`
	RootName   string       `json:"root"`
	Restricted bool         `json:"restricted"`
}

// NewBindNameRequestHandlerFn returns an HTTP handler for binding names to addresses.
func NewBindNameRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BindNameRequest
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		address, err := sdk.AccAddressFromBech32(req.Address)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}
		rootAddress, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}
		msg := types.NewMsgBindNameRequest(
			types.NameRecord{
				Name:       req.Name,
				Address:    address.String(),
				Restricted: req.Restricted,
			},
			types.NameRecord{
				Name:    req.RootName,
				Address: rootAddress.String(),
			},
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
