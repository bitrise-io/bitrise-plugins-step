package create

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateIDFromString(t *testing.T) {
	require.Equal(t, "a-simple-test", generateIDFromString("a-simple-test"))
	require.Equal(t, "a-simple-test", generateIDFromString("a_simple_test"))
	require.Equal(t, "a-simple-test", generateIDFromString("A simple Test"))
	require.Equal(t, "a-simple-test", generateIDFromString("A simple.Test"))
	require.Equal(t, "a-simple-test", generateIDFromString("A simple.Test."))
	require.Equal(t, "a-simple-test", generateIDFromString("--A simple.Test.    --"))
	require.Equal(t, "a-simple-test", generateIDFromString("    --A simple.Test.    --      "))
	require.Equal(t, "a--simple-test", generateIDFromString("A  simple Test"))
	//
	require.Equal(t, "a-simple-test-2", generateIDFromString("A simple test 2"))
}

func Test_evaluateTemplate(t *testing.T) {
	const (
		invAuthor         = "UT Author"
		invTitle          = "UT Test Step"
		invID             = "ut-test-step"
		invSummary        = "UT summary"
		invDescription    = "UT description."
		invPrimaryTypeTag = "test"
	)

	t.Log("Go toolkit - step.yml template test")
	{
		evaluatedContent, err := evaluateTemplate("step.yml.gotemplate", InventoryModel{
			Author:         invAuthor,
			Title:          invTitle,
			ID:             invID,
			Summary:        invSummary,
			Description:    invDescription,
			PrimaryTypeTag: invPrimaryTypeTag,
			Year:           2017,
			//
			ToolkitType: toolkitTypeGo,
			GoToolkitInventory: GoToolkitInventoryModel{
				PackageID: "github.com/bitrise-io/bitrise-step-" + invID,
			},
		})
		require.NoError(t, err)
		require.Contains(t, evaluatedContent, `toolkit:
  go:
    package_name: github.com/bitrise-io/bitrise-step-ut-test-step`)
	}
}
