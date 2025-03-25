package models_test

import (
	"reflect"
	"testing"

	"github.com/andrejsstepanovs/andai/internal/employee/processor/models"
)

func TestAnswer_GetDeps(t *testing.T) {
	tests := []struct {
		name     string
		answer   models.Answer
		expected map[int][]int
	}{
		{
			name:     "empty issues",
			answer:   models.Answer{Issues: []models.AnswerIssues{}},
			expected: map[int][]int{},
		},
		{
			name: "single issue with no dependencies",
			answer: models.Answer{
				Issues: []models.AnswerIssues{
					{ID: 1},
				},
			},
			expected: map[int][]int{
				1: {},
			},
		},
		{
			name: "single issue with multiple blocked_by entries",
			answer: models.Answer{
				Issues: []models.AnswerIssues{
					{ID: 2, BlockedBy: []int{3, 4}},
				},
			},
			expected: map[int][]int{
				2: {3, 4},
			},
		},
		{
			name: "multiple issues with various dependencies",
			answer: models.Answer{
				Issues: []models.AnswerIssues{
					{ID: 1, BlockedBy: []int{5}},
					{ID: 2},
					{ID: 3, BlockedBy: []int{1, 2}},
				},
			},
			expected: map[int][]int{
				1: {5},
				2: {},
				3: {1, 2},
			},
		},
		{
			name: "duplicate blocked_by entries",
			answer: models.Answer{
				Issues: []models.AnswerIssues{
					{ID: 4, BlockedBy: []int{5, 5}},
				},
			},
			expected: map[int][]int{
				4: {5, 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.answer.GetDeps()
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetDeps() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateNoSelfReference(t *testing.T) {
	tests := []struct {
		name    string
		answer  models.Answer
		wantErr bool
	}{
		{
			name:    "no issues",
			answer:  models.Answer{Issues: []models.AnswerIssues{}},
			wantErr: false,
		},
		{
			name: "no dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1},
				{ID: 2},
			}},
			wantErr: false,
		},
		{
			name: "no self-dependency",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2},
			}},
			wantErr: false,
		},
		{
			name: "self-dependency",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{1}},
			}},
			wantErr: true,
		},
		{
			name: "multiple self-dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{1, 2}},
				{ID: 2, BlockedBy: []int{2}},
			}},
			wantErr: true,
		},
		{
			name: "self-dependency with other dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{1, 2}},
				{ID: 2},
			}},
			wantErr: true,
		},
		{
			name: "self-dependency with other dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{1, 2}},
				{ID: 2},
			}},
			wantErr: true,
		},
		{
			name: "self-dependency with multiple dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{1, 2, 3}},
				{ID: 2},
				{ID: 3},
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.answer.ValidateNoSelfReference()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDependentOnSelf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDependenciesExist(t *testing.T) {
	tests := []struct {
		name    string
		answer  models.Answer
		wantErr bool
	}{
		{
			name:    "no issues",
			answer:  models.Answer{Issues: []models.AnswerIssues{}},
			wantErr: false,
		},
		{
			name: "no dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1},
				{ID: 2},
			}},
			wantErr: false,
		},
		{
			name: "all dependencies exist",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2},
			}},
			wantErr: false,
		},
		{
			name: "dependency on non-existent issue",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
			}},
			wantErr: true,
		},
		{
			name: "multiple missing dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2, 3}},
				{ID: 3, BlockedBy: []int{4}},
			}},
			wantErr: true,
		},
		{
			name: "self dependency exists",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{1}},
			}},
			wantErr: false,
		},
		{
			name: "zero ID exists",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 0, BlockedBy: []int{0}},
			}},
			wantErr: false,
		},
		{
			name: "negative ID exists",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: -1, BlockedBy: []int{-2}},
				{ID: -2},
			}},
			wantErr: false,
		},
		{
			name: "negative ID missing",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: -1, BlockedBy: []int{-2}},
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.answer.ValidateDependenciesExist()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDependenciesExist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCircularDependency(t *testing.T) {
	tests := []struct {
		name    string
		answer  models.Answer
		wantErr bool
	}{
		{
			name:    "no issues",
			answer:  models.Answer{Issues: []models.AnswerIssues{}},
			wantErr: false,
		},
		{
			name: "no dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1},
				{ID: 2},
			}},
			wantErr: false,
		},
		{
			name: "self dependency",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{1}},
			}},
			wantErr: true,
		},
		{
			name: "two-issue cycle",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2, BlockedBy: []int{1}},
			}},
			wantErr: true,
		},
		{
			name: "three-issue cycle",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2, BlockedBy: []int{3}},
				{ID: 3, BlockedBy: []int{1}},
			}},
			wantErr: true,
		},
		{
			name: "no cycle with multiple dependencies",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2, 3}},
				{ID: 2},
				{ID: 3},
			}},
			wantErr: false,
		},
		{
			name: "cycle in subgraph",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2, BlockedBy: []int{3}},
				{ID: 3, BlockedBy: []int{2}},
				{ID: 4},
			}},
			wantErr: true,
		},
		{
			name: "chain without cycle",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2, BlockedBy: []int{3}},
				{ID: 3},
			}},
			wantErr: false,
		},
		{
			name: "negative ID cycle",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: -1, BlockedBy: []int{-2}},
				{ID: -2, BlockedBy: []int{-1}},
			}},
			wantErr: true,
		},
		{
			name: "zero ID cycle",
			answer: models.Answer{Issues: []models.AnswerIssues{
				{ID: 0, BlockedBy: []int{0}},
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.answer.ValidateCircularDependency()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCircularDependency() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
