package handlers_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	. "github.com/provenance-io/provenance/internal/handlers"
)

// assertEqualFloats checks that two floats are equal to a given precision (number of digits past the decimal).
// For example, 3.141 and 3.142 are equal with a precision of 2 but not 3.
// Returns true if they're equal, false if not.
func assertEqualFloats(t *testing.T, expected, actual float32, precision uint16, msgAndArgs ...interface{}) bool {
	ffmt := "%." + strconv.Itoa(int(precision)) + "f"
	exp := fmt.Sprintf(ffmt, expected)
	act := fmt.Sprintf(ffmt, actual)
	return assert.Equal(t, exp, act, msgAndArgs...)
}

func TestRestrictionOptions_CalcMaxValPct(t *testing.T) {
	tests := []struct {
		name     string
		this     *RestrictionOptions // Defaults to DefaultRestrictionOptions.
		valCount int
		exp      float32
	}{
		{
			name:     "default options: zero validators",
			valCount: 0, // gets a division by zero, which it then thinks is larger than the max.
			exp:      DefaultMaxCapPercent,
		},
		{
			name:     "unlimited options: zero validators",
			this:     UnlimitedRestrictionOptions,
			valCount: 0, // gets a division by zero, which it then thinks is larger than the max.
			exp:      1.0,
		},
		{
			name: "custom options: zero validators",
			this: &RestrictionOptions{
				MaxConcentrationMultiple: 123.456,
				MaxBondedCapPercent:      12.98,
				MinBondedCapPercent:      3.4,
			},
			valCount: 0, // gets a division by zero, which it then thinks is larger than the max.
			exp:      12.98,
		},
		{
			name: "calc below min",
			this: &RestrictionOptions{
				MaxConcentrationMultiple: 10.0,
				MaxBondedCapPercent:      1.0,
				MinBondedCapPercent:      0.5,
			},
			valCount: 21, // 10.0 / 21 = 0.47619 = less than min.
			exp:      0.5,
		},
		{
			name: "calc above max",
			this: &RestrictionOptions{
				MaxConcentrationMultiple: 10.0,
				MaxBondedCapPercent:      0.5,
				MinBondedCapPercent:      0.01,
			},
			valCount: 19, // 10.0 / 19 = 0.52631578947368 = more than max.
			exp:      0.5,
		},
		{name: "unlimited one validator", this: UnlimitedRestrictionOptions, valCount: 1, exp: 1.0},
		{name: "unlimited 1000 validators", this: UnlimitedRestrictionOptions, valCount: 1000, exp: 1.0},
		{name: "4 validators", valCount: 4, exp: DefaultMaxCapPercent},   // 5.5 / 4 = 1.375 = more than max.
		{name: "5 validators", valCount: 5, exp: DefaultMaxCapPercent},   // 5.5 / 5 = 1.1 = more than max.
		{name: "10 validators", valCount: 10, exp: DefaultMaxCapPercent}, // 5.5 / 10 = .55 = more than max.
		{name: "15 validators", valCount: 15, exp: DefaultMaxCapPercent}, // 5.5 / 15 = .36667 = more than max.
		{ // Largest number of validators that use the max. 5.5 / 16 = 0.34375 = more than max.
			name: "16 validators", valCount: 16, exp: DefaultMaxCapPercent},
		{name: "17 validators", valCount: 17, exp: 0.323529411764705882},
		{name: "18 validators", valCount: 18, exp: 0.30555555555556},
		{name: "19 validators", valCount: 19, exp: 0.28947368421053},
		{name: "20 validators", valCount: 20, exp: 0.275},
		{name: "21 validators", valCount: 21, exp: 0.26190476190476},
		{name: "22 validators", valCount: 22, exp: 0.25},
		{name: "23 validators", valCount: 23, exp: 0.23913043478261},
		{name: "24 validators", valCount: 24, exp: 0.22916666666667},
		{name: "25 validators", valCount: 25, exp: 0.22},
		{name: "26 validators", valCount: 26, exp: 0.21153846153846},
		{name: "27 validators", valCount: 27, exp: 0.2037037037037},
		{name: "28 validators", valCount: 28, exp: 0.19642857142857},
		{name: "29 validators", valCount: 29, exp: 0.18965517241379},
		{name: "30 validators", valCount: 30, exp: 0.183333333333333333},
		{name: "31 validators", valCount: 31, exp: 0.17741935483871},
		{name: "32 validators", valCount: 32, exp: 0.171875},
		{name: "33 validators", valCount: 33, exp: 0.16666666666667},
		{name: "34 validators", valCount: 34, exp: 0.16176470588235},
		{name: "35 validators", valCount: 35, exp: 0.157142857142857143},
		{name: "36 validators", valCount: 36, exp: 0.15277777777778},
		{name: "37 validators", valCount: 37, exp: 0.14864864864865},
		{name: "38 validators", valCount: 38, exp: 0.14473684210526},
		{name: "39 validators", valCount: 39, exp: 0.14102564102564},
		{name: "40 validators", valCount: 40, exp: 0.1375},
		{name: "41 validators", valCount: 41, exp: 0.13414634146341},
		{name: "42 validators", valCount: 42, exp: 0.13095238095238},
		{name: "43 validators", valCount: 43, exp: 0.12790697674419},
		{name: "44 validators", valCount: 44, exp: 0.125},
		{name: "45 validators", valCount: 45, exp: 0.122222222222222222},
		{name: "46 validators", valCount: 46, exp: 0.1195652173913},
		{name: "47 validators", valCount: 47, exp: 0.11702127659574},
		{name: "48 validators", valCount: 48, exp: 0.11458333333333},
		{name: "49 validators", valCount: 49, exp: 0.11224489795918},
		{name: "50 validators", valCount: 50, exp: 0.11},
		{name: "51 validators", valCount: 51, exp: 0.1078431372549},
		{name: "52 validators", valCount: 52, exp: 0.10576923076923},
		{name: "53 validators", valCount: 53, exp: 0.10377358490566},
		{name: "54 validators", valCount: 54, exp: 0.10185185185185},
		{name: "55 validators", valCount: 55, exp: 0.1},
		{name: "56 validators", valCount: 56, exp: 0.09821428571429},
		{name: "57 validators", valCount: 57, exp: 0.09649122807018},
		{name: "58 validators", valCount: 58, exp: 0.0948275862069},
		{name: "59 validators", valCount: 59, exp: 0.09322033898305},
		{name: "60 validators", valCount: 60, exp: 0.091666666666666667},
		{name: "61 validators", valCount: 61, exp: 0.09016393442623},
		{name: "62 validators", valCount: 62, exp: 0.08870967741935},
		{name: "63 validators", valCount: 63, exp: 0.08730158730159},
		{name: "64 validators", valCount: 64, exp: 0.0859375},
		{name: "65 validators", valCount: 65, exp: 0.084615384615384615},
		{name: "66 validators", valCount: 66, exp: 0.08333333333333},
		{name: "67 validators", valCount: 67, exp: 0.08208955223881},
		{name: "68 validators", valCount: 68, exp: 0.08088235294118},
		{name: "69 validators", valCount: 69, exp: 0.07971014492754},
		{name: "70 validators", valCount: 70, exp: 0.078571428571428571},
		{name: "71 validators", valCount: 71, exp: 0.07746478873239},
		{name: "72 validators", valCount: 72, exp: 0.07638888888889},
		{name: "73 validators", valCount: 73, exp: 0.07534246575342},
		{name: "74 validators", valCount: 74, exp: 0.07432432432432},
		{name: "75 validators", valCount: 75, exp: 0.073333333333333333},
		{name: "76 validators", valCount: 76, exp: 0.07236842105263},
		{name: "77 validators", valCount: 77, exp: 0.07142857142857},
		{name: "78 validators", valCount: 78, exp: 0.07051282051282},
		{name: "79 validators", valCount: 79, exp: 0.06962025316456},
		{name: "80 validators", valCount: 80, exp: 0.06875},
		{name: "81 validators", valCount: 81, exp: 0.06790123456790},
		{name: "82 validators", valCount: 82, exp: 0.06707317073171},
		{name: "83 validators", valCount: 83, exp: 0.06626506024096},
		{name: "84 validators", valCount: 84, exp: 0.06547619047619},
		{name: "85 validators", valCount: 85, exp: 0.06470588235294},
		{name: "86 validators", valCount: 86, exp: 0.06395348837209},
		{name: "87 validators", valCount: 87, exp: 0.0632183908046},
		{name: "88 validators", valCount: 88, exp: 0.0625},
		{name: "89 validators", valCount: 89, exp: 0.06179775280899},
		{name: "90 validators", valCount: 90, exp: 0.061111111111111111},
		{ // When only looking at 2 digits, this is the largest number of validators that did not use the min.
			name: "91 validators", valCount: 91, exp: 0.060439560439560439},
		{ // When only looking at 2 digits, this is the smallest number of validators that used the min.
			name: "92 validators", valCount: 92, exp: 0.059782608695652173},
		{name: "93 validators", valCount: 93, exp: 0.05913978494624},
		{name: "94 validators", valCount: 94, exp: 0.05851063829787},
		{name: "95 validators", valCount: 95, exp: 0.05789473684211},
		{name: "96 validators", valCount: 96, exp: 0.057291666666666667},
		{name: "97 validators", valCount: 97, exp: 0.05670103092784},
		{name: "98 validators", valCount: 98, exp: 0.05612244897959},
		{name: "99 validators", valCount: 99, exp: 0.055555555555555556},
		{name: "100 validators", valCount: 100, exp: 0.055}, // Max number of validators allowed by consensus params.
		{name: "109 validators", valCount: 109, exp: 0.050458715596330275},
		{ // Smallest number of validators that use the min. 5.5 / 110 = 0.05 = the min.
			name: "110 validators", valCount: 110, exp: DefaultMinCapPercent},
		{name: "111 validators", valCount: 111, exp: DefaultMinCapPercent}, // 5.5 / 111 = 0.0495 = less than min.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.this == nil {
				tc.this = DefaultRestrictionOptions
			}

			// Repeat this 1000 times to check determinism.
			for i := range 1000 {
				var act float32
				testFunc := func() {
					act = tc.this.CalcMaxValPct(tc.valCount)
				}

				if assert.NotPanics(t, testFunc, "[%d]: %#v.CalcMaxValPct(%d)", i, tc.this, tc.valCount) {
					// Only the first 6 digits of the result are used, so we only need to make sure these
					// values are the same out to 6 digits. But to be safe, we'll do it to 10 digits.
					if assertEqualFloats(t, tc.exp, act, 10, "[%d]: %#v.CalcMaxValPct(%d)", i, tc.this, tc.valCount) {
						continue
					}
				}

				// It only gets here if one of those assertions failed.
				// If it was the first one, we give up on the rest, otherwise, keep going to see how often it fails.
				if i == 0 {
					break
				}
			}
		})
	}
}

func TestCalcMaxValBond(t *testing.T) {
	// newInt is a way to create an int from a string (for when it's bigger than an int64).
	newInt := func(amt string) sdkmath.Int {
		rv, ok := sdkmath.NewIntFromString(amt)
		require.True(t, ok, "NewIntFromString(%s)", amt)
		return rv
	}

	// bigTotal an amount of total nhash delegated at one point in time. It's about 14.3 billion hash.
	bigTotal := newInt("14340544074388063121")

	// oneBil = one Billion. Handy in these tests. Since there's 6 digits used from the percent,
	// using 1 billion leaves three extra zeros where we can detect floating point problems.
	oneBil := sdkmath.NewInt(1_000_000_000)

	// dCon = default concentration (but with a shorter name).
	// Doing lots of tests with this because that's the value most likely to be used in these calcs.
	dCon := float32(DefaultConcentrationMultiple)

	tests := []struct {
		name      string
		totalBond sdkmath.Int
		maxValPct float32
		exp       sdkmath.Int
	}{
		{
			name:      "difficult float 0.1",
			totalBond: oneBil,
			maxValPct: float32(1) + (float32(1) / 10) - 1, // should be 0.1, but impossible to represent as float.
			exp:       sdkmath.NewInt(100_000_000),
		},
		{
			name:      "difficult float 0.2",
			totalBond: oneBil,
			maxValPct: float32(1) / 5, // should be 0.2, but impossible to represent as float.
			exp:       sdkmath.NewInt(200_000_000),
		},
		{
			name:      "difficult float 0.3",
			totalBond: oneBil,
			maxValPct: float32(1)/10 + float32(2)/10, // should be 0.3, but impossible to represent as float.
			exp:       sdkmath.NewInt(300_000_000),
		},
		{
			name:      "difficult float 1/3",
			totalBond: oneBil,
			maxValPct: float32(1) / 3, // should be 0.3333..., but impossible to represent as float.
			exp:       sdkmath.NewInt(333_333_000),
		},
		{
			name:      "difficult float 1/3 + 1/2",
			totalBond: oneBil,
			maxValPct: float32(1)/3 + float32(1)/2, // should be 0.8333..., but impossible to represent as float.
			exp:       sdkmath.NewInt(833_333_000),
		},
		{
			name:      "difficult float 1/3 + 1/3",
			totalBond: oneBil,
			maxValPct: float32(1)/3 + float32(1)/3, // should be 0.666..., but impossible to represent as float.
			exp:       sdkmath.NewInt(666_666_000),
		},
		{
			name:      "difficult float 1/3 + 1/10",
			totalBond: oneBil,
			maxValPct: float32(1)/3 + float32(1)/10, // should be 0.34333..., but impossible to represent as float.
			exp:       sdkmath.NewInt(433_333_000),
		},
		{
			name:      "exactly six digits from percent are used",
			totalBond: oneBil,
			maxValPct: 0.12345678, // Should only use 0.123456 of it.
			exp:       sdkmath.NewInt(123_456_000),
		},
		{
			name:      "big total, min pct",
			totalBond: bigTotal,
			maxValPct: DefaultMinCapPercent,
			exp:       newInt("717027203719403156"),
		},
		{
			name:      "big total, max pct",
			totalBond: bigTotal,
			maxValPct: DefaultMaxCapPercent,
			exp:       newInt("4732379544548060829"),
		},
		{
			name:      "big total, 91 validators: raw pct from division",
			totalBond: bigTotal,
			maxValPct: 5.5 / 91, // = 0.06043956043956
			exp:       newInt("866728143311940146"),
		},
		{
			name:      "big total, 91 validators: pct truncated to 2 digits",
			totalBond: bigTotal,
			maxValPct: 0.06, // 5.5 / 91 = 0.0604395, => 0.06
			exp:       newInt("860432644463283787"),
		},
		{
			name:      "big total, 91 validators: pct truncated to 3 digits",
			totalBond: bigTotal,
			maxValPct: 0.060, // 5.5 / 91 = 0.0604395 => 0.060
			exp:       newInt("860432644463283787"),
		},
		{
			name:      "big total, 91 validators: pct truncated to 4 digits",
			totalBond: bigTotal,
			maxValPct: 0.0604, // 5.5 / 91 = 0.0604395 => 0.0604
			exp:       newInt("866168862093039012"),
		},
		{
			name:      "big total, 91 validators: pct truncated to 5 digits",
			totalBond: bigTotal,
			maxValPct: 0.06043, // 5.5 / 91 = 0.0604395 => 0.06043
			exp:       newInt("866599078415270654"),
		},
		{
			name:      "big total, 91 validators: pct truncated to 6 digits",
			totalBond: bigTotal,
			maxValPct: 0.060439, // 5.5 / 91 = 0.0604395 => 0.060439
			exp:       newInt("866728143311940146"),
		},
		{
			name:      "big total, 91 validators: pct truncated to 7 digits",
			totalBond: bigTotal,
			maxValPct: 0.0604395, // 5.5 / 91 = // = 0.0604395
			exp:       newInt("866728143311940146"),
		},
		{
			name:      "big total, 100 validators",
			totalBond: bigTotal,
			maxValPct: dCon / 100,
			exp:       newInt("788729924091343471"),
		},
		{name: "min pct", totalBond: oneBil, maxValPct: DefaultMinCapPercent, exp: sdkmath.NewInt(50_000_000)},
		{name: "max pct", totalBond: oneBil, maxValPct: DefaultMaxCapPercent, exp: sdkmath.NewInt(330_000_000)},
		{name: "17 validators", totalBond: oneBil, maxValPct: dCon / 17, exp: sdkmath.NewInt(323_529_000)},
		{name: "18 validators", totalBond: oneBil, maxValPct: dCon / 18, exp: sdkmath.NewInt(305_555_000)},
		{name: "19 validators", totalBond: oneBil, maxValPct: dCon / 19, exp: sdkmath.NewInt(289_473_000)},
		{name: "20 validators", totalBond: oneBil, maxValPct: dCon / 20, exp: sdkmath.NewInt(275_000_000)},
		{name: "21 validators", totalBond: oneBil, maxValPct: dCon / 21, exp: sdkmath.NewInt(261_904_000)},
		{name: "22 validators", totalBond: oneBil, maxValPct: dCon / 22, exp: sdkmath.NewInt(250_000_000)},
		{name: "23 validators", totalBond: oneBil, maxValPct: dCon / 23, exp: sdkmath.NewInt(239_130_000)},
		{name: "24 validators", totalBond: oneBil, maxValPct: dCon / 24, exp: sdkmath.NewInt(229_166_000)},
		{name: "25 validators", totalBond: oneBil, maxValPct: dCon / 25, exp: sdkmath.NewInt(220_000_000)},
		{name: "26 validators", totalBond: oneBil, maxValPct: dCon / 26, exp: sdkmath.NewInt(211_538_000)},
		{name: "27 validators", totalBond: oneBil, maxValPct: dCon / 27, exp: sdkmath.NewInt(203_703_000)},
		{name: "28 validators", totalBond: oneBil, maxValPct: dCon / 28, exp: sdkmath.NewInt(196_428_000)},
		{name: "29 validators", totalBond: oneBil, maxValPct: dCon / 29, exp: sdkmath.NewInt(189_655_000)},
		{name: "30 validators", totalBond: oneBil, maxValPct: dCon / 30, exp: sdkmath.NewInt(183_333_000)},
		{name: "31 validators", totalBond: oneBil, maxValPct: dCon / 31, exp: sdkmath.NewInt(177_419_000)},
		{name: "32 validators", totalBond: oneBil, maxValPct: dCon / 32, exp: sdkmath.NewInt(171_875_000)},
		{name: "33 validators", totalBond: oneBil, maxValPct: dCon / 33, exp: sdkmath.NewInt(166_666_000)},
		{name: "34 validators", totalBond: oneBil, maxValPct: dCon / 34, exp: sdkmath.NewInt(161_764_000)},
		{name: "35 validators", totalBond: oneBil, maxValPct: dCon / 35, exp: sdkmath.NewInt(157_142_000)},
		{name: "36 validators", totalBond: oneBil, maxValPct: dCon / 36, exp: sdkmath.NewInt(152_777_000)},
		{name: "37 validators", totalBond: oneBil, maxValPct: dCon / 37, exp: sdkmath.NewInt(148_648_000)},
		{name: "38 validators", totalBond: oneBil, maxValPct: dCon / 38, exp: sdkmath.NewInt(144_736_000)},
		{name: "39 validators", totalBond: oneBil, maxValPct: dCon / 39, exp: sdkmath.NewInt(141_025_000)},
		{name: "40 validators", totalBond: oneBil, maxValPct: dCon / 40, exp: sdkmath.NewInt(137_500_000)},
		{name: "41 validators", totalBond: oneBil, maxValPct: dCon / 41, exp: sdkmath.NewInt(134_146_000)},
		{name: "42 validators", totalBond: oneBil, maxValPct: dCon / 42, exp: sdkmath.NewInt(130_952_000)},
		{name: "43 validators", totalBond: oneBil, maxValPct: dCon / 43, exp: sdkmath.NewInt(127_906_000)},
		{name: "44 validators", totalBond: oneBil, maxValPct: dCon / 44, exp: sdkmath.NewInt(125_000_000)},
		{name: "45 validators", totalBond: oneBil, maxValPct: dCon / 45, exp: sdkmath.NewInt(122_222_000)},
		{name: "46 validators", totalBond: oneBil, maxValPct: dCon / 46, exp: sdkmath.NewInt(119_565_000)},
		{name: "47 validators", totalBond: oneBil, maxValPct: dCon / 47, exp: sdkmath.NewInt(117_021_000)},
		{name: "48 validators", totalBond: oneBil, maxValPct: dCon / 48, exp: sdkmath.NewInt(114_583_000)},
		{name: "49 validators", totalBond: oneBil, maxValPct: dCon / 49, exp: sdkmath.NewInt(112_244_000)},
		{name: "50 validators", totalBond: oneBil, maxValPct: dCon / 50, exp: sdkmath.NewInt(110_000_000)},
		{name: "51 validators", totalBond: oneBil, maxValPct: dCon / 51, exp: sdkmath.NewInt(107_843_000)},
		{name: "52 validators", totalBond: oneBil, maxValPct: dCon / 52, exp: sdkmath.NewInt(105_769_000)},
		{name: "53 validators", totalBond: oneBil, maxValPct: dCon / 53, exp: sdkmath.NewInt(103_773_000)},
		{name: "54 validators", totalBond: oneBil, maxValPct: dCon / 54, exp: sdkmath.NewInt(101_851_000)},
		{name: "55 validators", totalBond: oneBil, maxValPct: dCon / 55, exp: sdkmath.NewInt(100_000_000)},
		{name: "56 validators", totalBond: oneBil, maxValPct: dCon / 56, exp: sdkmath.NewInt(98_214_000)},
		{name: "57 validators", totalBond: oneBil, maxValPct: dCon / 57, exp: sdkmath.NewInt(96_491_000)},
		{name: "58 validators", totalBond: oneBil, maxValPct: dCon / 58, exp: sdkmath.NewInt(94_827_000)},
		{name: "59 validators", totalBond: oneBil, maxValPct: dCon / 59, exp: sdkmath.NewInt(93_220_000)},
		{name: "60 validators", totalBond: oneBil, maxValPct: dCon / 60, exp: sdkmath.NewInt(91_666_000)},
		{name: "61 validators", totalBond: oneBil, maxValPct: dCon / 61, exp: sdkmath.NewInt(90_163_000)},
		{name: "62 validators", totalBond: oneBil, maxValPct: dCon / 62, exp: sdkmath.NewInt(88_709_000)},
		{name: "63 validators", totalBond: oneBil, maxValPct: dCon / 63, exp: sdkmath.NewInt(87_301_000)},
		{name: "64 validators", totalBond: oneBil, maxValPct: dCon / 64, exp: sdkmath.NewInt(85_937_000)},
		{name: "65 validators", totalBond: oneBil, maxValPct: dCon / 65, exp: sdkmath.NewInt(84_615_000)},
		{name: "66 validators", totalBond: oneBil, maxValPct: dCon / 66, exp: sdkmath.NewInt(83_333_000)},
		{name: "67 validators", totalBond: oneBil, maxValPct: dCon / 67, exp: sdkmath.NewInt(82_089_000)},
		{name: "68 validators", totalBond: oneBil, maxValPct: dCon / 68, exp: sdkmath.NewInt(80_882_000)},
		{name: "69 validators", totalBond: oneBil, maxValPct: dCon / 69, exp: sdkmath.NewInt(79_710_000)},
		{name: "70 validators", totalBond: oneBil, maxValPct: dCon / 70, exp: sdkmath.NewInt(78_571_000)},
		{name: "71 validators", totalBond: oneBil, maxValPct: dCon / 71, exp: sdkmath.NewInt(77_464_000)},
		{name: "72 validators", totalBond: oneBil, maxValPct: dCon / 72, exp: sdkmath.NewInt(76_388_000)},
		{name: "73 validators", totalBond: oneBil, maxValPct: dCon / 73, exp: sdkmath.NewInt(75_342_000)},
		{name: "74 validators", totalBond: oneBil, maxValPct: dCon / 74, exp: sdkmath.NewInt(74_324_000)},
		{name: "75 validators", totalBond: oneBil, maxValPct: dCon / 75, exp: sdkmath.NewInt(73_333_000)},
		{name: "76 validators", totalBond: oneBil, maxValPct: dCon / 76, exp: sdkmath.NewInt(72_368_000)},
		{name: "77 validators", totalBond: oneBil, maxValPct: dCon / 77, exp: sdkmath.NewInt(71_428_000)},
		{name: "78 validators", totalBond: oneBil, maxValPct: dCon / 78, exp: sdkmath.NewInt(70_512_000)},
		{name: "79 validators", totalBond: oneBil, maxValPct: dCon / 79, exp: sdkmath.NewInt(69_620_000)},
		{name: "80 validators", totalBond: oneBil, maxValPct: dCon / 80, exp: sdkmath.NewInt(68_750_000)},
		{name: "81 validators", totalBond: oneBil, maxValPct: dCon / 81, exp: sdkmath.NewInt(67_901_000)},
		{name: "82 validators", totalBond: oneBil, maxValPct: dCon / 82, exp: sdkmath.NewInt(67_073_000)},
		{name: "83 validators", totalBond: oneBil, maxValPct: dCon / 83, exp: sdkmath.NewInt(66_265_000)},
		{name: "84 validators", totalBond: oneBil, maxValPct: dCon / 84, exp: sdkmath.NewInt(65_476_000)},
		{name: "85 validators", totalBond: oneBil, maxValPct: dCon / 85, exp: sdkmath.NewInt(64_705_000)},
		{name: "86 validators", totalBond: oneBil, maxValPct: dCon / 86, exp: sdkmath.NewInt(63_953_000)},
		{name: "87 validators", totalBond: oneBil, maxValPct: dCon / 87, exp: sdkmath.NewInt(63_218_000)},
		{name: "88 validators", totalBond: oneBil, maxValPct: dCon / 88, exp: sdkmath.NewInt(62_500_000)},
		{name: "89 validators", totalBond: oneBil, maxValPct: dCon / 89, exp: sdkmath.NewInt(61_797_000)},
		{name: "90 validators", totalBond: oneBil, maxValPct: dCon / 90, exp: sdkmath.NewInt(61_111_000)},
		{name: "91 validators", totalBond: oneBil, maxValPct: dCon / 91, exp: sdkmath.NewInt(60_439_000)},
		{name: "92 validators", totalBond: oneBil, maxValPct: dCon / 92, exp: sdkmath.NewInt(59_782_000)},
		{name: "93 validators", totalBond: oneBil, maxValPct: dCon / 93, exp: sdkmath.NewInt(59_139_000)},
		{name: "94 validators", totalBond: oneBil, maxValPct: dCon / 94, exp: sdkmath.NewInt(58_510_000)},
		{name: "95 validators", totalBond: oneBil, maxValPct: dCon / 95, exp: sdkmath.NewInt(57_894_000)},
		{name: "96 validators", totalBond: oneBil, maxValPct: dCon / 96, exp: sdkmath.NewInt(57_291_000)},
		{name: "97 validators", totalBond: oneBil, maxValPct: dCon / 97, exp: sdkmath.NewInt(56_701_000)},
		{name: "98 validators", totalBond: oneBil, maxValPct: dCon / 98, exp: sdkmath.NewInt(56_122_000)},
		{name: "99 validators", totalBond: oneBil, maxValPct: dCon / 99, exp: sdkmath.NewInt(55_555_000)},
		{name: "100 validators", totalBond: oneBil, maxValPct: dCon / 100, exp: sdkmath.NewInt(55_000_000)},
		{name: "109 validators", totalBond: oneBil, maxValPct: dCon / 109, exp: sdkmath.NewInt(50_458_000)},
		{name: "110 validators", totalBond: oneBil, maxValPct: dCon / 110, exp: sdkmath.NewInt(50_000_000)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Repeat this 1000 times to check determinism.
			for i := range 1000 {
				var act sdkmath.Int
				testFunc := func() {
					act = CalcMaxValBond(tc.totalBond, tc.maxValPct)
				}

				if assert.NotPanics(t, testFunc, "[%d]: CalcMaxValBond(%s, %.12f)", i, tc.totalBond, tc.maxValPct) {
					if assert.Equal(t, tc.exp.String(), act.String(), "[%d]: CalcMaxValBond(%s, %.12f) result", i, tc.totalBond, tc.maxValPct) {
						continue
					}
				}

				// It only gets here if one of those assertions failed.
				// If it was the first one, we give up on the rest, otherwise, keep going to see how often it fails.
				if i == 0 {
					break
				}
			}
		})
	}
}

func TestCalcsTogether(t *testing.T) {
	tests := []struct {
		name      string
		opts      *RestrictionOptions // Defaults to DefaultRestrictionOptions.
		valCount  int
		totalBond sdkmath.Int // Defaults to 10,000,000,000,000,000,000 = 10 billion hash as nhash.
		exp       sdkmath.Int
	}{
		{valCount: 4, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 5, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 6, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 7, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 8, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 9, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 10, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 11, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 12, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 13, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 14, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 15, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 16, exp: sdkmath.NewInt(3_300_000_000_000000000)},
		{valCount: 17, exp: sdkmath.NewInt(3_235_290_000_000000000)},
		{valCount: 18, exp: sdkmath.NewInt(3_055_550_000_000000000)},
		{valCount: 19, exp: sdkmath.NewInt(2_894_730_000_000000000)},
		{valCount: 20, exp: sdkmath.NewInt(2_750_000_000_000000000)},
		{valCount: 21, exp: sdkmath.NewInt(2_619_040_000_000000000)},
		{valCount: 22, exp: sdkmath.NewInt(2_500_000_000_000000000)},
		{valCount: 23, exp: sdkmath.NewInt(2_391_300_000_000000000)},
		{valCount: 24, exp: sdkmath.NewInt(2_291_660_000_000000000)},
		{valCount: 25, exp: sdkmath.NewInt(2_200_000_000_000000000)},
		{valCount: 26, exp: sdkmath.NewInt(2_115_380_000_000000000)},
		{valCount: 27, exp: sdkmath.NewInt(2_037_030_000_000000000)},
		{valCount: 28, exp: sdkmath.NewInt(1_964_280_000_000000000)},
		{valCount: 29, exp: sdkmath.NewInt(1_896_550_000_000000000)},
		{valCount: 30, exp: sdkmath.NewInt(1_833_330_000_000000000)},
		{valCount: 31, exp: sdkmath.NewInt(1_774_190_000_000000000)},
		{valCount: 32, exp: sdkmath.NewInt(1_718_750_000_000000000)},
		{valCount: 33, exp: sdkmath.NewInt(1_666_660_000_000000000)},
		{valCount: 34, exp: sdkmath.NewInt(1_617_640_000_000000000)},
		{valCount: 35, exp: sdkmath.NewInt(1_571_420_000_000000000)},
		{valCount: 36, exp: sdkmath.NewInt(1_527_770_000_000000000)},
		{valCount: 37, exp: sdkmath.NewInt(1_486_480_000_000000000)},
		{valCount: 38, exp: sdkmath.NewInt(1_447_360_000_000000000)},
		{valCount: 39, exp: sdkmath.NewInt(1_410_250_000_000000000)},
		{valCount: 40, exp: sdkmath.NewInt(1_375_000_000_000000000)},
		{valCount: 41, exp: sdkmath.NewInt(1_341_460_000_000000000)},
		{valCount: 42, exp: sdkmath.NewInt(1_309_520_000_000000000)},
		{valCount: 43, exp: sdkmath.NewInt(1_279_060_000_000000000)},
		{valCount: 44, exp: sdkmath.NewInt(1_250_000_000_000000000)},
		{valCount: 45, exp: sdkmath.NewInt(1_222_220_000_000000000)},
		{valCount: 46, exp: sdkmath.NewInt(1_195_650_000_000000000)},
		{valCount: 47, exp: sdkmath.NewInt(1_170_210_000_000000000)},
		{valCount: 48, exp: sdkmath.NewInt(1_145_830_000_000000000)},
		{valCount: 49, exp: sdkmath.NewInt(1_122_440_000_000000000)},
		{valCount: 50, exp: sdkmath.NewInt(1_100_000_000_000000000)},
		{valCount: 51, exp: sdkmath.NewInt(1_078_430_000_000000000)},
		{valCount: 52, exp: sdkmath.NewInt(1_057_690_000_000000000)},
		{valCount: 53, exp: sdkmath.NewInt(1_037_730_000_000000000)},
		{valCount: 54, exp: sdkmath.NewInt(1_018_510_000_000000000)},
		{valCount: 55, exp: sdkmath.NewInt(1_000_000_000_000000000)},
		{valCount: 56, exp: sdkmath.NewInt(982_140_000_000000000)},
		{valCount: 57, exp: sdkmath.NewInt(964_910_000_000000000)},
		{valCount: 58, exp: sdkmath.NewInt(948_270_000_000000000)},
		{valCount: 59, exp: sdkmath.NewInt(932_200_000_000000000)},
		{valCount: 60, exp: sdkmath.NewInt(916_660_000_000000000)},
		{valCount: 61, exp: sdkmath.NewInt(901_630_000_000000000)},
		{valCount: 62, exp: sdkmath.NewInt(887_090_000_000000000)},
		{valCount: 63, exp: sdkmath.NewInt(873_010_000_000000000)},
		{valCount: 64, exp: sdkmath.NewInt(859_370_000_000000000)},
		{valCount: 65, exp: sdkmath.NewInt(846_150_000_000000000)},
		{valCount: 66, exp: sdkmath.NewInt(833_330_000_000000000)},
		{valCount: 67, exp: sdkmath.NewInt(820_890_000_000000000)},
		{valCount: 68, exp: sdkmath.NewInt(808_820_000_000000000)},
		{valCount: 69, exp: sdkmath.NewInt(797_100_000_000000000)},
		{valCount: 70, exp: sdkmath.NewInt(785_710_000_000000000)},
		{valCount: 71, exp: sdkmath.NewInt(774_640_000_000000000)},
		{valCount: 72, exp: sdkmath.NewInt(763_880_000_000000000)},
		{valCount: 73, exp: sdkmath.NewInt(753_420_000_000000000)},
		{valCount: 74, exp: sdkmath.NewInt(743_240_000_000000000)},
		{valCount: 75, exp: sdkmath.NewInt(733_330_000_000000000)},
		{valCount: 76, exp: sdkmath.NewInt(723_680_000_000000000)},
		{valCount: 77, exp: sdkmath.NewInt(714_280_000_000000000)},
		{valCount: 78, exp: sdkmath.NewInt(705_120_000_000000000)},
		{valCount: 79, exp: sdkmath.NewInt(696_200_000_000000000)},
		{valCount: 80, exp: sdkmath.NewInt(687_500_000_000000000)},
		{valCount: 81, exp: sdkmath.NewInt(679_010_000_000000000)},
		{valCount: 82, exp: sdkmath.NewInt(670_730_000_000000000)},
		{valCount: 83, exp: sdkmath.NewInt(662_650_000_000000000)},
		{valCount: 84, exp: sdkmath.NewInt(654_760_000_000000000)},
		{valCount: 85, exp: sdkmath.NewInt(647_050_000_000000000)},
		{valCount: 86, exp: sdkmath.NewInt(639_530_000_000000000)},
		{valCount: 87, exp: sdkmath.NewInt(632_180_000_000000000)},
		{valCount: 88, exp: sdkmath.NewInt(625_000_000_000000000)},
		{valCount: 89, exp: sdkmath.NewInt(617_970_000_000000000)},
		{valCount: 90, exp: sdkmath.NewInt(611_110_000_000000000)},
		{valCount: 91, exp: sdkmath.NewInt(604_390_000_000000000)},
		{valCount: 92, exp: sdkmath.NewInt(597_820_000_000000000)},
		{valCount: 93, exp: sdkmath.NewInt(591_390_000_000000000)},
		{valCount: 94, exp: sdkmath.NewInt(585_100_000_000000000)},
		{valCount: 95, exp: sdkmath.NewInt(578_940_000_000000000)},
		{valCount: 96, exp: sdkmath.NewInt(572_910_000_000000000)},
		{valCount: 97, exp: sdkmath.NewInt(567_010_000_000000000)},
		{valCount: 98, exp: sdkmath.NewInt(561_220_000_000000000)},
		{valCount: 99, exp: sdkmath.NewInt(555_550_000_000000000)},
		{valCount: 100, exp: sdkmath.NewInt(550_000_000_000000000)},
		{valCount: 101, exp: sdkmath.NewInt(544_550_000_000000000)},
		{valCount: 102, exp: sdkmath.NewInt(539_210_000_000000000)},
		{valCount: 103, exp: sdkmath.NewInt(533_980_000_000000000)},
		{valCount: 104, exp: sdkmath.NewInt(528_840_000_000000000)},
		{valCount: 105, exp: sdkmath.NewInt(523_800_000_000000000)},
		{valCount: 106, exp: sdkmath.NewInt(518_860_000_000000000)},
		{valCount: 107, exp: sdkmath.NewInt(514_010_000_000000000)},
		{valCount: 108, exp: sdkmath.NewInt(509_250_000_000000000)},
		{valCount: 109, exp: sdkmath.NewInt(504_580_000_000000000)},
		{valCount: 110, exp: sdkmath.NewInt(500_000_000_000000000)},
		{valCount: 111, exp: sdkmath.NewInt(500_000_000_000000000)},
		{valCount: 112, exp: sdkmath.NewInt(500_000_000_000000000)},
	}

	for _, tc := range tests {
		name := tc.name
		if len(name) == 0 {
			name = fmt.Sprintf("%d validators", tc.valCount)
		}
		t.Run(name, func(t *testing.T) {
			if tc.opts == nil {
				tc.opts = DefaultRestrictionOptions
			}
			if tc.totalBond.IsNil() {
				tc.totalBond = sdkmath.NewIntFromUint64(10_000_000_000_000000000)
			}

			// Repeat this 1000 times to check determinism.
			for i := range 1000 {
				var maxValPct float32
				var act sdkmath.Int
				testFunc := func() {
					maxValPct = tc.opts.CalcMaxValPct(tc.valCount)
					act = CalcMaxValBond(tc.totalBond, maxValPct)
				}
				if assert.NotPanics(t, testFunc, "[%d]: CalcMaxValPct(%d) and CalcMaxValBond(%s, %.10f)", i, tc.valCount, tc.totalBond, maxValPct) {
					if assert.Equal(t, tc.exp.String(), act.String(), "[%d]: CalcMaxValPct(%d) and CalcMaxValBond(%s, %.10f) result", i, tc.valCount, tc.totalBond, maxValPct) {
						continue
					}
				}

				// It only gets here if one of those assertions failed.
				// If it was the first one, we give up on the rest, otherwise, keep going to see how often it fails.
				if i == 0 {
					break
				}
			}
		})
	}
}
