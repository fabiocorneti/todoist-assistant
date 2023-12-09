package todoist

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindProjectId(t *testing.T) {
	client := Client{}

	type testCase struct {
		name           string
		projects       []Project
		projectName    string
		expectedError  error
		expectedResult string
	}

	projects := []Project{
		{
			ID:   "1",
			Name: "Foo",
		},
		{
			ID:   "2",
			Name: "Bar",
		},
		{
			ID:       "3",
			ParentID: "2",
			Name:     "Baz",
		},
		{
			ID:       "4",
			ParentID: "3",
			Name:     "Bar",
		},
	}

	testCases := []testCase{
		{
			name:           "Single match",
			projects:       projects,
			projectName:    "Foo",
			expectedError:  nil,
			expectedResult: "1",
		},
		{
			name:           "Multiple matches",
			projects:       projects,
			projectName:    "Bar",
			expectedError:  errors.New("found more than one project for name Bar"),
			expectedResult: "",
		},
		{
			name:           "No matches",
			projects:       projects,
			projectName:    "Far",
			expectedError:  errors.New("project Far not found"),
			expectedResult: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.FindProjectID(tc.projects, tc.projectName)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
