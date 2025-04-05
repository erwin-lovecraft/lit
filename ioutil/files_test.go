package ioutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit/testutil"
)

func TestReadFile(t *testing.T) {
	SetResourceDir("testdata")

	tcs := map[string]struct {
		fileName  string
		expResult string
		expErr    error
	}{
		"success": {
			fileName: "bugdiary.txt",
			expResult: `// Bug Logger's Diary

Day 1: Found a bug. Easy fix.
Day 2: Bug still there. Not so easy.
Day 3: Bug seems sentient. Considering negotiation.
Day 4: The bug has a family. Can't delete now.
Day 5: The bug filed a PR to fix itself. Merged.

TEST COVERAGE:
Expectations: 100%
Reality: 42%
Excuses: 100%

FINAL REPORT:
Unit tests passed: YES*
*Definition of "passed" may vary

// TODO: Delete this file before code review
// TODO: Remember to delete this TODO
`,
		},
		"file not found": {
			fileName: "nonexistent.txt",
			expErr:   errors.New("open testdata/nonexistent.txt: no such file or directory"),
		},
	}

	for scenario, tc := range tcs {
		tc := tc
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()

			// GIVEN
			// WHEN
			rs, err := ReadFile(tc.fileName)

			// THEN
			if tc.expErr != nil {
				require.EqualError(t, err, tc.expErr.Error())
			} else {
				require.NoError(t, err)
				testutil.Equal(t, tc.expResult, string(rs))
			}
		})
	}
}
