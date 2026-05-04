package apiclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFormatItemID tests the FormatItemID function
func TestFormatItemID(t *testing.T) {

	t.Run("Standard 32-character lowercase ID", func(t *testing.T) {
		input := "87f82eeb362b4923900176b29b448600"
		expected := "{87F82EEB-362B-4923-9001-76B29B448600}"
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})

	t.Run("Standard 32-character uppercase ID", func(t *testing.T) {
		input := "87F82EEB362B4923900176B29B448600"
		expected := "{87F82EEB-362B-4923-9001-76B29B448600}"
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})

	t.Run("ID with braces but no dashes", func(t *testing.T) {
		input := "{87F82EEB362B4923900176B29B448600}"
		expected := "{87F82EEB-362B-4923-9001-76B29B448600}"
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})

	t.Run("Already formatted ID", func(t *testing.T) {
		input := "{87F82EEB-362B-4923-9001-76B29B448600}"
		expected := "{87F82EEB-362B-4923-9001-76B29B448600}"
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})

	t.Run("Non-standard ID (not 32 hex chars)", func(t *testing.T) {
		input := "test-id"
		expected := "test-id"
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})

	t.Run("Empty ID", func(t *testing.T) {
		input := ""
		expected := ""
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})

	t.Run("ID with non-hex characters", func(t *testing.T) {
		input := "a58aab49fe074gt5b03f927c581e74d7"
		expected := "a58aab49fe074gt5b03f927c581e74d7"
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})

	t.Run("ID that's too short", func(t *testing.T) {
		input := "87f82eeb362b4923900176b29b4486"
		expected := "87f82eeb362b4923900176b29b4486"
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})

	t.Run("ID that's too long", func(t *testing.T) {
		input := "87f82eeb362b4923900176b29b448600123"
		expected := "87f82eeb362b4923900176b29b448600123"
		actual := FormatItemID(input)
		assert.Equal(t, expected, actual)
	})
}
