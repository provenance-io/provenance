package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	// "github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/provenance-io/provenance/x/marker/types"
)

func RegisterTxRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc(
		"/marker/{denom}/mint",
		mintSupplyHandlerFn(clientCtx),
	).Methods("POST")
	r.HandleFunc(
		"/marker/{denom}/burn",
		burnSupplyHandlerFn(clientCtx),
	).Methods("POST")
	r.HandleFunc(
		"/marker/{denom}/status",
		updateStatusHandlerFn(clientCtx),
	).Methods("POST")
	r.HandleFunc(
		"/marker/{denom}/create",
		createMarkerHandlerFn(clientCtx),
	).Methods("POST")
	r.HandleFunc(
		"/marker/{denom}/withdraw",
		withdrawSupplyHandlerFn(clientCtx),
	).Methods("POST")
	r.HandleFunc(
		"/marker/{denom}/grant",
		grantAccessHandlerFn(clientCtx),
	).Methods("POST")
	r.HandleFunc(
		"/marker/{denom}/revoke",
		revokeAccessHandlerFn(clientCtx),
	).Methods("POST")
}

type (
	// NewMarkerRequest defines the basic properties to create a marker
	NewMarkerRequest struct {
		BaseReq    rest.BaseReq   `json:"base_req" yaml:"base_req"`
		Supply     sdk.Int        `json:"supply" yaml:"supply"`
		Manager    sdk.AccAddress `json:"manager" yaml:"manager"`
		MarkerType string         `json:"marker_type" yaml:"marker_type"`
	}

	// MarkerAccessRequest is used for grant/revoke permissions for a given address on a marker
	MarkerAccessRequest struct {
		BaseReq rest.BaseReq   `json:"base_req" yaml:"base_req"`
		Address sdk.AccAddress `json:"address" yaml:"address"`
		Grant   string         `json:"grant" yaml:"grant"`
	}

	// SupplyRequest defines the properties of a request to mint/burn supply for a marker
	SupplyRequest struct {
		BaseReq   rest.BaseReq   `json:"base_req" yaml:"base_req"`
		Recipient sdk.AccAddress `json:"recipient" yaml:"recipient"`
		Amount    sdk.Coin       `json:"amount" yaml:"amount"`
	}

	// StatusChangeRequest attempts to update the status of a marker
	StatusChangeRequest struct {
		BaseReq   rest.BaseReq `json:"base_req" yaml:"base_req"`
		NewStatus string       `json:"new_status" yaml:"new_status"`
	}
)

func mintSupplyHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SupplyRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		denom := mux.Vars(r)["denom"]
		err := sdk.ValidateDenom(denom)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if denom != req.Amount.Denom {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "denom to mint must match marker denom")
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		msg := types.NewMsgMintRequest(fromAddr, req.Amount)
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

func burnSupplyHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SupplyRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		denom := mux.Vars(r)["denom"]
		err := sdk.ValidateDenom(denom)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if denom != req.Amount.Denom {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "denom to burn must match marker denom")
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgBurnRequest(fromAddr, req.Amount)
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

func updateStatusHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req StatusChangeRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		denom := mux.Vars(r)["denom"]
		err := sdk.ValidateDenom(denom)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		status, err := types.MarkerStatusFromString(req.NewStatus)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		var msg sdk.Msg
		switch status {
		case types.StatusActive:
			msg = types.NewMsgActivateRequest(denom, fromAddr)
		case types.StatusFinalized:
			msg = types.NewMsgFinalizeRequest(denom, fromAddr)
		case types.StatusCancelled:
			msg = types.NewMsgCancelRequest(denom, fromAddr)
		case types.StatusDestroyed:
			msg = types.NewMsgDeleteRequest(denom, fromAddr)
		default:
			rest.WriteErrorResponse(w, http.StatusBadRequest, "invalid status change request")
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

func createMarkerHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req NewMarkerRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		denom := mux.Vars(r)["denom"]
		err := sdk.ValidateDenom(denom)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if req.Manager.Empty() {
			req.Manager = fromAddr
		}

		typeValue, err := types.MarkerTypeFromString(req.MarkerType)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		msg := types.NewMsgAddMarkerRequest(denom, req.Supply, fromAddr, req.Manager, typeValue, false, false)
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

func withdrawSupplyHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SupplyRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		denom := mux.Vars(r)["denom"]
		err := sdk.ValidateDenom(denom)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		msg := types.NewMsgWithdrawRequest(fromAddr, req.Recipient, denom, sdk.NewCoins(req.Amount))
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

func grantAccessHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MarkerAccessRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		denom := mux.Vars(r)["denom"]
		err := sdk.ValidateDenom(denom)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		grant := types.NewAccessGrant(req.Address, types.AccessListByNames(req.Grant))

		msg := types.NewMsgAddAccessRequest(denom, fromAddr, *grant)
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

func revokeAccessHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MarkerAccessRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		denom := mux.Vars(r)["denom"]
		if err := sdk.ValidateDenom(denom); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		msg := types.NewDeleteAccessRequest(denom, fromAddr, req.Address)
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
