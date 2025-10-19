package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/domain/model"
	mock_service "github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/domain/service/mock"
	"go.uber.org/mock/gomock"
)

func TestControllerImpl_GetJobs(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*mock_service.MockService)
		expectedJobs  []model.Job
		expectedError string
		checkJobsNil  bool
	}{
		{
			name: "Success: Multiple jobs are returned",
			mockSetup: func(m *mock_service.MockService) {
				m.EXPECT().FetchJobs(gomock.Any()).Return([]model.Job{
					{
						ID:          "1",
						Title:       "Test Job 1",
						Company:     "Test Company",
						Location:    "Tokyo",
						Description: "Test Description",
					},
					{
						ID:          "2",
						Title:       "Test Job 2",
						Company:     "Another Company",
						Location:    "Osaka",
						Description: "Another Description",
					},
				}, nil)
			},
			expectedJobs: []model.Job{
				{
					ID:          "1",
					Title:       "Test Job 1",
					Company:     "Test Company",
					Location:    "Tokyo",
					Description: "Test Description",
				},
				{
					ID:          "2",
					Title:       "Test Job 2",
					Company:     "Another Company",
					Location:    "Osaka",
					Description: "Another Description",
				},
			},
			expectedError: "",
			checkJobsNil:  false,
		},
		{
			name: "Success: Empty job list is returned",
			mockSetup: func(m *mock_service.MockService) {
				m.EXPECT().FetchJobs(gomock.Any()).Return([]model.Job{}, nil)
			},
			expectedJobs:  []model.Job{},
			expectedError: "",
			checkJobsNil:  false,
		},
		{
			name: "Error: Service returns error",
			mockSetup: func(m *mock_service.MockService) {
				m.EXPECT().FetchJobs(gomock.Any()).Return(nil, errors.New("service error occurred"))
			},
			expectedJobs:  nil,
			expectedError: "service error occurred",
			checkJobsNil:  true,
		},
		{
			name: "Success: Single job is returned",
			mockSetup: func(m *mock_service.MockService) {
				m.EXPECT().FetchJobs(gomock.Any()).Return([]model.Job{
					{
						ID:          "100",
						Title:       "Single Job",
						Company:     "Single Company",
						Location:    "Nagoya",
						Description: "Single description",
					},
				}, nil)
			},
			expectedJobs: []model.Job{
				{
					ID:          "100",
					Title:       "Single Job",
					Company:     "Single Company",
					Location:    "Nagoya",
					Description: "Single description",
				},
			},
			expectedError: "",
			checkJobsNil:  false,
		},
		{
			name: "Error: Database connection error",
			mockSetup: func(m *mock_service.MockService) {
				m.EXPECT().FetchJobs(gomock.Any()).Return(nil, errors.New("database connection failed"))
			},
			expectedJobs:  nil,
			expectedError: "database connection failed",
			checkJobsNil:  true,
		},
		{
			name: "Success: Large number of jobs are returned",
			mockSetup: func(m *mock_service.MockService) {
				manyJobs := make([]model.Job, 100)
				for i := 0; i < 100; i++ {
					manyJobs[i] = model.Job{
						ID:          string(rune(i)),
						Title:       "Job Title",
						Company:     "Company",
						Location:    "Location",
						Description: "Description",
					}
				}
				m.EXPECT().FetchJobs(gomock.Any()).Return(manyJobs, nil)
			},
			expectedJobs: func() []model.Job {
				manyJobs := make([]model.Job, 100)
				for i := 0; i < 100; i++ {
					manyJobs[i] = model.Job{
						ID:          string(rune(i)),
						Title:       "Job Title",
						Company:     "Company",
						Location:    "Location",
						Description: "Description",
					}
				}
				return manyJobs
			}(),
			expectedError: "",
			checkJobsNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: gomockのコントローラとモックを作成
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			tt.mockSetup(mockService)

			controller := NewController(mockService)
			ctx := context.Background()

			// Act: テスト対象のメソッドを実行
			jobs, err := controller.GetJobs(ctx)

			// Assert: エラーの検証
			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("Expected error '%s', got nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}

			// Assert: Jobsの検証
			if tt.checkJobsNil {
				if jobs != nil {
					t.Errorf("Expected nil jobs, got %v", jobs)
				}
			} else {
				if len(jobs) != len(tt.expectedJobs) {
					t.Fatalf("Expected %d jobs, got %d", len(tt.expectedJobs), len(jobs))
				}
				for i, expectedJob := range tt.expectedJobs {
					if jobs[i] != expectedJob {
						t.Errorf("Job[%d] mismatch:\n  expected: %+v\n  got:      %+v", i, expectedJob, jobs[i])
					}
				}
			}
		})
	}
}
