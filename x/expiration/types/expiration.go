package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewExpiration(
	moduleAssetID string,
	owner string,
	time time.Time,
	deposit sdk.Coin,
	message types.Any,
) *Expiration {
	return &Expiration{
		ModuleAssetId: moduleAssetID,
		Owner:         owner,
		Time:          time,
		Deposit:       deposit,
		Message:       message,
	}
}

// ValidateBasic basic format checking of the data
func (e *Expiration) ValidateBasic() error {
	now := time.Now()
	if strings.TrimSpace(e.ModuleAssetId) == "" {
		return ErrEmptyModuleAssetID
	}
	if err := validateAddress(e.ModuleAssetId); err != nil {
		return err
	}
	if strings.TrimSpace(e.Owner) == "" {
		return ErrEmptyOwnerAddress
	}
	if e.Time.Before(now) {
		return ErrTimeInPast
	}
	if !e.Deposit.IsValid() {
		return ErrInvalidDeposit
	}
	if err := e.validateMessage(); err != nil {
		return sdkerrors.New(
			ErrInvalidMessage.Codespace(),
			ErrInvalidMessage.ABCICode(),
			err.Error())
	}
	return nil
}

// validateMessage validates `expiration.Message` conforms to `sdk.Msg`
// and is registered and whitelisted with the InterfaceRegistry
func (e *Expiration) validateMessage() error {
	// this may occur in message decoding
	if e.Message.Value == nil {
		return fmt.Errorf("expecting non nil value to validate an Any message")
	}
	if e.Message.TypeUrl == "" {
		return fmt.Errorf("expecting message type URL to unpack Any")
	}
	// validate message is a whitelisted sdk.Msg during a BroadcastTx
	var msg sdk.Msg
	if err := ModuleCdc.UnpackAny(&e.Message, &msg); err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("failed to unpack Any message: %v", msg)
	}
	return msg.ValidateBasic()
}

var reDuration = regexp.MustCompile(`(^[1-9]\d{0,11})([ywdh])$`)

// ParseDuration parses a duration string
func ParseDuration(s string) (*time.Duration, error) {
	// FindStringSubmatch returns a slice of strings holding the text of the
	// leftmost match of the regular expression in s and the matches, if any.
	matches := reDuration.FindStringSubmatch(s)
	if len(matches) != 3 {
		return nil, ErrDurationValue
	}

	// parse digits
	digit, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return nil, err
	}

	period := matches[2]
	var duration time.Duration
	switch period {
	case "y":
		duration = time.Duration(digit) * 24 * 365 * time.Hour
	case "w":
		duration = time.Duration(digit) * 24 * 7 * time.Hour
	case "d":
		duration = time.Duration(digit) * 24 * time.Hour
	case "h":
		duration = time.Duration(digit) * time.Hour
	default:
		// as a sanity check in case the regex check above fails
		return nil, ErrDurationValue
	}

	return &duration, nil
}
