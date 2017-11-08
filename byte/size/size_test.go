package size

import (
	"testing"

	"encoding/json"

	"github.com/cuigh/auxo/test/assert"
)

func TestParse(t *testing.T) {
	testCases := map[string]uint64{
		"2B":   2,
		"2 B":  2,
		"2  B": 2,
		"2KB":  2 << 10,
		"2K":   2 << 10,
		"2MB":  2 << 20,
		"2M":   2 << 20,
		"2GB":  2 << 30,
		"2G":   2 << 30,
		"2TB":  2 << 40,
		"2T":   2 << 40,
		"2PB":  2 << 50,
		"2P":   2 << 50,
		"2EB":  2 << 60,
		"2E":   2 << 60,
	}
	for v, r := range testCases {
		s, err := Parse(v)
		assert.NoError(t, err)
		assert.Equal(t, r, uint64(s))
	}
}

func TestSizeString(t *testing.T) {
	testCases := map[uint64]string{
		2:       "2 B",
		1025:    "1 KB",
		1536:    "1.5 KB",
		2 << 10: "2 KB",
		2 << 20: "2 MB",
		2 << 30: "2 GB",
		2 << 40: "2 TB",
		2 << 50: "2 PB",
		2 << 60: "2 EB",
	}
	for v, r := range testCases {
		s := Size(v)
		assert.Equal(t, r, s.String())
	}
}

func TestSizeMarshalJSON(t *testing.T) {
	const input = Size(2048)
	const expected = `"2 KB"`

	actual, err := json.Marshal(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(actual))
}

func TestSizeUnmarshalJSON(t *testing.T) {
	const input = `"2 KB"`
	const expected = Size(2048)

	var s Size
	err := json.Unmarshal([]byte(input), &s)
	assert.NoError(t, err)
	assert.Equal(t, expected, s)
}

func TestSizeUnmarshalOption(t *testing.T) {
	const input = "2 K"
	const expected = Size(2048)

	var s Size
	s.UnmarshalOption(input)
	assert.Equal(t, expected, s)
}
