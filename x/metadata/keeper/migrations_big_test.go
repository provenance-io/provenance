package keeper_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type MigrationsBigTestSuite struct {
	suite.Suite

	funcDepth int
	startTime time.Time

	app   *simapp.App
	ctx   sdk.Context
	store sdk.KVStore

	loadDir string
}

func (s *MigrationsBigTestSuite) SetupTest() {
	s.funcDepth = -1
	s.startTime = time.Now()

	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.store = s.ctx.KVStore(s.app.GetKey(types.ModuleName))

	// s.loadDir = "/Users/danielwedul/random-work/metadata/all/testnet-2022-01-03--19-34"
	s.loadDir = "/Users/danielwedul/random-work/metadata/all/2022-01-03--16-58"
	// s.loadDir = "/Users/danielwedul/random-work/metadata/all/small-2022-01-03--16-58"

	s.Require().NoError(s.LoadData(s.loadDir), "loading data")
}

type MDType string

const (
	ContractSpecs MDType = "contractspecs"
	RecordSpecs   MDType = "recordspecs"
	ScopeSpecs    MDType = "scopespecs"
	Scopes        MDType = "scopes"
	Sessions      MDType = "sessions"
	Records       MDType = "records"
)

type AddressFunc func(md interface{}) types.MetadataAddress

type ParserFunc func(bz []byte) ([]codec.ProtoMarshaler, error)

// FuncStarting logs that a function is starting.
// It returns the params needed by FuncEnding or FuncEndingAlways.
//
// Arguments provided will be converted to stings using %v and included as part of the function name.
// Minimal values needed to differentiate start/stop output lines should be provided.
// Long strings and complex structs should be avoided.
//
// Example 1: In a function named "foo", you have this:
//     FuncStarting()
//   The printed message will note that "foo" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Example 2: In a function named "bar", you have this:
//     FuncStarting(3 * time.Second)
//   The printed message will note that "bar: 3s" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Example 3:
//     func sum(ints ...int) {
//         FuncStarting(ints...)
//     }
//     sum(1, 2, 3, 4, 20, 21, 22)
//   The printed message will note that "sum: 1, 2, 3, 4, 20, 21, 22" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Standard Usage: defer s.FuncEnding(s.FuncStarting())
func (s *MigrationsBigTestSuite) FuncStarting(a ...interface{}) (time.Time, string) {
	s.funcDepth++
	name := GetFuncName(1, a...)
	s.LogAs(name, "Starting.")
	return time.Now(), name
}

// FuncEnding logs that a function is ending and includes the function duration.
//
// Usage: defer s.FuncEnding(s.FuncStarting())
func (s *MigrationsBigTestSuite) FuncEnding(start time.Time, name string) {
	s.LogAs(name, "Done. Duration: [%s].", time.Since(start))
	if s.funcDepth > -1 {
		s.funcDepth--
	}
}

// LogAs logs a message as the given func.
func (s MigrationsBigTestSuite) LogAs(funcName, format string, a ...interface{}) {
	s.T().Log(s.GetOutputPrefix(funcName) + fmt.Sprintf(format, a...))
}

// Log logs a message indicating the function it's being logged from.
func (s MigrationsBigTestSuite) Log(format string, a ...interface{}) {
	s.LogAs(GetFuncName(1), format, a...)
}

// GetOutputPrefix gets the prefix to add to all output.
func (s MigrationsBigTestSuite) GetOutputPrefix(funcName string) string {
	tabs := ""
	if s.funcDepth > 0 {
		tabs = strings.Repeat("  ", s.funcDepth)
	}
	return fmt.Sprintf("(%14s) %s[%s] ", DurClock(time.Since(s.startTime)), tabs, funcName)
}

// DurClock converts a duration to a string in minimal clock notation with nanosecond precision.
//
// - If one or more hours, format is "H:MM:SS.NNNNNNNNN", e.g. "12:01:02.000000000"
// - If less than one hour, format is "M:SS.NNNNNNNNN",   e.g. "34:00.000000789"
// - If less than one minute, format is "S.NNNNNNNNN",    e.g. "56.000456000"
// - If less than one second, format is "0.NNNNNNNNN",    e.g. "0.123000000"
func DurClock(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes())
	s := int(d.Seconds())
	n := int(d.Nanoseconds()) - 1000000000*s
	s = s - 60*m
	m = m - 60*h
	switch {
	case h > 0:
		return fmt.Sprintf("%d:%02d:%02d.%09d", h, m, s, n)
	case m > 0:
		return fmt.Sprintf("%d:%02d.%09d", m, s, n)
	default:
		return fmt.Sprintf("%d.%09d", s, n)
	}
}

// GetFuncName gets the name of the function at the given depth.
//
// depth 0 = the function calling GetFuncName.
// depth 1 = the function calling the function calling GetFuncName.
// etc.
//
// Extra arguments provided will be converted to stings using %v and included as part of the function name.
// Only values needed to differentiate start/stop output lines should be provided.
// Long strings and complex structs should be avoided.
func GetFuncName(depth int, a ...interface{}) string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, more := frames.Next()
	for more && depth > 0 {
		frame, more = frames.Next()
		depth--
	}
	name := strings.TrimPrefix(frame.Function, "main.")
	name = strings.TrimPrefix(name, "github.com/provenance-io/provenance/x/metadata/keeper_test.MigrationsBigTestSuite.")
	name = strings.TrimPrefix(name, "github.com/provenance-io/provenance/x/metadata/keeper_test.(*MigrationsBigTestSuite).")
	// Using a switch to prevent calling strings.Join for small (common) use cases. Saves a little mem and processing.
	switch len(a) {
	case 0:
		// do nothing
	case 1:
		name += fmt.Sprintf(": %v", a[0])
	case 2:
		name += fmt.Sprintf(": %v, %v", a[0], a[1])
	case 3:
		name += fmt.Sprintf(": %v, %v, %v", a[0], a[1], a[2])
	default:
		args := make([]string, len(a))
		for i, arg := range a {
			args[i] = fmt.Sprintf("%v", arg)
		}
		name += fmt.Sprintf(": %s", strings.Join(args, ", "))
	}
	return name
}

func (s MigrationsBigTestSuite) getFuncs(mdType MDType) (AddressFunc, ParserFunc, error) {
	switch mdType {
	case ContractSpecs:
		return func(md interface{}) types.MetadataAddress { return md.(*types.ContractSpecification).SpecificationId },
			func(bz []byte) ([]codec.ProtoMarshaler, error) {
				var resp types.ContractSpecificationsAllResponse
				err := s.app.AppCodec().UnmarshalJSON(bz, &resp)
				if err != nil {
					return nil, err
				}
				rv := make([]codec.ProtoMarshaler, len(resp.ContractSpecifications))
				for i, v := range resp.ContractSpecifications {
					rv[i] = v.Specification
				}
				return rv, nil
			},
			nil
	case RecordSpecs:
		return func(md interface{}) types.MetadataAddress { return md.(*types.RecordSpecification).SpecificationId },
			func(bz []byte) ([]codec.ProtoMarshaler, error) {
				var resp types.RecordSpecificationsAllResponse
				err := s.app.AppCodec().UnmarshalJSON(bz, &resp)
				if err != nil {
					return nil, err
				}
				rv := make([]codec.ProtoMarshaler, len(resp.RecordSpecifications))
				for i, v := range resp.RecordSpecifications {
					rv[i] = v.Specification
				}
				return rv, nil
			},
			nil
	case ScopeSpecs:
		return func(md interface{}) types.MetadataAddress { return md.(*types.ScopeSpecification).SpecificationId },
			func(bz []byte) ([]codec.ProtoMarshaler, error) {
				var resp types.ScopeSpecificationsAllResponse
				err := s.app.AppCodec().UnmarshalJSON(bz, &resp)
				if err != nil {
					return nil, err
				}
				rv := make([]codec.ProtoMarshaler, len(resp.ScopeSpecifications))
				for i, v := range resp.ScopeSpecifications {
					rv[i] = v.Specification
				}
				return rv, nil
			},
			nil
	case Scopes:
		return func(md interface{}) types.MetadataAddress { return md.(*types.Scope).ScopeId },
			func(bz []byte) ([]codec.ProtoMarshaler, error) {
				var resp types.ScopesAllResponse
				err := s.app.AppCodec().UnmarshalJSON(bz, &resp)
				if err != nil {
					return nil, err
				}
				rv := make([]codec.ProtoMarshaler, len(resp.Scopes))
				for i, v := range resp.Scopes {
					rv[i] = v.Scope
				}
				return rv, nil
			},
			nil
	case Sessions:
		return func(md interface{}) types.MetadataAddress { return md.(*types.Session).SessionId },
			func(bz []byte) ([]codec.ProtoMarshaler, error) {
				var resp types.SessionsAllResponse
				err := s.app.AppCodec().UnmarshalJSON(bz, &resp)
				if err != nil {
					return nil, err
				}
				rv := make([]codec.ProtoMarshaler, len(resp.Sessions))
				for i, v := range resp.Sessions {
					rv[i] = v.Session
				}
				return rv, nil
			},
			nil
	case Records:
		return func(md interface{}) types.MetadataAddress { return md.(*types.Record).GetRecordAddress() },
			func(bz []byte) ([]codec.ProtoMarshaler, error) {
				var resp types.RecordsAllResponse
				err := s.app.AppCodec().UnmarshalJSON(bz, &resp)
				if err != nil {
					return nil, err
				}
				rv := make([]codec.ProtoMarshaler, len(resp.Records))
				for i, v := range resp.Records {
					rv[i] = v.Record
				}
				return rv, nil
			},
			nil
	default:
		return nil, nil, fmt.Errorf("Unknown metadata type [%s].", mdType)
	}
}

func (s *MigrationsBigTestSuite) LoadData(dir string) error {
	defer s.FuncEnding(s.FuncStarting())
	s.loadDir = dir

	order := []MDType{ContractSpecs, RecordSpecs, ScopeSpecs, Scopes, Sessions, Records}

	for i, mdType := range order {
		logLead := fmt.Sprintf("%d/%d %s:", i+1, len(order), mdType)
		s.Log("%s Loading", logLead)
		addressF, constructorF, err := s.getFuncs(mdType)
		if err != nil {
			s.Log("%s Error: %v", logLead, err)
			return err
		}
		mdTypeDir := filepath.Join(s.loadDir, "q-res-all-"+string(mdType))
		err = s.LoadAllFromDir(logLead, mdTypeDir, addressF, constructorF)
		if err != nil {
			s.Log("%s Error: %v", logLead, err)
			return err
		}
		s.Log("%s Done", logLead)
	}

	return nil
}

func (s *MigrationsBigTestSuite) LoadAllFromDir(logLead, dir string, addresser AddressFunc, parser ParserFunc) error {
	defer s.FuncEnding(s.FuncStarting(dir))
	files, err := getFilesIn(dir)
	if err != nil {
		return err
	}
	s.Log("%s Found %d files in %s", logLead, len(files), dir)
	count := 0
	for f, file := range files {
		flogLead := fmt.Sprintf("%s [%d/%d] %s:", logLead, f+1, len(files), file)
		data, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			s.Log("%s Error reading file: %v", flogLead, err)
			return err
		}
		entries, err := parser(data)
		if err != nil {
			s.Log("%s Error parsing json: %v", flogLead, err)
			return err
		}
		for i, entry := range entries {
			var bz []byte
			bz, err = s.app.AppCodec().Marshal(entry)
			if err != nil {
				s.Log("%s entry %d of %d: could not marshall to proto: %v", flogLead, i+1, len(entries), err)
				return err
			}
			addr := addresser(entry)
			s.store.Set(addr, bz)
		}
		count += len(entries)
		s.Log("%s Done. Added %d new entries. Now have %d total.", flogLead, len(entries), count)
	}
	return nil
}

func getFilesIn(dir string) ([]string, error) {
	dirContents, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(dirContents))
	for _, entry := range dirContents {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func TestMigrationsBigTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationsBigTestSuite))
}

func (s MigrationsBigTestSuite) TestTheLoad() {
	s.Fail("Just failing to see the logs.")
}
