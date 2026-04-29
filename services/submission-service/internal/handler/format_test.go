package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router

}

func TestHealthCheck(t *testing.T) {
	router := setupRouter()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}

func TestFormatCheckHandler(t *testing.T) {
	handler := NewFormatCheckHandler()

	router := setupRouter()
	router.POST("/api/submissions/check-format", handler.CheckFormat)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedValid  bool
	}{
		{
			name: "Valid submission",
			body: map[string]interface{}{
				"title":    "A Study on Machine Learning Applications in Healthcare",
				"abstract": "This paper explores the applications of machine learning in healthcare, focusing on diagnostic imaging, drug discovery, and personalized treatment plans. We review recent advances and discuss future directions.",
				"keywords": []string{"machine learning", "healthcare", "artificial intelligence"},
				"content":  "Introduction: Machine learning has revolutionized many industries. Methods: We conducted a comprehensive review. Results: ML shows promise in diagnostics. Discussion: Further research is needed. Conclusion: ML will transform healthcare.",
			},
			expectedStatus: 200,
			expectedValid:  true,
		},
		{
			name: "Short title",
			body: map[string]interface{}{
				"title":    "Short",
				"abstract": "This is a test abstract with enough characters to pass the minimum length requirement for the format checker validation.",
				"keywords": []string{"test", "example", "sample"},
				"content":  "Introduction content here. Methods content here. Results content here. Discussion content here. Conclusion content here.",
			},
			expectedStatus: 200,
			expectedValid:  true,
		},
		{
			name: "Missing required fields",
			body: map[string]interface{}{
				"abstract": "Only abstract provided",
			},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/submissions/check-format", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == 200 {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				data := response["data"].(map[string]interface{})
				if data["valid"] != tt.expectedValid {
					t.Errorf("Expected valid=%v, got %v", tt.expectedValid, data["valid"])
				}
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"  hello  world  ", 2},
		{"one two three four five", 5},
	}

	for _, tt := range tests {
		result := countWords(tt.input)
		if result != tt.expected {
			t.Errorf("countWords(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestCheckDocumentStructure(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Has full structure",
			content:  "Introduction: This is the intro. Methods: We used these methods. Results: Here are the results. Discussion: Let's discuss. Conclusion: Final thoughts.",
			expected: true,
		},
		{
			name:     "Partial structure",
			content:  "Introduction: This is the intro. Methods: We used these methods.",
			expected: false,
		},
		{
			name:     "Empty content",
			content:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkDocumentStructure(tt.content)
			if result != tt.expected {
				t.Errorf("checkDocumentStructure(%q) = %v, want %v", tt.content, result, tt.expected)
			}
		})
	}
}
