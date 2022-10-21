package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/cosmos/cosmos-sdk/types/errors"
	tmlog "github.com/tendermint/tendermint/libs/log"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"

	"github.com/jedib0t/go-pretty/v6/table"
	dbm "github.com/tendermint/tm-db"

	"regexp"
	"strings"
)

// Analyzer is used to analyze the application database.
type Analyzer struct {
	// HomePath is the path to the home directory (should contain the config and data directories).
	HomePath string
}

// Initialize prepares this Migrator by doing the following:
//  1. Calls ApplyDefaults()
//  2. Checks ValidateBasic()
//  3. Calls ReadSourceDataDir()
func (m *Analyzer) Initialize() error {
	m.ApplyDefaults()
	var err error
	if err = m.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

// ApplyDefaults fills in the defaults that it can, for values that aren't set yet.
func (m *Analyzer) ApplyDefaults() {
}

// ValidateBasic makes sure that everything is set in this Migrator.
func (m Analyzer) ValidateBasic() error {
	return nil
}

// Execute database analysis functionality.
func (m *Analyzer) Analyze(logger tmlog.Logger) error {
	dataPath := fmt.Sprintf("%s/data", m.HomePath)
	logger.Info(fmt.Sprintf("Analyzing application database in: %s", dataPath))
	db, err := dbm.NewDB("application", dbm.GoLevelDBBackend, dataPath)
	if err != nil {
		return errors.ErrIO.Wrapf("unable to open application database: %v", err)
	}
	defer db.Close()

	goLevelDB, ok := db.(*dbm.GoLevelDB)
	if !ok {
		return errors.ErrInvalidType.Wrapf("invalid logical DB type; expected: %T, got: %T", &dbm.GoLevelDB{}, db)
	}

	// print native stats
	levelDBStats, err := goLevelDB.DB().GetProperty("leveldb.stats")
	if err != nil {
		logger.Error("failed to get LevelDB stats: %v", err)
	}

	logger.Info(fmt.Sprintf("%s\n", levelDBStats))
	RenderTable(goLevelDB)

	return nil
}

func RenderTable(goLevelDB *dbm.GoLevelDB) {
	var (
		totalKeys    int
		totalKeySize int
		totalValSize int

		moduleStats = make(map[string][]int)
		moduleRe    = regexp.MustCompile(`s\/k:(\w+)\/`)
	)

	iter := goLevelDB.DB().NewIterator(nil, nil)
	for iter.Next() {
		keySize := len(iter.Key())
		valSize := len(iter.Value())

		totalKeys++
		totalKeySize += keySize
		totalValSize += valSize

		var statKey string

		keyStr := string(iter.Key())
		if strings.HasPrefix(keyStr, "s/k:") {
			tokens := moduleRe.FindStringSubmatch(keyStr)
			statKey = tokens[1]
		} else {
			statKey = "misc"
		}

		if moduleStats[statKey] == nil {
			// XXX/TODO: Move this into a struct
			//
			// 0: total set size
			// 1: total key size
			// 2: total value size
			moduleStats[statKey] = make([]int, 3)
		}

		moduleStats[statKey][0]++
		moduleStats[statKey][1] += keySize
		moduleStats[statKey][2] += valSize
	}

	// print application-specific stats
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Module", "Avg Key Size", "Avg Value Size", "Total Key Size", "Total Value Size", "Total Key Pairs"})

	modules := maps.Keys(moduleStats)
	SortSlice(modules)

	for _, m := range modules {
		stats := moduleStats[m]
		t.AppendRow([]interface{}{
			m,
			ByteCountDecimal(stats[1] / stats[0]),
			ByteCountDecimal(stats[2] / stats[0]),
			ByteCountDecimal(stats[1]),
			ByteCountDecimal(stats[2]),
			stats[0],
		})
	}

	t.AppendFooter(table.Row{"Total", "", "", ByteCountDecimal(totalKeySize), ByteCountDecimal(totalValSize), totalKeys})

	t.Render()
}

func SortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func ByteCountDecimal(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := int64(b) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
