package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/domain/model"
	mock_controller "github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/infra/controller/mock"
	"go.uber.org/mock/gomock"
)

func TestRouter_HandleRoot(t *testing.T) {
	tests := []struct {
		name                string
		expectedStatusCode  int
		expectedMessage     string
		expectedStatus      string
		expectedContentType string
	}{
		{
			name:                "正常系: ルートエンドポイントが正しく応答",
			expectedStatusCode:  http.StatusOK,
			expectedMessage:     "Japan Tech Careers API is running",
			expectedStatus:      "healthy",
			expectedContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockController := mock_controller.NewMockController(ctrl)
			router := NewRouter(mockController)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			var response map[string]string
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response["message"] != tt.expectedMessage {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMessage, response["message"])
			}

			if response["status"] != tt.expectedStatus {
				t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, response["status"])
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != tt.expectedContentType {
				t.Errorf("Expected Content-Type '%s', got '%s'", tt.expectedContentType, contentType)
			}
		})
	}
}

func TestRouter_HandleGetJobs(t *testing.T) {
	tests := []struct {
		name                  string
		mockSetup             func(*mock_controller.MockController)
		expectedStatusCode    int
		expectedCount         int
		expectedError         string
		checkFirstJobTitle    string
		checkFirstJobLocation string
		isErrorResponse       bool
	}{
		{
			name: "正常系: 複数のJobが返される",
			mockSetup: func(m *mock_controller.MockController) {
				m.EXPECT().GetJobs(gomock.Any()).Return([]model.Job{
					{
						ID:          "1",
						Title:       "Senior Go Developer",
						Company:     "Tech Company",
						Location:    "Tokyo",
						Description: "Great opportunity",
					},
					{
						ID:          "2",
						Title:       "Backend Engineer",
						Company:     "Startup",
						Location:    "Osaka",
						Description: "Exciting role",
					},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedCount:      2,
			checkFirstJobTitle: "Senior Go Developer",
			isErrorResponse:    false,
		},
		{
			name: "正常系: 空のJobリストが返される",
			mockSetup: func(m *mock_controller.MockController) {
				m.EXPECT().GetJobs(gomock.Any()).Return([]model.Job{}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedCount:      0,
			isErrorResponse:    false,
		},
		{
			name: "異常系: Controllerからエラーが返される",
			mockSetup: func(m *mock_controller.MockController) {
				m.EXPECT().GetJobs(gomock.Any()).Return(nil, errors.New("database connection failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "Failed to fetch jobs",
			isErrorResponse:    true,
		},
		{
			name: "正常系: 1件のJobが返される",
			mockSetup: func(m *mock_controller.MockController) {
				m.EXPECT().GetJobs(gomock.Any()).Return([]model.Job{
					{
						ID:          "99",
						Title:       "Single Job",
						Company:     "Single Company",
						Location:    "Fukuoka",
						Description: "Single description",
					},
				}, nil)
			},
			expectedStatusCode:    http.StatusOK,
			expectedCount:         1,
			checkFirstJobLocation: "Fukuoka",
			isErrorResponse:       false,
		},
		{
			name: "異常系: サービスタイムアウトエラー",
			mockSetup: func(m *mock_controller.MockController) {
				m.EXPECT().GetJobs(gomock.Any()).Return(nil, errors.New("request timeout"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "Failed to fetch jobs",
			isErrorResponse:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockController := mock_controller.NewMockController(ctrl)
			tt.mockSetup(mockController)

			router := NewRouter(mockController)
			req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert: ステータスコードの検証
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			// Assert: Content-Typeの検証
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Assert: レスポンスボディの検証
			if tt.isErrorResponse {
				var response map[string]string
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedError, response["error"])
				}
			} else {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				count, ok := response["count"].(float64)
				if !ok {
					t.Fatal("Expected 'count' field in response")
				}
				if int(count) != tt.expectedCount {
					t.Errorf("Expected count %d, got %d", tt.expectedCount, int(count))
				}

				jobs, ok := response["jobs"].([]interface{})
				if !ok {
					t.Fatal("Expected 'jobs' field in response")
				}
				if len(jobs) != tt.expectedCount {
					t.Errorf("Expected %d jobs, got %d", tt.expectedCount, len(jobs))
				}

				// 最初のJobの検証（指定されている場合のみ）
				if tt.checkFirstJobTitle != "" && len(jobs) > 0 {
					firstJob, ok := jobs[0].(map[string]interface{})
					if !ok {
						t.Fatal("Expected first job to be a map")
					}
					if firstJob["title"] != tt.checkFirstJobTitle {
						t.Errorf("Expected first job title '%s', got '%s'", tt.checkFirstJobTitle, firstJob["title"])
					}
				}

				if tt.checkFirstJobLocation != "" && len(jobs) > 0 {
					firstJob, ok := jobs[0].(map[string]interface{})
					if !ok {
						t.Fatal("Expected first job to be a map")
					}
					if firstJob["location"] != tt.checkFirstJobLocation {
						t.Errorf("Expected first job location '%s', got '%s'", tt.checkFirstJobLocation, firstJob["location"])
					}
				}
			}
		})
	}
}
