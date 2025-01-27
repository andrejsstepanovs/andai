package processor_test

import (
	"testing"

	"github.com/andrejsstepanovs/andai/pkg/employee/processor"
)

func TestValidateDependenciesExist(t *testing.T) {
	tests := []struct {
		name    string
		answer  processor.Answer
		wantErr bool
	}{
		{
			name:    "no issues",
			answer:  processor.Answer{Issues: []processor.AnswerIssues{}},
			wantErr: false,
		},
		{
			name: "no dependencies",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1},
				{ID: 2},
			}},
			wantErr: false,
		},
		{
			name: "all dependencies exist",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2},
			}},
			wantErr: false,
		},
		{
			name: "dependency on non-existent issue",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
			}},
			wantErr: true,
		},
		{
			name: "multiple missing dependencies",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{2, 3}},
				{ID: 3, BlockedBy: []int{4}},
			}},
			wantErr: true,
		},
		{
			name: "self dependency exists",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{1}},
			}},
			wantErr: false,
		},
		{
			name: "zero ID exists",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 0, BlockedBy: []int{0}},
			}},
			wantErr: false,
		},
		{
			name: "negative ID exists",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: -1, BlockedBy: []int{-2}},
				{ID: -2},
			}},
			wantErr: false,
		},
		{
			name: "negative ID missing",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
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
		answer  processor.Answer
		wantErr bool
	}{
		{
			name:    "no issues",
			answer:  processor.Answer{Issues: []processor.AnswerIssues{}},
			wantErr: false,
		},
		{
			name: "no dependencies",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1},
				{ID: 2},
			}},
			wantErr: false,
		},
		{
			name: "self dependency",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{1}},
			}},
			wantErr: true,
		},
		{
			name: "two-issue cycle",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2, BlockedBy: []int{1}},
			}},
			wantErr: true,
		},
		{
			name: "three-issue cycle",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2, BlockedBy: []int{3}},
				{ID: 3, BlockedBy: []int{1}},
			}},
			wantErr: true,
		},
		{
			name: "no cycle with multiple dependencies",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{2, 3}},
				{ID: 2},
				{ID: 3},
			}},
			wantErr: false,
		},
		{
			name: "cycle in subgraph",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2, BlockedBy: []int{3}},
				{ID: 3, BlockedBy: []int{2}},
				{ID: 4},
			}},
			wantErr: true,
		},
		{
			name: "chain without cycle",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: 1, BlockedBy: []int{2}},
				{ID: 2, BlockedBy: []int{3}},
				{ID: 3},
			}},
			wantErr: false,
		},
		{
			name: "negative ID cycle",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
				{ID: -1, BlockedBy: []int{-2}},
				{ID: -2, BlockedBy: []int{-1}},
			}},
			wantErr: true,
		},
		{
			name: "zero ID cycle",
			answer: processor.Answer{Issues: []processor.AnswerIssues{
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
