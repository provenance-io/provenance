package rest

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/provenance-io/provenance/x/attribute/types"
)

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	// Register handler adding account attributes
	r.HandleFunc(
		fmt.Sprintf("/%s/attributes", types.StoreKey),
		NewAddAttributeRequestHandlerFn(clientCtx),
	).Methods("POST")
	// Register handler removing account attributes
	r.HandleFunc(
		fmt.Sprintf("/%s/attributes", types.StoreKey),
		NewDeleteAttributeHandlerFn(clientCtx),
	).Methods("DELETE")
}

// Request type for setting account attributes.
type AddAccountAttributeRequest struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Name    string       `json:"name"`
	Value   string       `json:"value"`
	Type    string       `json:"type"`
	Account string       `json:"account"`
}

// The HTTP handler for setting account attributes.
func NewAddAttributeRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AddAccountAttributeRequest
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		owner, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		value, err := base64.StdEncoding.DecodeString(req.Value)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		at, err := types.AttributeTypeFromString(req.Type)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		msg := types.NewMsgAddAttributeRequest(
			req.Account,
			owner, // name must resolve to this address
			req.Name,
			at,
			value,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

// Request type for removing account attributes.
type DeleteAccountAttributeRequest struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Name    string       `json:"name"`
	Account string       `json:"account"`
}

// The HTTP handler for removing account attributes.
func NewDeleteAttributeHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DeleteAccountAttributeRequest
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		owner, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		msg := types.NewMsgDeleteAttributeRequest(
			req.Account,
			owner, // name must resolve to this address
			req.Name,
		)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
