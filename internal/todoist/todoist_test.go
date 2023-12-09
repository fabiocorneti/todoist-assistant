package todoist

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveLabelsFromTask(t *testing.T) {
	testCases := []struct {
		name           string
		taskID         string
		currentLabels  []string
		labelsToRemove []string
		expectedLabels []string
		expectError    bool
	}{
		{
			name:           "Remove single Label",
			taskID:         "1",
			currentLabels:  []string{"urgent", "work", "personal"},
			labelsToRemove: []string{"personal"},
			expectedLabels: []string{"urgent", "work"},
			expectError:    false,
		},
		{
			name:           "Remove multiple Labels",
			taskID:         "2",
			currentLabels:  []string{"urgent", "work", "personal"},
			labelsToRemove: []string{"urgent", "personal"},
			expectedLabels: []string{"work"},
			expectError:    false,
		},
		{
			name:           "Remove non existent Label",
			taskID:         "3",
			currentLabels:  []string{"urgent", "work"},
			labelsToRemove: []string{"home"},
			expectedLabels: []string{"urgent", "work"},
			expectError:    false,
		},
		{
			name:           "Error retrieving Labels",
			taskID:         "4",
			currentLabels:  nil,
			labelsToRemove: []string{"urgent"},
			expectedLabels: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockTransport := MockTransport{}
			client := Client{
				transport: &mockTransport,
			}

			if tc.expectError {
				mockTransport.On("getTaskLabels", tc.taskID).Return(nil, errors.New("Could not get labels"))
			} else {
				mockTransport.On("getTaskLabels", tc.taskID).Return(tc.currentLabels, nil)
				mockTransport.On("updateTaskLabels", tc.taskID, tc.expectedLabels).Return(nil)
			}

			err := client.RemoveLabelsFromTask(tc.taskID, tc.labelsToRemove)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockTransport.AssertExpectations(t)
		})
	}
}

func TestAddLabelsToTask(t *testing.T) {
	testCases := []struct {
		name           string
		taskID         string
		existingLabels []string
		labelsToAdd    []string
		expectedLabels []string
		expectError    bool
	}{
		{
			name:           "Add new Labels",
			taskID:         "1",
			existingLabels: []string{"urgent", "work"},
			labelsToAdd:    []string{"personal", "home"},
			expectedLabels: []string{"urgent", "work", "personal", "home"},
			expectError:    false,
		},
		{
			name:           "Add existing Label",
			taskID:         "2",
			existingLabels: []string{"urgent", "work"},
			labelsToAdd:    []string{"work"},
			expectedLabels: []string{"urgent", "work"},
			expectError:    false,
		},
		{
			name:           "Error retrieving Labels",
			taskID:         "3",
			existingLabels: nil,
			labelsToAdd:    []string{"urgent"},
			expectedLabels: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockTransport := MockTransport{}
			client := Client{
				transport: &mockTransport,
			}

			if tc.expectError {
				mockTransport.On("getTaskLabels", tc.taskID).Return(nil, errors.New("error"))
			} else {
				mockTransport.On("getTaskLabels", tc.taskID).Return(tc.existingLabels, nil)
				mockTransport.On("updateTaskLabels", tc.taskID, tc.expectedLabels).Return(nil)
			}

			err := client.AddLabelsToTask(tc.taskID, tc.labelsToAdd)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockTransport.AssertCalled(t, "updateTaskLabels", tc.taskID, tc.expectedLabels)
			}

			mockTransport.AssertExpectations(t)
		})
	}
}
