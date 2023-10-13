package exchange

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Transfer contains bank inputs and outputs indicating a transfer that needs to be made.
type Transfer struct {
	// Inputs are the inputs that make up this transfer.
	Inputs []banktypes.Input
	// Outputs are the outputs that make up this transfer.
	Outputs []banktypes.Output
}

// Settlement contains information on how a set of orders is to be settled.
type Settlement struct {
	// Transfers are all of the inputs and outputs needed to facilitate movement of assets and price.
	Transfers []*Transfer
	// FeeInputs are the inputs needed to facilitate payment of order fees.
	FeeInputs []banktypes.Input
	// FullyFilledOrders are all the orders that are fully filled in this settlement.
	// If there is an order that's being partially filled, it will not be included in this list.
	FullyFilledOrders []*FilledOrder
	// PartialOrderFilled is a partially filled order with amounts indicating how much was filled.
	// This is not included in FullyFilledOrders.
	PartialOrderFilled *FilledOrder
	// PartialOrderLeft is what's left of the partially filled order.
	PartialOrderLeft *Order
}

// BuildSettlement processes the provided orders, identifying how the provided orders can be settled.
func BuildSettlement(askOrders, bidOrders []*Order, sellerFeeRatioLookup func(denom string) (*FeeRatio, error)) (*Settlement, error) {
	if err := validateCanSettle(askOrders, bidOrders); err != nil {
		return nil, err
	}

	askOFs := newOrderFulfillments(askOrders)
	bidOFs := newOrderFulfillments(bidOrders)

	// Allocate the assets first.
	if err := allocateAssets(askOFs, bidOFs); err != nil {
		return nil, err
	}

	settlement := &Settlement{}
	// Identify any partial order and update its entry in the order fulfillments.
	if err := splitPartial(askOFs, bidOFs, settlement); err != nil {
		return nil, err
	}

	// Allocate the prices.
	if err := allocatePrice(askOFs, bidOFs); err != nil {
		return nil, err
	}

	// Set the fees in the fulfillments
	sellerFeeRatio, err := sellerFeeRatioLookup(askOFs[0].GetPrice().Denom)
	if err != nil {
		return nil, err
	}
	if err = setFeesToPay(askOFs, bidOFs, sellerFeeRatio); err != nil {
		return nil, err
	}

	// Make sure everything adds up.
	if err = validateFulfillments(askOFs, bidOFs); err != nil {
		return nil, err
	}

	// Create the transfers
	if err = buildTransfers(askOFs, bidOFs, settlement); err != nil {
		return nil, err
	}

	// Indicate what's been filled in full and partially.
	populateFilled(askOFs, bidOFs, settlement)

	return settlement, nil
}

// IndexedAddrAmts is a set of addresses and amounts.
type IndexedAddrAmts struct {
	// addrs are a list of all addresses that have amounts.
	addrs []string
	// amts are a list of the coin amounts for each address (by slice index).
	amts []sdk.Coins
	// indexes are the index value for each address.
	indexes map[string]int
}

func NewIndexedAddrAmts() *IndexedAddrAmts {
	return &IndexedAddrAmts{
		indexes: make(map[string]int),
	}
}

// Add adds the coins to the given address.
// Panics if a provided coin is invalid.
func (i *IndexedAddrAmts) Add(addr string, coins ...sdk.Coin) {
	for _, coin := range coins {
		if err := coin.Validate(); err != nil {
			panic(fmt.Errorf("cannot index and add invalid coin amount %q", coin))
		}
	}
	n, known := i.indexes[addr]
	if !known {
		n = len(i.addrs)
		i.indexes[addr] = n
		i.addrs = append(i.addrs, addr)
		i.amts = append(i.amts, sdk.NewCoins())
	}
	i.amts[n] = i.amts[n].Add(coins...)
}

// GetAsInputs returns all the entries as bank Inputs.
// Panics if this is nil, has no addrs, or has a negative coin amount.
func (i *IndexedAddrAmts) GetAsInputs() []banktypes.Input {
	if i == nil || len(i.addrs) == 0 {
		return nil
	}
	rv := make([]banktypes.Input, len(i.addrs))
	for n, addr := range i.addrs {
		if !i.amts[n].IsAllPositive() {
			panic(fmt.Errorf("invalid indexed amount %q for address %q: cannot be zero or negative", addr, i.amts[n]))
		}
		rv[n] = banktypes.Input{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// GetAsOutputs returns all the entries as bank Outputs.
// Panics if this is nil, has no addrs, or has a negative coin amount.
func (i *IndexedAddrAmts) GetAsOutputs() []banktypes.Output {
	if i == nil || len(i.addrs) == 0 {
		return nil
	}
	rv := make([]banktypes.Output, len(i.addrs))
	for n, addr := range i.addrs {
		if !i.amts[n].IsAllPositive() {
			panic(fmt.Errorf("invalid indexed amount %q for address %q: cannot be zero or negative", addr, i.amts[n]))
		}
		rv[n] = banktypes.Output{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// distribution indicates an address and an amount that will either go to, or come from that address.
type distribution struct {
	// Address is a bech32 address string
	Address string
	// Amount is the amount that will either go to, or come from that address.
	Amount sdkmath.Int
}

// orderFulfillment is used to figure out how an order should be fulfilled.
type orderFulfillment struct {
	// Order is the original order with all its information.
	Order *Order

	// AssetDists contains distribution info for this order's assets.
	AssetDists []*distribution
	// AssetDists contains distribution info for this order's price.
	PriceDists []*distribution

	// AssetsFilledAmt is the total amount of assets being fulfilled for the order.
	AssetsFilledAmt sdkmath.Int
	// AssetsUnfilledAmt is the amount of order assets that have not yet been fulfilled for the order.
	AssetsUnfilledAmt sdkmath.Int
	// PriceAppliedAmt is the total price amount involved in this order fulfillment.
	// If this is a bid order, it's the actual amount the buyer will pay.
	// If this is an ask order it's the actual amount the seller will receive.
	PriceAppliedAmt sdkmath.Int
	// PriceLeftAmt is the price that has not yet been fulfilled for the order.
	// This can be negative for ask orders that are being filled at a higher price than requested.
	PriceLeftAmt sdkmath.Int
	// FeesToPay is the amount of settlement fees the order owner should pay to settle this order.
	FeesToPay sdk.Coins
}

var _ OrderI = (*orderFulfillment)(nil)

// newOrderFulfillment creates a new orderFulfillment wrapping the provided order.
func newOrderFulfillment(order *Order) *orderFulfillment {
	return &orderFulfillment{
		Order:             order,
		AssetsFilledAmt:   sdkmath.ZeroInt(),
		AssetsUnfilledAmt: order.GetAssets().Amount,
		PriceAppliedAmt:   sdkmath.ZeroInt(),
		PriceLeftAmt:      order.GetPrice().Amount,
	}
}

// newOrderFulfillments creates a new orderFulfillment for each of the provided orders.
func newOrderFulfillments(orders []*Order) []*orderFulfillment {
	rv := make([]*orderFulfillment, len(orders))
	for i, o := range orders {
		rv[i] = newOrderFulfillment(o)
	}
	return rv
}

// AssetCoin returns a coin with the given amount and the same denom as this order's assets.
func (f orderFulfillment) AssetCoin(amt sdkmath.Int) sdk.Coin {
	return sdk.Coin{Denom: f.GetAssets().Denom, Amount: amt}
}

// PriceCoin returns a coin with the given amount and the same denom as this order's price.
func (f orderFulfillment) PriceCoin(amt sdkmath.Int) sdk.Coin {
	return sdk.Coin{Denom: f.GetPrice().Denom, Amount: amt}
}

// GetAssetsFilled gets the coin value of the assets that have been filled in this fulfillment.
func (f orderFulfillment) GetAssetsFilled() sdk.Coin {
	return f.AssetCoin(f.AssetsFilledAmt)
}

// GetAssetsUnfilled gets the coin value of the assets left to fill in this fulfillment.
func (f orderFulfillment) GetAssetsUnfilled() sdk.Coin {
	return f.AssetCoin(f.AssetsUnfilledAmt)
}

// GetPriceApplied gets the coin value of the price that has been filled in this fulfillment.
func (f orderFulfillment) GetPriceApplied() sdk.Coin {
	return f.PriceCoin(f.PriceAppliedAmt)
}

// GetPriceLeft gets the coin value of the price left to fill in this fulfillment.
func (f orderFulfillment) GetPriceLeft() sdk.Coin {
	return f.PriceCoin(f.PriceLeftAmt)
}

// GetOrderID gets this fulfillment's order's id.
func (f orderFulfillment) GetOrderID() uint64 {
	return f.Order.GetOrderID()
}

// IsAskOrder returns true if this is an ask order.
func (f orderFulfillment) IsAskOrder() bool {
	return f.Order.IsAskOrder()
}

// IsBidOrder returns true if this is an ask order.
func (f orderFulfillment) IsBidOrder() bool {
	return f.Order.IsBidOrder()
}

// GetMarketID gets this fulfillment's order's market id.
func (f orderFulfillment) GetMarketID() uint32 {
	return f.Order.GetMarketID()
}

// GetOwner gets this fulfillment's order's owner.
func (f orderFulfillment) GetOwner() string {
	return f.Order.GetOwner()
}

// GetAssets gets this fulfillment's order's assets.
func (f orderFulfillment) GetAssets() sdk.Coin {
	return f.Order.GetAssets()
}

// GetPrice gets this fulfillment's order's price.
func (f orderFulfillment) GetPrice() sdk.Coin {
	return f.Order.GetPrice()
}

// GetSettlementFees gets this fulfillment's order's settlement fees.
func (f orderFulfillment) GetSettlementFees() sdk.Coins {
	return f.Order.GetSettlementFees()
}

// PartialFillAllowed gets this fulfillment's order's AllowPartial flag.
func (f orderFulfillment) PartialFillAllowed() bool {
	return f.Order.PartialFillAllowed()
}

// GetExternalID gets this fulfillment's external id.
func (f orderFulfillment) GetExternalID() string {
	return f.Order.GetExternalID()
}

// GetOrderType gets this fulfillment's order's type string.
func (f orderFulfillment) GetOrderType() string {
	return f.Order.GetOrderType()
}

// GetOrderTypeByte gets this fulfillment's order's type byte.
func (f orderFulfillment) GetOrderTypeByte() byte {
	return f.Order.GetOrderTypeByte()
}

// GetHoldAmount gets this fulfillment's order's hold amount.
func (f orderFulfillment) GetHoldAmount() sdk.Coins {
	return f.Order.GetHoldAmount()
}

// DistributeAssets records the distribution of assets in the provided amount to/from the given order.
func (f *orderFulfillment) DistributeAssets(order OrderI, amount sdkmath.Int) error {
	if f.AssetsUnfilledAmt.LT(amount) {
		return fmt.Errorf("cannot fill %s order %d having assets left %q with %q from %s order %d: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetAssetsUnfilled(), f.AssetCoin(amount), order.GetOrderType(), order.GetOrderID())
	}

	f.AssetsUnfilledAmt = f.AssetsUnfilledAmt.Sub(amount)
	f.AssetsFilledAmt = f.AssetsFilledAmt.Add(amount)
	f.AssetDists = append(f.AssetDists, &distribution{
		Address: order.GetOwner(),
		Amount:  amount,
	})
	return nil
}

// distributeAssets records the distribution of assets in the provided amount to/from the given fulfillments.
func distributeAssets(of1, of2 *orderFulfillment, amount sdkmath.Int) error {
	errs := []error{
		of1.DistributeAssets(of2, amount),
		of2.DistributeAssets(of1, amount),
	}
	return errors.Join(errs...)
}

// DistributePrice records the distribution of price in the provided amount to/from the given order.
func (f *orderFulfillment) DistributePrice(order OrderI, amount sdkmath.Int) error {
	if f.PriceLeftAmt.LT(amount) && f.IsBidOrder() {
		return fmt.Errorf("cannot fill %s order %d having price left %q to %s order %d at a price of %q: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetPriceLeft(), order.GetOrderType(), order.GetOrderID(), f.PriceCoin(amount))
	}

	f.PriceLeftAmt = f.PriceLeftAmt.Sub(amount)
	f.PriceAppliedAmt = f.PriceAppliedAmt.Add(amount)
	f.PriceDists = append(f.PriceDists, &distribution{
		Address: order.GetOwner(),
		Amount:  amount,
	})
	return nil
}

// distributePrice records the distribution of price in the provided amount to/from the given fulfillments.
func distributePrice(of1, of2 *orderFulfillment, amount sdkmath.Int) error {
	errs := []error{
		of1.DistributePrice(of2, amount),
		of2.DistributePrice(of1, amount),
	}
	return errors.Join(errs...)
}

// SplitOrder splits this order on the amount of assets filled.
// This order fulfillment is updated to have the filled order, and the unfilled portion is returned.
func (f *orderFulfillment) SplitOrder() (*Order, error) {
	filled, unfilled, err := f.Order.Split(f.AssetsFilledAmt)
	if err != nil {
		return nil, err
	}

	f.Order = filled
	f.AssetsUnfilledAmt = sdkmath.ZeroInt()
	f.PriceLeftAmt = filled.GetPrice().Amount.Sub(f.PriceAppliedAmt)
	return unfilled, nil
}

// AsFilledOrder creates a FilledOrder from this order fulfillment.
func (f orderFulfillment) AsFilledOrder() *FilledOrder {
	return NewFilledOrder(f.Order, f.GetPriceApplied(), f.FeesToPay)
}

// sumAssetsAndPrice gets the sum of assets, and the sum of prices of the provided orders.
func sumAssetsAndPrice(orders []*Order) (assets sdk.Coins, price sdk.Coins) {
	for _, o := range orders {
		assets = assets.Add(o.GetAssets())
		price = price.Add(o.GetPrice())
	}
	return
}

// sumPriceLeft gets the sum of price left of the provided fulfillments.
func sumPriceLeft(fulfillments []*orderFulfillment) sdkmath.Int {
	rv := sdkmath.ZeroInt()
	for _, f := range fulfillments {
		rv = rv.Add(f.PriceLeftAmt)
	}
	return rv
}

// validateCanSettle does some superficial checking of the provided orders to make sure we can try to settle them.
func validateCanSettle(askOrders, bidOrders []*Order) error {
	var errs []error
	if len(askOrders) == 0 {
		errs = append(errs, errors.New("no ask orders provided"))
	}
	if len(bidOrders) == 0 {
		errs = append(errs, errors.New("no bid orders provided"))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	for i, askOrder := range askOrders {
		if !askOrder.IsAskOrder() {
			errs = append(errs, fmt.Errorf("%s order %d is not an ask order but is in the askOrders list at index %d",
				askOrder.GetOrderType(), askOrder.GetOrderID(), i))
		}
	}
	for i, bidOrder := range bidOrders {
		if !bidOrder.IsBidOrder() {
			errs = append(errs, fmt.Errorf("%s order %d is not a bid order but is in the bidOrders list at index %d",
				bidOrder.GetOrderType(), bidOrder.GetOrderID(), i))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	totalAssetsForSale, totalAskPrice := sumAssetsAndPrice(askOrders)
	totalAssetsToBuy, totalBidPrice := sumAssetsAndPrice(bidOrders)

	if len(totalAssetsForSale) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple ask order asset denoms %q", totalAssetsForSale))
	}
	if len(totalAskPrice) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple ask order price denoms %q", totalAskPrice))
	}
	if len(totalAssetsToBuy) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple bid order asset denoms %q", totalAssetsToBuy))
	}
	if len(totalBidPrice) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple bid order price denoms %q", totalBidPrice))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if totalAssetsForSale[0].Denom != totalAssetsToBuy[0].Denom {
		errs = append(errs, fmt.Errorf("cannot settle different ask %q and bid %q asset denoms",
			totalAssetsForSale, totalAssetsToBuy))
	}
	if totalAskPrice[0].Denom != totalBidPrice[0].Denom {
		errs = append(errs, fmt.Errorf("cannot settle different ask %q and bid %q price denoms",
			totalAskPrice, totalBidPrice))
	}

	// Note: We don't compare the total asset and price amounts because we need to know what's
	// being partially filled before we can make an assertions about the amounts.

	return errors.Join(errs...)
}

// allocateAssets distributes the assets among the fulfillments.
func allocateAssets(askOFs, bidOFs []*orderFulfillment) error {
	a, b := 0, 0
	for a < len(askOFs) && b < len(bidOFs) {
		assetsFilledAmt, err := getFulfillmentAssetsAmt(askOFs[a], bidOFs[b])
		if err != nil {
			return err
		}
		if err = distributeAssets(askOFs[a], bidOFs[b], assetsFilledAmt); err != nil {
			return err
		}

		askFilled := askOFs[a].AssetsUnfilledAmt.IsZero()
		bidFilled := bidOFs[b].AssetsUnfilledAmt.IsZero()
		if !askFilled && !bidFilled {
			return fmt.Errorf("neither %s order %d nor %s order %d could have assets filled in full",
				askOFs[a].GetOrderType(), askOFs[a].GetOrderID(), bidOFs[b].GetOrderType(), bidOFs[b].GetOrderID())
		}
		if askFilled {
			a++
		}
		if bidFilled {
			b++
		}
	}

	return nil
}

// getFulfillmentAssetsAmt figures out the assets that can be fulfilled with the two provided orders.
func getFulfillmentAssetsAmt(of1, of2 *orderFulfillment) (sdkmath.Int, error) {
	if !of1.AssetsUnfilledAmt.IsPositive() || !of2.AssetsUnfilledAmt.IsPositive() {
		return sdkmath.ZeroInt(), fmt.Errorf("cannot fill %s order %d having assets left %q "+
			"with %s order %d having assets left %q: zero or negative assets left",
			of1.GetOrderType(), of1.GetOrderID(), of1.GetAssetsUnfilled(),
			of2.GetOrderType(), of2.GetOrderID(), of2.GetAssetsUnfilled())
	}

	return MinSDKInt(of1.AssetsUnfilledAmt, of2.AssetsUnfilledAmt), nil
}

// splitPartial checks the provided fulfillments for a partial order and splits it out, updating the applicable fulfillment.
// This will possibly populate the PartialOrderLeft in the provided Settlement.
func splitPartial(askOFs, bidOFs []*orderFulfillment, settlement *Settlement) error {
	if err := splitOrderFulfillments(askOFs, settlement); err != nil {
		return err
	}
	return splitOrderFulfillments(bidOFs, settlement)
}

// splitOrderFulfillments checks each of the orderFulfillment for partial (or incomplete) fills.
// If an appropriate partial fill is found, its orderFulfillment is update and Settlement.PartialOrderLeft is set.
func splitOrderFulfillments(fulfillments []*orderFulfillment, settlement *Settlement) error {
	lastI := len(fulfillments) - 1
	for i, f := range fulfillments {
		if f.AssetsFilledAmt.IsZero() {
			return fmt.Errorf("%s order %d (at index %d) has no assets filled",
				f.GetOrderType(), f.GetOrderID(), i)
		}
		if !f.AssetsUnfilledAmt.IsZero() {
			if i != lastI {
				return fmt.Errorf("%s order %d (at index %d) is not filled in full and is not the last %s order provided",
					f.GetOrderType(), f.GetOrderID(), i, f.GetOrderType())
			}
			if settlement.PartialOrderLeft != nil {
				return fmt.Errorf("%s order %d and %s order %d cannot both be partially filled",
					settlement.PartialOrderLeft.GetOrderType(), settlement.PartialOrderLeft.GetOrderID(),
					f.GetOrderType(), f.GetOrderID())
			}
			var err error
			settlement.PartialOrderLeft, err = f.SplitOrder()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// allocatePrice distributes the prices among the fulfillments.
func allocatePrice(askOFs, bidOFs []*orderFulfillment) error {
	// Check that the total ask price is not more than the total bid price.
	totalAskPriceAmt := sumPriceLeft(askOFs)
	totalBidPriceAmt := sumPriceLeft(bidOFs)
	if totalAskPriceAmt.GT(totalBidPriceAmt) {
		return fmt.Errorf("total ask price %q is greater than total bid price %q",
			askOFs[0].PriceCoin(totalAskPriceAmt), bidOFs[0].PriceCoin(totalBidPriceAmt))
	}

	// First pass at price distribution: Give all the asks their price.
	b := 0
	totalFilledFirstPass := sdkmath.ZeroInt()
	for _, askOF := range askOFs {
		for askOF.PriceLeftAmt.IsPositive() && bidOFs[b].PriceLeftAmt.IsPositive() {
			priceFilledAmt, err := getFulfillmentPriceAmt(askOF, bidOFs[b])
			if err != nil {
				return err
			}
			if err = distributePrice(askOF, bidOFs[b], priceFilledAmt); err != nil {
				return err
			}
			totalFilledFirstPass = totalFilledFirstPass.Add(priceFilledAmt)
			if !bidOFs[b].PriceLeftAmt.IsPositive() {
				b++
			}
		}
	}

	// Above, we made sure that ask price <= bid price.
	// So here, we can assume that total price filled = ask price.
	// If it's also the total bid price, we're done!
	if totalFilledFirstPass.Equal(totalBidPriceAmt) {
		return nil
	}

	// We also know that total price filled <= bid price, so bid price - total price left will be positive.
	totalLeftoverPriceAmt := totalBidPriceAmt.Sub(totalFilledFirstPass)
	// We need an assets total so we can allocate the leftovers by assets amounts.
	totalAssetsAmt := sdkmath.ZeroInt()
	for _, askOF := range askOFs {
		totalAssetsAmt = totalAssetsAmt.Add(askOF.AssetsFilledAmt)
	}

	// Now, distribute the leftovers per asset (truncated).
	// Start over on the asks, but start where we left off on the bids (since that's the first with any price left).
	a := -1
	firstPass := true
	leftoverPriceAmt := totalLeftoverPriceAmt
	for !leftoverPriceAmt.IsZero() {
		// Move on to the next ask, looping back to the first if we need to.
		a++
		if a == len(askOFs) {
			a = 0
			firstPass = false
		}

		// If there's no bids left to get this from, something's logically wrong with this process.
		if b >= len(bidOFs) {
			panic(fmt.Errorf("total ask price %q, total bid price %q, difference %q, left to allocate %q: "+
				"no bid orders left to allocate leftovers from",
				askOFs[0].PriceCoin(totalAskPriceAmt), bidOFs[0].PriceCoin(totalBidPriceAmt),
				askOFs[0].PriceCoin(totalLeftoverPriceAmt), askOFs[0].PriceCoin(leftoverPriceAmt)))
		}

		// Figure out how much additional price this order should get.
		addPriceAmt := totalLeftoverPriceAmt.Mul(askOFs[a].AssetsFilledAmt).Quo(totalAssetsAmt)
		if addPriceAmt.IsZero() {
			if firstPass {
				continue
			}
			// If this isn't the first time through the asks, distribute at least one price to each ask.
			addPriceAmt = sdkmath.OneInt()
		}
		// We can only add what's actually left over.
		addPriceAmt = MinSDKInt(addPriceAmt, leftoverPriceAmt)

		// If it can't all come out of the current bid, use the rest of what this bid has and move to the next bid.
		for !addPriceAmt.IsZero() && b < len(bidOFs) && bidOFs[b].PriceLeftAmt.LTE(addPriceAmt) {
			bidPriceLeft := bidOFs[b].PriceLeftAmt
			if err := distributePrice(askOFs[a], bidOFs[b], bidPriceLeft); err != nil {
				return err
			}
			addPriceAmt = addPriceAmt.Sub(bidPriceLeft)
			leftoverPriceAmt = leftoverPriceAmt.Sub(bidPriceLeft)
			b++
		}

		// If there's still additional price left, it can all come out of the current bid.
		if !addPriceAmt.IsZero() && b < len(bidOFs) {
			if err := distributePrice(askOFs[a], bidOFs[b], addPriceAmt); err != nil {
				return err
			}
			leftoverPriceAmt = leftoverPriceAmt.Sub(addPriceAmt)
			// If that was all of it, move on to the next.
			if bidOFs[b].PriceLeftAmt.IsZero() {
				b++
			}
		}
	}

	return nil
}

// getFulfillmentPriceAmt figures out the price that can be fulfilled with the two provided orders.
func getFulfillmentPriceAmt(of1, of2 *orderFulfillment) (sdkmath.Int, error) {
	if !of1.PriceLeftAmt.IsPositive() || !of2.PriceLeftAmt.IsPositive() {
		return sdkmath.ZeroInt(), fmt.Errorf("cannot fill %s order %d having price left %q "+
			"with %s order %d having price left %q: zero or negative price left",
			of1.GetOrderType(), of1.GetOrderID(), of1.GetPriceLeft(),
			of2.GetOrderType(), of2.GetOrderID(), of2.GetPriceLeft())
	}

	return MinSDKInt(of1.PriceLeftAmt, of2.PriceLeftAmt), nil
}

// setFeesToPay sets the FeesToPay on each fulfillment.
func setFeesToPay(askOFs, bidOFs []*orderFulfillment, sellerFeeRatio *FeeRatio) error {
	var errs []error
	for _, askOF := range askOFs {
		feesToPay := askOF.GetSettlementFees()
		if sellerFeeRatio != nil {
			fee, err := sellerFeeRatio.ApplyToLoosely(askOF.GetPriceApplied())
			if err != nil {
				errs = append(errs, fmt.Errorf("failed calculate ratio fee for %s order %d: %w",
					askOF.GetOrderType(), askOF.GetOrderID(), err))
				continue
			}
			feesToPay = feesToPay.Add(fee)
		}
		askOF.FeesToPay = feesToPay
	}

	for _, bidOF := range bidOFs {
		bidOF.FeesToPay = bidOF.GetSettlementFees()
	}

	return errors.Join(errs...)
}

// validateFulfillments runs .Validate on each fulfillment, returning any problems.
func validateFulfillments(askOFs, bidOFs []*orderFulfillment) error {
	var errs []error
	for _, askOF := range askOFs {
		if err := askOF.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	for _, bidOF := range bidOFs {
		if err := bidOF.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Validate makes sure the assets filled and price applied are acceptable for this fulfillment.
func (f orderFulfillment) Validate() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	orderPrice := f.GetPrice()
	switch {
	case f.IsAskOrder():
		if orderPrice.Amount.GT(f.PriceAppliedAmt) {
			return fmt.Errorf("%s order %d price %q is more than price filled %q",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceApplied())
		}
	case f.IsBidOrder():
		if !orderPrice.Amount.Equal(f.PriceAppliedAmt) {
			return fmt.Errorf("%s order %d price %q is not equal to price filled %q",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceApplied())
		}
	default:
		// This is here in case something new implements SubOrderI but a case isn't added here.
		panic(fmt.Errorf("%s order %d: unknown order type", f.GetOrderType(), f.GetOrderID()))
	}

	orderAssets := f.GetAssets()
	if !orderAssets.Amount.Equal(f.AssetsFilledAmt) {
		return fmt.Errorf("%s order %d assets %q does not equal filled assets %q",
			f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled())
	}

	return nil
}

// buildTransfers creates the transfers, inputs for fee payments,
// and fee total and sets those fields in the provided Settlement.
// This will populate the Transfers and FeeInputs fields in the provided Settlement.
func buildTransfers(askOFs, bidOFs []*orderFulfillment, settlement *Settlement) error {
	var errs []error
	indexedFees := NewIndexedAddrAmts()
	transfers := make([]*Transfer, 0, len(askOFs)+len(bidOFs))

	record := func(of *orderFulfillment, getter func(fulfillment *orderFulfillment) (*Transfer, error)) {
		assetTrans, err := getter(of)
		if err != nil {
			errs = append(errs, err)
		} else {
			transfers = append(transfers, assetTrans)
		}

		if !of.FeesToPay.IsZero() {
			if of.FeesToPay.IsAnyNegative() {
				errs = append(errs, fmt.Errorf("%s order %d cannot pay %q in fees: negative amount",
					of.GetOrderType(), of.GetOrderID(), of.FeesToPay))
			} else {
				indexedFees.Add(of.GetOwner(), of.FeesToPay...)
			}
		}
	}

	// If we got both the asset and price transfers from all OrderFulfillments, we'd be doubling
	// up on what's being traded. So we need to only get each from only one list.
	// Since we need to loop through both lists to make note of the fees, though, I thought it
	// would be good to get one from one list, and the other from the other. So I decided to get
	// the asset transfers from the asks, and price transfers from the bids so that each transfer
	// will have one input and multiple (or one) outputs. I'm not sure that it matters too much
	// where we get each transfer type from, though.

	for _, of := range askOFs {
		record(of, getAssetTransfer)
	}

	for _, of := range bidOFs {
		record(of, getPriceTransfer)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	settlement.Transfers = transfers
	settlement.FeeInputs = indexedFees.GetAsInputs()

	return nil
}

// populateFilled creates all the FilledOrder entries and stores them in the provided Settlement.
// This will populate the FullyFilledOrders and PartialOrderFilled fields in the provided Settlement.
func populateFilled(askOFs, bidOFs []*orderFulfillment, settlement *Settlement) {
	settlement.FullyFilledOrders = make([]*FilledOrder, 0, len(askOFs)+len(bidOFs))

	for _, f := range askOFs {
		if settlement.PartialOrderLeft != nil && settlement.PartialOrderLeft.GetOrderID() == f.GetOrderID() {
			settlement.PartialOrderFilled = f.AsFilledOrder()
		} else {
			settlement.FullyFilledOrders = append(settlement.FullyFilledOrders, f.AsFilledOrder())
		}
	}

	for _, f := range bidOFs {
		if settlement.PartialOrderLeft != nil && settlement.PartialOrderLeft.GetOrderID() == f.GetOrderID() {
			settlement.PartialOrderFilled = f.AsFilledOrder()
		} else {
			settlement.FullyFilledOrders = append(settlement.FullyFilledOrders, f.AsFilledOrder())
		}
	}
}

// getAssetTransfer gets the inputs and outputs to facilitate the transfers of assets for this order fulfillment.
func getAssetTransfer(f *orderFulfillment) (*Transfer, error) {
	assetsFilled := f.GetAssetsFilled()
	if !assetsFilled.Amount.IsPositive() {
		return nil, fmt.Errorf("%s order %d cannot be filled with %q assets: amount not positive",
			f.GetOrderType(), f.GetOrderID(), assetsFilled)
	}

	indexedDists := NewIndexedAddrAmts()
	sumDists := sdkmath.ZeroInt()
	for _, dist := range f.AssetDists {
		if !dist.Amount.IsPositive() {
			return nil, fmt.Errorf("%s order %d cannot have %q assets in a transfer: amount not positive",
				f.GetOrderType(), f.GetOrderID(), f.AssetCoin(dist.Amount))
		}
		indexedDists.Add(dist.Address, f.AssetCoin(dist.Amount))
		sumDists = sumDists.Add(dist.Amount)
	}

	if !sumDists.Equal(assetsFilled.Amount) {
		return nil, fmt.Errorf("%s order %d assets filled %q does not equal assets distributed %q",
			f.GetOrderType(), f.GetOrderID(), assetsFilled, f.AssetCoin(sumDists))
	}

	if f.IsAskOrder() {
		return &Transfer{
			Inputs:  []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.NewCoins(assetsFilled)}},
			Outputs: indexedDists.GetAsOutputs(),
		}, nil
	}
	if f.IsBidOrder() {
		return &Transfer{
			Inputs:  indexedDists.GetAsInputs(),
			Outputs: []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.NewCoins(assetsFilled)}},
		}, nil
	}

	// This is here in case a new SubTypeI is made that isn't accounted for in here.
	panic(fmt.Errorf("%s order %d: unknown order type", f.GetOrderType(), f.GetOrderID()))
}

// getPriceTransfer gets the inputs and outputs to facilitate the transfers for the price of this order fulfillment.
func getPriceTransfer(f *orderFulfillment) (*Transfer, error) {
	priceApplied := f.GetPriceApplied()
	if !priceApplied.Amount.IsPositive() {
		return nil, fmt.Errorf("%s order %d cannot be filled at price %q: amount not positive",
			f.GetOrderType(), f.GetOrderID(), priceApplied)
	}

	indexedDists := NewIndexedAddrAmts()
	sumDists := sdkmath.ZeroInt()
	for _, dist := range f.PriceDists {
		if !dist.Amount.IsPositive() {
			return nil, fmt.Errorf("%s order %d cannot have price %q in a transfer: amount not positive",
				f.GetOrderType(), f.GetOrderID(), f.PriceCoin(dist.Amount))
		}
		indexedDists.Add(dist.Address, f.PriceCoin(dist.Amount))
		sumDists = sumDists.Add(dist.Amount)
	}

	if !sumDists.Equal(priceApplied.Amount) {
		return nil, fmt.Errorf("%s order %d price filled %q does not equal price distributed %q",
			f.GetOrderType(), f.GetOrderID(), priceApplied, f.PriceCoin(sumDists))
	}

	if f.IsAskOrder() {
		return &Transfer{
			Inputs:  indexedDists.GetAsInputs(),
			Outputs: []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.NewCoins(priceApplied)}},
		}, nil
	}
	if f.IsBidOrder() {
		return &Transfer{
			Inputs:  []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.NewCoins(priceApplied)}},
			Outputs: indexedDists.GetAsOutputs(),
		}, nil
	}

	// This is here in case a new SubTypeI is made that isn't accounted for in here.
	panic(fmt.Errorf("%s order %d: unknown order type", f.GetOrderType(), f.GetOrderID()))
}
