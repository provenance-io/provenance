package keeper_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal"
	"github.com/provenance-io/provenance/internal/provutils"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/testutil/testlog"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type wrappedKVStore struct {
	storetypes.KVStore
	calls *storeCalls
}

func (s *wrappedKVStore) Delete(key []byte) {
	s.calls.Deletions = append(s.calls.Deletions, key)
	if s.KVStore != nil {
		s.KVStore.Delete(key)
	}
}

type storeCalls struct {
	Deletions [][]byte
}

type testKeeper3To4 struct {
	keeper.Keeper3To4

	logBuffer bytes.Buffer

	storeCalls *storeCalls

	unmarshalErrs []string

	setScopeValueOwnerErrs   []string
	setScopeValueOwnersCalls []*provutils.Pair[types.MetadataAddress, string]
}

func newTestKeeper3To4(kpr keeper.Keeper) *testKeeper3To4 {
	return &testKeeper3To4{
		Keeper3To4: keeper.NewKeeper3To4(kpr),
		storeCalls: &storeCalls{},
	}
}

func (k *testKeeper3To4) Logger(_ sdk.Context) log.Logger {
	return internal.NewBufferedInfoLogger(&k.logBuffer)
}

func (k *testKeeper3To4) GetStore(ctx sdk.Context) storetypes.KVStore {
	store := k.Keeper3To4.GetStore(ctx)
	return &wrappedKVStore{
		KVStore: store,
		calls:   k.storeCalls,
	}
}

func (k *testKeeper3To4) Unmarshal(bz []byte, ptr proto.Message) error {
	if len(k.unmarshalErrs) > 0 {
		rv := k.unmarshalErrs[0]
		k.unmarshalErrs = k.unmarshalErrs[1:]
		if len(rv) > 0 {
			return errors.New(rv)
		}
	}
	return k.Keeper3To4.Unmarshal(bz, ptr)
}

func (k *testKeeper3To4) SetScopeValueOwner(ctx sdk.Context, scopeID types.MetadataAddress, newValueOwner string) error {
	k.setScopeValueOwnersCalls = append(k.setScopeValueOwnersCalls, provutils.NewPair(scopeID, newValueOwner))
	if len(k.setScopeValueOwnerErrs) > 0 {
		rv := k.setScopeValueOwnerErrs[0]
		k.setScopeValueOwnerErrs = k.setScopeValueOwnerErrs[1:]
		if len(rv) > 0 {
			return errors.New(rv)
		}
	}
	return nil
}

// GetLogOutput gets the log buffer contents. This (probably) also clears the log buffer.
func (k *testKeeper3To4) GetLogOutput(t *testing.T, msg string, args ...interface{}) []string {
	return getLogOutput(t, k.logBuffer, msg, args...)
}

// getLogOutput gets the log buffer contents. This (probably) also clears the log buffer.
func getLogOutput(t *testing.T, logBuffer bytes.Buffer, msg string, args ...interface{}) []string {
	logOutput := logBuffer.String()
	t.Logf(msg+" log output:\n%s", append(args, logOutput))
	return internal.SplitLogLines(logOutput)
}

func writeScope(t *testing.T, kpr keeper.Keeper, ctx sdk.Context, scope types.Scope, msgAndArgs ...interface{}) {
	if len(msgAndArgs) == 0 {
		msgAndArgs = append(msgAndArgs, "V3WriteNewScope")
	} else {
		switch v := msgAndArgs[0].(type) {
		case string:
			msgAndArgs[0] = "V3WriteNewScope: " + v
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			if len(msgAndArgs) == 1 {
				msgAndArgs = []interface{}{"[%d]: V3WriteNewScope", v}
			}
		}
	}
	err := kpr.V3WriteNewScope(ctx, scope)
	require.NoError(t, err, msgAndArgs...)
}

func TestMigrate3to4(t *testing.T) {
	addrs := []string{
		newAddr("one").String(),   // cosmos1dahx2h6lta047h6lta047h6lta047h6lq2tdll
		newAddr("two").String(),   // cosmos1w3mk7h6lta047h6lta047h6lta047h6lakjg9t
		newAddr("three").String(), // cosmos1w358yet9ta047h6lta047h6lta047h6lma20rt
		newAddr("four").String(),  // cosmos1vehh2ujlta047h6lta047h6lta047h6l6dna47
		newAddr("five").String(),  // cosmos1ve5hve2lta047h6lta047h6lta047h6ltfdqga
		newAddr("six").String(),   // cosmos1wd5hsh6lta047h6lta047h6lta047h6la8xq2y
		newAddr("seven").String(), // cosmos1wdjhvetwta047h6lta047h6lta047h6ldw3pw9
		newAddr("eight").String(), // cosmos1v45kw6r5ta047h6lta047h6lta047h6l3j2t8s
		newAddr("nine").String(),  // cosmos1de5kue2lta047h6lta047h6lta047h6lfr0cjd
		newAddr("ten").String(),   // cosmos1w3jkuh6lta047h6lta047h6lta047h6lqcnhxg
	}

	newUUID := func(i int) uuid.UUID {
		// Sixteen 9's is the largest number we can handle; one more and it's 17 digits.
		require.LessOrEqual(t, i, 9999999999999999, "value provided to newScopeID")
		str := fmt.Sprintf("________________%d", i)
		str = str[len(str)-16:]
		rv, err := uuid.FromBytes([]byte(str))
		require.NoError(t, err, "uuid.FromBytes([]byte(%q))", str)
		return rv
	}
	newScopeID := func(i int) types.MetadataAddress {
		return types.ScopeMetadataAddress(newUUID(i))
	}
	newSpecID := func(i int) types.MetadataAddress {
		// The spec id shouldn't really matter in here, but I want it different from a scope's i.
		// So I do some math to make it seem kind of random, but is still deterministic.
		// 48, 67, and 81 were picked randomly and have no special meaning.
		// 50,000 was chosen so that maybe some spec ids get used more than once.
		j := (i + 48) * (i + 67) * (i + 81)
		return types.ScopeSpecMetadataAddress(newUUID(j % 50_000))
	}
	newScope := func(i int) types.Scope {
		rv := types.Scope{
			ScopeId:           newScopeID(i),
			SpecificationId:   newSpecID(i),
			ValueOwnerAddress: addrs[i%len(addrs)],
		}
		if i%7 == 0 {
			rv.ValueOwnerAddress = ""
		}
		ownerCount := (i % 3) + 1 // 1 to 3.
		if ownerCount > 0 {
			rv.Owners = make([]types.Party, ownerCount)
			for o := range rv.Owners {
				rv.Owners[o].Address = addrs[(i*i+o)%len(addrs)]
				rv.Owners[o].Role = types.PartyType(1 + (i+o)%11) // 11 different roles, 1 to 11.
			}
		}
		return rv
	}

	var logBuffer bytes.Buffer
	app := func() *simapp.App {
		// Swap in a logger maker that writes to our logBuffer, and defer a call to set it back when we're done.
		defer simapp.SetLoggerMaker(simapp.SetLoggerMaker(simapp.BufferedInfoLoggerMaker(&logBuffer)))
		return simapp.Setup(t)
	}()
	ctx1 := FreshCtx(app)
	for i := 1; i <= 100; i++ {
		j := i * i * i * i * i * i * i // i^7. When i = 100, j = 100,000,000,000,000 = 15 digits.
		if i%2 == 0 {
			j = 1_000_000_000_000_000 - j
		}
		writeScope(t, app.MetadataKeeper, ctx1, newScope(i), i)
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T, ctx sdk.Context)
		expErr  string
		expLogs []string
	}{
		{
			name: "error from one",
			setup: func(t *testing.T, ctx sdk.Context) {
				key := newScopeID(5000000000000000)
				value := []byte{0, 0, 0}
				store := ctx.KVStore(app.MetadataKeeper.GetStoreKey())
				store.Set(key, value)
			},
			expErr: "error reading scope " + newScopeID(5000000000000000).String() + " from state: proto: Scope: illegal tag 0 (wire type 0)",
			expLogs: []string{
				"INF Starting migration of x/metadata from 3 to 4. module=x/metadata",
				"INF Moving scope value owner data into x/bank ledger. module=x/metadata",
				"ERR [1]: ScopeID=\"" + newScopeID(5000000000000000).String() + "\" bytes=\"\\x00\\x00\\x00\" module=x/metadata",
				"ERR Error migrating scope value owners. error=\"error reading scope " +
					newScopeID(5000000000000000).String() + " from state: proto: Scope: illegal tag 0 (wire type 0)\" " +
					"module=x/metadata",
			},
		},
		{
			name: "all good",
			expLogs: []string{
				"INF Starting migration of x/metadata from 3 to 4. module=x/metadata",
				"INF Moving scope value owner data into x/bank ledger. module=x/metadata",
				"INF Done moving scope value owners into bank module. module=x/metadata scopes=100 value owners=86",
				"INF Done migrating x/metadata from 3 to 4. module=x/metadata",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := FreshCtx(app).CacheContext()
			if tc.setup != nil {
				tc.setup(t, ctx)
			}

			logBuffer.Reset()
			migrator := keeper.NewMigrator(app.MetadataKeeper)
			var err error
			testFunc := func() {
				err = migrator.Migrate3To4(ctx)
			}
			require.NotPanics(t, testFunc, "Migrate3To4")
			assertions.AssertErrorValue(t, err, tc.expErr, "error from Migrate3To4")
			actLogs := getLogOutput(t, logBuffer, "Migrate3To4")
			assert.Equal(t, tc.expLogs, actLogs, "logs messages emitted during Migrate3To4")
		})
	}
}

func TestMigrateValueOwners(t *testing.T) {
	newUUID := func(i int) uuid.UUID {
		// Sixteen 9's is the largest number we can handle; one more and it's 17 digits.
		require.LessOrEqual(t, i, 9999999999999999, "value provided to newScopeID")
		str := fmt.Sprintf("________________%d", i)
		str = str[len(str)-16:]
		rv, err := uuid.FromBytes([]byte(str))
		require.NoError(t, err, "uuid.FromBytes([]byte(%q))", str)
		return rv
	}
	newScopeID := func(i int) types.MetadataAddress {
		return types.ScopeMetadataAddress(newUUID(i))
	}
	newSpecID := func(i int) types.MetadataAddress {
		// The spec id shouldn't really matter in here, but I want it different from a scope's i.
		// So I do some math to make it seem kind of random, but is still deterministic.
		// 48, 67, and 81 were picked randomly and have no special meaning.
		// 50,000 was chosen so that maybe some spec ids get used more than once.
		j := (i + 48) * (i + 67) * (i + 81)
		return types.ScopeSpecMetadataAddress(newUUID(j % 50_000))
	}
	newScope := func(i int, valueOwnerBase string, owners ...string) types.Scope {
		rv := types.Scope{
			ScopeId:         newScopeID(i),
			SpecificationId: newSpecID(i),
		}
		if len(valueOwnerBase) > 0 {
			rv.ValueOwnerAddress = newAddr(valueOwnerBase).String()
		}
		for _, owner := range owners {
			rv.Owners = append(rv.Owners, types.Party{Address: newAddr(owner).String(), Role: 5})
		}
		return rv
	}

	app := simapp.Setup(t)

	tests := []struct {
		name             string
		setup            func(t *testing.T, ctx sdk.Context)
		injSetErrs       []string
		injUnmarshalErrs []string
		expErr           string
		expLogs          []string
		expSetCount      int
		expDelCount      int
	}{
		{
			name: "no scopes",
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
				"INF Done moving scope value owners into bank module. scopes=0 value owners=0",
			},
		},
		{
			name: "one scope: unmarshal error",
			setup: func(t *testing.T, ctx sdk.Context) {
				writeScope(t, app.MetadataKeeper, ctx, newScope(123_456_789, "vo_addr1"))
			},
			injUnmarshalErrs: []string{"yoko was not wrong"},
			expErr:           "error reading scope " + newScopeID(123_456_789).String() + " from state: yoko was not wrong",
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
				"ERR [1]: ScopeID=\"" + newScopeID(123_456_789).String() +
					"\" bytes=\"\\n\\x11\\x00_______123456789\\x12\\x11\\x04___________30944*-" + newAddr("vo_addr1").String() + "\"",
			},
		},
		{
			name: "one scope: set error",
			setup: func(t *testing.T, ctx sdk.Context) {
				writeScope(t, app.MetadataKeeper, ctx, newScope(23, "vo_addr2"))
			},
			injSetErrs: []string{"maybe jethro tull was the greatest"},
			expErr: "could not migrate scope " + newScopeID(23).String() + " value owner \"" +
				newAddr("vo_addr2").String() + "\" to bank module: maybe jethro tull was the greatest",
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
			},
			expSetCount: 1,
			expDelCount: 0,
		},
		{
			name: "one scope: no value owner",
			setup: func(t *testing.T, ctx sdk.Context) {
				writeScope(t, app.MetadataKeeper, ctx, newScope(37373, ""))
			},
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
				"INF Done moving scope value owners into bank module. scopes=1 value owners=0",
			},
			expSetCount: 0,
			expDelCount: 0,
		},
		{
			name: "one scope: with value owner",
			setup: func(t *testing.T, ctx sdk.Context) {
				writeScope(t, app.MetadataKeeper, ctx, newScope(37373, "mineminemine"))
			},
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
				"INF Done moving scope value owners into bank module. scopes=1 value owners=1",
			},
			expSetCount: 1,
			expDelCount: 2,
		},
		{
			name: "three scopes: unmarshal error from second",
			setup: func(t *testing.T, ctx sdk.Context) {
				writeScope(t, app.MetadataKeeper, ctx, newScope(5, "addr1", "addr1"), 1)
				writeScope(t, app.MetadataKeeper, ctx, newScope(6, "addr2"), 2)
				writeScope(t, app.MetadataKeeper, ctx, newScope(7, "addr3"), 3)
			},
			injUnmarshalErrs: []string{"", "radiohead is only okay"},
			expErr:           "error reading scope " + newScopeID(6).String() + " from state: radiohead is only okay",
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
				"ERR [2]: ScopeID=\"" + newScopeID(6).String() +
					"\" bytes=\"\\n\\x11\\x00_______________6\\x12\\x11\\x04___________42954*-" + newAddr("addr2").String() + "\"",
			},
			expSetCount: 1,
			expDelCount: 1,
		},
		{
			name: "three scopes: set error from second",
			setup: func(t *testing.T, ctx sdk.Context) {
				writeScope(t, app.MetadataKeeper, ctx, newScope(71, "ayyy", "ayyy"), 1)
				writeScope(t, app.MetadataKeeper, ctx, newScope(82, "bee"), 2)
				writeScope(t, app.MetadataKeeper, ctx, newScope(93, "see"), 3)
			},
			injSetErrs: []string{"", "fatboy slim lost that fight"},
			expErr: "could not migrate scope " + newScopeID(82).String() + " value owner \"" +
				newAddr("bee").String() + "\" to bank module: fatboy slim lost that fight",
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
			},
			expSetCount: 2,
			expDelCount: 1,
		},
		{
			name: "three scopes: no value owner in second",
			setup: func(t *testing.T, ctx sdk.Context) {
				writeScope(t, app.MetadataKeeper, ctx, newScope(765, "one", "one"), 1)
				writeScope(t, app.MetadataKeeper, ctx, newScope(876, "", "two"), 2)
				writeScope(t, app.MetadataKeeper, ctx, newScope(987, "three"), 3)
			},
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
				"INF Done moving scope value owners into bank module. scopes=3 value owners=2",
			},
			expSetCount: 2,
			expDelCount: 3, // = one from 765 + two from 987 (none from 876),
		},
		{
			name: "30,005 scopes",
			setup: func(t *testing.T, ctx sdk.Context) {
				addrs := []string{
					newAddr("one").String(),   // cosmos1dahx2h6lta047h6lta047h6lta047h6lq2tdll
					newAddr("two").String(),   // cosmos1w3mk7h6lta047h6lta047h6lta047h6lakjg9t
					newAddr("three").String(), // cosmos1w358yet9ta047h6lta047h6lta047h6lma20rt
					newAddr("four").String(),  // cosmos1vehh2ujlta047h6lta047h6lta047h6l6dna47
					newAddr("five").String(),  // cosmos1ve5hve2lta047h6lta047h6lta047h6ltfdqga
					newAddr("six").String(),   // cosmos1wd5hsh6lta047h6lta047h6lta047h6la8xq2y
					newAddr("seven").String(), // cosmos1wdjhvetwta047h6lta047h6lta047h6ldw3pw9
					newAddr("eight").String(), // cosmos1v45kw6r5ta047h6lta047h6lta047h6l3j2t8s
					newAddr("nine").String(),  // cosmos1de5kue2lta047h6lta047h6lta047h6lfr0cjd
					newAddr("ten").String(),   // cosmos1w3jkuh6lta047h6lta047h6lta047h6lqcnhxg
				}

				for i := 1; i <= 30_005; i++ {
					scope := types.Scope{
						ScopeId:           newScopeID(i),
						SpecificationId:   newSpecID(i),
						ValueOwnerAddress: addrs[i%len(addrs)],
					}
					if i%7 == 0 {
						scope.ValueOwnerAddress = ""
					}
					ownerCount := (i % 3) + 1 // 1 to 3.
					if ownerCount > 0 {
						scope.Owners = make([]types.Party, ownerCount)
						for o := range scope.Owners {
							scope.Owners[o].Address = addrs[(i*i+o)%len(addrs)]
							scope.Owners[o].Role = types.PartyType(1 + (i+o)%11) // 11 different roles, 1 to 11.
						}
					}
					writeScope(t, app.MetadataKeeper, ctx, scope, i)
				}
			},
			expLogs: []string{
				"INF Moving scope value owner data into x/bank ledger.",
				"INF Progress update: scopes=10000 value owners=8571",
				"INF Progress update: scopes=20000 value owners=17143",
				"INF Progress update: scopes=30000 value owners=25715",
				"INF Done moving scope value owners into bank module. scopes=30005 value owners=25719",
			},
			expSetCount: 25719,
			expDelCount: 41150,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := FreshCtx(app).CacheContext()
			if tc.setup != nil {
				tc.setup(t, ctx)
			}

			kpr := newTestKeeper3To4(app.MetadataKeeper)
			kpr.setScopeValueOwnerErrs = tc.injSetErrs
			kpr.unmarshalErrs = tc.injUnmarshalErrs

			var err error
			testFunc := func() {
				err = keeper.MigrateValueOwners(ctx, kpr)
			}
			require.NotPanics(t, testFunc, "migrateValueOwners")
			actLogs := kpr.GetLogOutput(t, "migrateValueOwners")
			assertions.AssertErrorValue(t, err, tc.expErr, "error from migrateValueOwners")
			assert.Equal(t, tc.expLogs, actLogs, "logs messages emitted during migrateValueOwners")

			actSetCount := len(kpr.setScopeValueOwnersCalls)
			assert.Equal(t, tc.expSetCount, actSetCount, "calls made to SetScopeValueOwner")
			actDelCount := len(kpr.storeCalls.Deletions)
			assert.Equal(t, tc.expDelCount, actDelCount, "store deletions made")
		})
	}
}

func TestMigrateValueOwnerToBank(t *testing.T) {
	newScopeID := func(b byte) types.MetadataAddress {
		rv := make(types.MetadataAddress, 17)
		rv[0] = types.ScopeKeyPrefix[0]
		for i := 1; i < len(rv); i++ {
			rv[i] = b
		}
		return rv
	}

	scopeID := newScopeID('a')
	addr := sdk.AccAddress("the_address_________").String()
	testlog.WriteVariables(t, "stuff", "scopeID", scopeID, "addr", addr)
	tests := []struct {
		name      string
		scope     types.Scope
		injectErr string
		expErr    string
	}{
		{
			name:      "error setting value owner",
			scope:     types.Scope{ScopeId: scopeID, ValueOwnerAddress: addr},
			injectErr: "nickleback was okay",
			expErr:    "could not migrate scope " + scopeID.String() + " value owner \"" + addr + "\" to bank module: nickleback was okay",
		},
		{
			name:  "all good",
			scope: types.Scope{ScopeId: scopeID, ValueOwnerAddress: addr},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kpr := &testKeeper3To4{}
			if len(tc.injectErr) > 0 {
				kpr.setScopeValueOwnerErrs = append(kpr.setScopeValueOwnerErrs, tc.injectErr)
			}
			store := &wrappedKVStore{calls: &storeCalls{}}
			expDels := len(tc.expErr) == 0
			expSetCalls := []*provutils.Pair[types.MetadataAddress, string]{
				provutils.NewPair(tc.scope.ScopeId, tc.scope.ValueOwnerAddress),
			}

			var err error
			testFunc := func() {
				err = keeper.MigrateValueOwnerToBank(sdk.Context{}, kpr, store, tc.scope)
			}
			require.NotPanics(t, testFunc, "migrateValueOwnerToBank")
			assertions.AssertErrorValue(t, err, tc.expErr, "error from migrateValueOwnerToBank")
			actSetCalls := kpr.setScopeValueOwnersCalls
			assert.Equal(t, expSetCalls, actSetCalls, "calls made to SetScopeValueOwner in migrateValueOwnerToBank")
			actDels := store.calls.Deletions
			if expDels {
				assert.NotEmpty(t, actDels, "store deletions")
			} else {
				assert.Empty(t, actDels, "store deletions")
			}
		})
	}
}

func TestDeleteValueOwnerIndexEntries(t *testing.T) {
	owner1 := sdk.AccAddress("1_owner_address_____").String()    // cosmos1x90k7amwv4e97ctyv3ex2umnta047h6lvq72fg
	owner2 := sdk.AccAddress("2_owner_address_____").String()    // cosmos1xf0k7amwv4e97ctyv3ex2umnta047h6lw6hvgd
	owner3 := sdk.AccAddress("3_owner_address_____").String()    // cosmos1xd0k7amwv4e97ctyv3ex2umnta047h6lhtswsw
	otherAddr := sdk.AccAddress("other_address_______").String() // cosmos1da6xsetjtaskgerjv4ehxh6lta047h6l3cc9z2
	testlog.WriteVariables(t, "addresses",
		"owner1", owner1,
		"owner2", owner2,
		"owner3", owner3,
		"otherAddr", otherAddr,
	)

	newScopeID := func(b byte) types.MetadataAddress {
		rv := make(types.MetadataAddress, 17)
		rv[0] = types.ScopeKeyPrefix[0]
		for i := 1; i < len(rv); i++ {
			rv[i] = b
		}
		return rv
	}
	scopeID1 := newScopeID('1') // scope1qqcnzvf3xycnzvf3xycnzvf3xycs2xyeyk
	scopeID2 := newScopeID('2') // scope1qqeryv3jxgeryv3jxgeryv3jxgeqy48g0a
	scopeID3 := newScopeID('3') // scope1qqenxvenxvenxvenxvenxvenxvesqa360g
	testlog.WriteVariables(t, "scopes",
		"scopeID1", scopeID1,
		"scopeID2", scopeID2,
		"scopeID3", scopeID3,
	)

	owners := []types.Party{{Address: owner1}, {Address: owner2}, {Address: owner3}}
	newScope := func(id types.MetadataAddress, valueOwner string) types.Scope {
		return types.Scope{
			ScopeId:           id,
			Owners:            owners,
			ValueOwnerAddress: valueOwner,
		}
	}

	tests := []struct {
		name    string
		scope   types.Scope
		expDel1 bool
		expDel2 bool // if true, expDel1 is also treated as true.
	}{
		{
			name:  "empty value owner address",
			scope: newScope(scopeID1, ""),
		},
		{
			name:  "invalid value owner address",
			scope: newScope(scopeID1, "nope"),
		},
		{
			name:    "value owner also first owner of three",
			scope:   newScope(scopeID3, owner1),
			expDel1: true,
		},
		{
			name:    "value owner also second owner of three",
			scope:   newScope(scopeID2, owner2),
			expDel1: true,
		},
		{
			name:    "value owner also third owner of three",
			scope:   newScope(scopeID1, owner3),
			expDel1: true,
		},
		{
			name:    "value owner not owner",
			scope:   newScope(scopeID2, otherAddr),
			expDel2: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expDels [][]byte
			if tc.expDel1 || tc.expDel2 {
				expDels = make([][]byte, 1, 2)
				// If the value owner isn't valid, we shouldn't be expecting any deletions.
				vo, _ := sdk.AccAddressFromBech32(tc.scope.ValueOwnerAddress)
				expDels[0] = append(expDels[0], 0x18)
				expDels[0] = append(expDels[0], byte(len(vo)))
				expDels[0] = append(expDels[0], vo...)
				expDels[0] = append(expDels[0], tc.scope.ScopeId...)
				if tc.expDel2 {
					expDels = append(expDels, types.GetAddressScopeCacheKey(vo, tc.scope.ScopeId))
				}
			}

			store := &wrappedKVStore{calls: &storeCalls{}}
			testFunc := func() {
				keeper.DeleteValueOwnerIndexEntries(store, tc.scope)
			}
			require.NotPanics(t, testFunc, "deleteValueOwnerIndexEntries")
			actDels := store.calls.Deletions
			assert.Equal(t, expDels, actDels, "store deletions")
		})
	}
}
