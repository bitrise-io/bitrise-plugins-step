package create

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateIDFromString(t *testing.T) {
	require.Equal(t, "bitrise-step-a-simple-test", generateIDFromString("a-simple-test"))
	require.Equal(t, "bitrise-step-a-simple-test", generateIDFromString("a_simple_test"))
	require.Equal(t, "bitrise-step-a-simple-test", generateIDFromString("A simple Test"))
	require.Equal(t, "bitrise-step-a-simple-test", generateIDFromString("A simple.Test"))
	require.Equal(t, "bitrise-step-a-simple-test", generateIDFromString("A simple.Test."))
	require.Equal(t, "bitrise-step-a-simple-test", generateIDFromString("--A simple.Test.    --"))
	require.Equal(t, "bitrise-step-a-simple-test", generateIDFromString("    --A simple.Test.    --      "))
	require.Equal(t, "bitrise-step-a--simple-test", generateIDFromString("A  simple Test"))
}
