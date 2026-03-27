package app

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListInterfaceRegistry(t *testing.T) {
	// This test will output all interfaces known to the interface registry, then it will
	// output all concrete types in the interface registry for each of those interfaces.
	// Output is done to the test logs.

	// Example output:
	// Interfaces (39):
	// <... snip ...>
	//    6: cosmos.base.v1beta1.Msg
	// <... snip ...>
	// [6/39]: cosmos.base.v1beta1.Msg (246):
	// [6/39]:     1: /cosmos.auth.v1beta1.MsgUpdateParams
	// [6/39]:     2: /cosmos.authz.v1beta1.MsgExec
	// <... snip ...>

	// digitFormatForMax returns a format string of the length of the provided maximum number.
	// E.g. digitFormatForMax(10) returns "%2d". digitFormatForMax(382920) returns "%6d".
	digitFormatForMax := func(maximum int) string {
		return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", maximum)))
	}

	// addLineNumbers adds line numbers to each string.
	addLineNumbers := func(lines []string, startAt int) []string {
		if len(lines) == 0 {
			return []string{}
		}
		lineFmt := digitFormatForMax(len(lines)-1+startAt) + ": %s"
		rv := make([]string, len(lines))
		for i, line := range lines {
			rv[i] = fmt.Sprintf(lineFmt, i+startAt, line)
		}
		return rv
	}

	// prefixLines adds the provided pre string to the start of each line.
	prefixLines := func(pre string, lines []string) []string {
		if lines == nil {
			return nil
		}
		rv := make([]string, len(lines))
		for i, line := range lines {
			rv[i] = pre + line
		}
		return rv
	}

	// prefixNumberJoin adds line numbers (starting at 1), then adds the provided prefix to each line, then joins it all into a multi-line string.
	prefixNumberJoin := func(pre string, lines []string) string {
		return strings.Join(prefixLines(pre, addLineNumbers(lines, 1)), "\n")
	}

	// logList writes the provided list to the buffer under the provided header with the prefix on each line.
	logList := func(buffer *strings.Builder, header string, pre string, list []string) {
		_, err := fmt.Fprintf(buffer, "%s%s (%d):\n", pre, header, len(list))
		require.NoError(t, err, "Fprintf for %q", header)
		if len(list) > 0 {
			_, err = buffer.WriteString(prefixNumberJoin(pre+"  ", list) + "\n")
			require.NoError(t, err, "buffer.WriteString for %q list", header)
		}
	}

	app := Setup(t)
	ifaces := app.interfaceRegistry.ListAllInterfaces()
	slices.Sort(ifaces)
	var buffer strings.Builder
	buffer.WriteRune('\n')
	logList(&buffer, "Interfaces", "", ifaces)

	for i, iface := range ifaces {
		impls := app.interfaceRegistry.ListImplementations(iface)
		slices.Sort(impls)
		logList(&buffer, iface, fmt.Sprintf("[%d/%d]: ", i+1, len(ifaces)), impls)
	}

	t.Log(buffer.String())
}
