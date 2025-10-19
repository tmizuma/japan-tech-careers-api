package service

import (
	"context"
	"errors"
	"testing"

	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/domain/model"
	mock_httpclient "github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/infra/httpclient/mock"
	"go.uber.org/mock/gomock"
)

func TestServiceImpl_FetchJobs(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*mock_httpclient.MockHttpClient)
		expectedJobs  []model.Job
		expectedError string
		checkJobsNil  bool
	}{
		{
			name: "正常系: 複数のJobが返される",
			mockSetup: func(m *mock_httpclient.MockHttpClient) {
				m.EXPECT().GetJobs(gomock.Any()).Return([]model.Job{
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
			name: "正常系: 空のJobリストが返される",
			mockSetup: func(m *mock_httpclient.MockHttpClient) {
				m.EXPECT().GetJobs(gomock.Any()).Return([]model.Job{}, nil)
			},
			expectedJobs:  []model.Job{},
			expectedError: "",
			checkJobsNil:  false,
		},
		{
			name: "異常系: HttpClientからエラーが返される",
			mockSetup: func(m *mock_httpclient.MockHttpClient) {
				m.EXPECT().GetJobs(gomock.Any()).Return(nil, errors.New("failed to fetch jobs from external API"))
			},
			expectedJobs:  nil,
			expectedError: "failed to fetch jobs from external API",
			checkJobsNil:  true,
		},
		{
			name: "異常系: ネットワークエラー",
			mockSetup: func(m *mock_httpclient.MockHttpClient) {
				m.EXPECT().GetJobs(gomock.Any()).Return(nil, errors.New("network timeout"))
			},
			expectedJobs:  nil,
			expectedError: "network timeout",
			checkJobsNil:  true,
		},
		{
			name: "正常系: 1件のJobが返される",
			mockSetup: func(m *mock_httpclient.MockHttpClient) {
				m.EXPECT().GetJobs(gomock.Any()).Return([]model.Job{
					{
						ID:          "100",
						Title:       "Single Job",
						Company:     "Single Company",
						Location:    "Fukuoka",
						Description: "Single job description",
					},
				}, nil)
			},
			expectedJobs: []model.Job{
				{
					ID:          "100",
					Title:       "Single Job",
					Company:     "Single Company",
					Location:    "Fukuoka",
					Description: "Single job description",
				},
			},
			expectedError: "",
			checkJobsNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: gomockのコントローラとモックを作成
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock_httpclient.NewMockHttpClient(ctrl)
			tt.mockSetup(mockClient)

			svc := NewServiceImpl(mockClient)
			ctx := context.Background()

			// Act: テスト対象のメソッドを実行
			jobs, err := svc.FetchJobs(ctx)

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
