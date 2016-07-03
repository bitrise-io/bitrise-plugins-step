package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndentTextWithMaxLength(t *testing.T) {
	t.Log("Empty")
	{
		input := ""
		output := IndentTextWithMaxLength(input, 0, 80)
		require.Equal(t, "", output)
	}

	t.Log("One liner")
	{
		input := "one liner"
		output := IndentTextWithMaxLength(input, 0, 80)
		require.Equal(t, "one liner", output)
	}

	t.Log("One liner - with indent")
	{
		input := "one liner"
		output := IndentTextWithMaxLength(input, 3, 80)
		require.Equal(t, "   one liner", output)
	}

	t.Log("One liner - max width")
	{
		input := "one"
		output := IndentTextWithMaxLength(input, 0, 3)
		require.Equal(t, "one", output)
	}

	t.Log("One liner - longer than max width")
	{
		input := "onetwo"
		output := IndentTextWithMaxLength(input, 0, 3)
		require.Equal(t, "one\ntwo", output)
	}

	t.Log("One liner - max width - with indent")
	{
		input := "one"
		output := IndentTextWithMaxLength(input, 1, 3)
		require.Equal(t, " on\n e", output)
	}

	t.Log("One liner - longer than max width - with indent")
	{
		input := "onetwo"
		output := IndentTextWithMaxLength(input, 1, 3)
		require.Equal(t, " on\n et\n wo", output)
	}

	t.Log("Two lines, shorter than max")
	{
		input := `first line
second line`
		output := IndentTextWithMaxLength(input, 0, 80)
		require.Equal(t, `first line
second line`, output)
	}

	t.Log("Two lines, shorter than max - with indent")
	{
		input := `first line
second line`
		output := IndentTextWithMaxLength(input, 2, 80)
		require.Equal(t, `  first line
  second line`, output)
	}

	t.Log("Two lines, longer than max")
	{
		input := `firstline
secondline`
		output := IndentTextWithMaxLength(input, 0, 5)
		require.Equal(t, `first
line
secon
dline`, output)
	}
}
