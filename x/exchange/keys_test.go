package exchange

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetMarketAddress(t *testing.T) {
	// The goal here is to make sure that, a) they're significantly different
	// from each other, b) they never change, and c) they're 20 bytes long.
	// I got the expected values from running each entry and copying the result.
	// They should never change unexpectedly, though.

	// I'm defining them as hex strings that get converted to byte slices instead
	// of using a bech32 string because I want the actual bytes to be more visible.
	// Plus, a hex string is the most compact way to express them as bytes.
	addrFromHex := func(h string) sdk.AccAddress {
		rv, err := hex.DecodeString(h)
		require.NoError(t, err, "hex.DecodeString(%q)", h)
		return rv
	}

	tests := []struct {
		marketID uint32
		addr     sdk.AccAddress
	}{
		{marketID: 1, addr: addrFromHex("D18FB05DF1B728AFDD46597395A7633328D87A94")},
		{marketID: 2, addr: addrFromHex("ADAF1E5727865007ECD20745A56550E8DB291C3F")},
		{marketID: 3, addr: addrFromHex("71D44FD9168EAE5442A2706C7E6165CDEE19C60D")},
		{marketID: 10, addr: addrFromHex("BA24034E603CB283114FDDFF5A187B147CC3522F")},
		{marketID: 128, addr: addrFromHex("AD15421CCD165A4354C890114A1C00595FF4814C")},
		{marketID: 32_768, addr: addrFromHex("B77B8B74C236B3E9661590C5D783B905E871F105")},
		{marketID: 2_147_483_648, addr: addrFromHex("9B598419EBEC767EEEE35873D265E1FF18E98101")},
		{marketID: 4_294_967_295, addr: addrFromHex("9C1E5A70066142E072369976E9B94044BD10A943")},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("market id %d", tc.marketID), func(t *testing.T) {
			addr := GetMarketAddress(tc.marketID)
			assert.Equal(t, tc.addr, addr, "GetMarketAddress(%d)", tc.marketID)
		})
	}
}
