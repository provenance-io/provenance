package types

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"
)

func TestUtil(t *testing.T) {
	fmt.Println("--- Util StringToAny AnyToString Test ---")
	ir := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)
	fmt.Println("Registry and Codec created.")

	originalString := "direct check string"
	fmt.Printf("Attempting StringToAny for string: '%s' ...\n", originalString)
	anyMsg, err := StringToAny(originalString)
	if err != nil {
		t.Fatalf("StringToAny failed: %v", err)
	}

	fmt.Printf("Attempting util.AnyToString for TypeURL %s...\n", anyMsg.TypeUrl)
	strMsg, err := AnyToString(cdc, anyMsg)
	require.NoError(t, err, "AnyToString failed")
	require.Equal(t, originalString, strMsg, "Decoded string does not match original")

	fmt.Printf("Unpacking successful. Final string: %q\n", strMsg)
	fmt.Println("--- Registry Direct Check Test Passed ---")

}
