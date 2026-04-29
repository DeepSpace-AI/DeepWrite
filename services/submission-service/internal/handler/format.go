package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FormatCheckHandler struct{}

func NewFormatCheckHandler() *FormatCheckHandler {
	return &FormatCheckHandler{}
}

type FormatCheckRequest struct {
	Title    string   `json:"title" binding:"required"`
	Abstract string   `json:"abstract"`
	Keywords []string `json:"keywords"`
	Content  string   `json:"content"`
	JournalID int     `json:"journal_id"`
}

type FormatCheckResult struct {
	Valid    bool              `json:"valid"`
	Score    int               `json:"score"`
	Checks   []FormatCheckItem `json:"checks"`
	Summary  string            `json:"summary"`
}

type FormatCheckItem struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

func (h *FormatCheckHandler) CheckFormat(c *gin.Context) {
	var req FormatCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_INPUT", "message": err.Error()}})
		return
	}

	checks := []FormatCheckItem{}
	score := 100

	if len(req.Title) < 10 {
		checks = append(checks, FormatCheckItem{
			Name:     "标题长度",
			Status:   "warning",
			Message:  "标题过短，建议至少10个字符",
			Severity: "medium",
		})
		score -= 10
	} else if len(req.Title) > 200 {
		checks = append(checks, FormatCheckItem{
			Name:     "标题长度",
			Status:   "warning",
			Message:  "标题过长，建议不超过200个字符",
			Severity: "medium",
		})
		score -= 5
	} else {
		checks = append(checks, FormatCheckItem{
			Name:     "标题长度",
			Status:   "pass",
			Message:  "标题长度符合要求",
			Severity: "none",
		})
	}

	if len(req.Abstract) < 100 {
		checks = append(checks, FormatCheckItem{
			Name:     "摘要长度",
			Status:   "warning",
			Message:  "摘要过短，建议至少100个字符",
			Severity: "medium",
		})
		score -= 15
	} else if len(req.Abstract) > 3000 {
		checks = append(checks, FormatCheckItem{
			Name:     "摘要长度",
			Status:   "warning",
			Message:  "摘要过长，建议不超过3000个字符",
			Severity: "low",
		})
		score -= 5
	} else {
		checks = append(checks, FormatCheckItem{
			Name:     "摘要长度",
			Status:   "pass",
			Message:  "摘要长度符合要求",
			Severity: "none",
		})
	}

	if len(req.Keywords) < 3 {
		checks = append(checks, FormatCheckItem{
			Name:     "关键词数量",
			Status:   "warning",
			Message:  "关键词过少，建议至少3个",
			Severity: "medium",
		})
		score -= 10
	} else if len(req.Keywords) > 10 {
		checks = append(checks, FormatCheckItem{
			Name:     "关键词数量",
			Status:   "warning",
			Message:  "关键词过多，建议不超过10个",
			Severity: "low",
		})
		score -= 5
	} else {
		checks = append(checks, FormatCheckItem{
			Name:     "关键词数量",
			Status:   "pass",
			Message:  "关键词数量符合要求",
			Severity: "none",
		})
	}

	wordCount := countWords(req.Content)
	if wordCount < 3000 {
		checks = append(checks, FormatCheckItem{
			Name:     "正文长度",
			Status:   "warning",
			Message:  "正文过短，当前约" + formatNumber(wordCount) + "字，建议至少3000字",
			Severity: "high",
		})
		score -= 20
	} else if wordCount > 50000 {
		checks = append(checks, FormatCheckItem{
			Name:     "正文长度",
			Status:   "warning",
			Message:  "正文过长，当前约" + formatNumber(wordCount) + "字，建议不超过50000字",
			Severity: "low",
		})
		score -= 5
	} else {
		checks = append(checks, FormatCheckItem{
			Name:     "正文长度",
			Status:   "pass",
			Message:  "正文长度符合要求，当前约" + formatNumber(wordCount) + "字",
			Severity: "none",
		})
	}

	hasStructure := checkDocumentStructure(req.Content)
	if hasStructure {
		checks = append(checks, FormatCheckItem{
			Name:     "文档结构",
			Status:   "pass",
			Message:  "文档包含标准学术论文结构",
			Severity: "none",
		})
	} else {
		checks = append(checks, FormatCheckItem{
			Name:     "文档结构",
			Status:   "warning",
			Message:  "建议使用标准学术论文结构（引言、方法、结果、讨论）",
			Severity: "medium",
		})
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	summary := generateCheckSummary(score, checks)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": FormatCheckResult{
			Valid:   score >= 60,
			Score:   score,
			Checks:  checks,
			Summary: summary,
		},
	})
}

func countWords(text string) int {
	if text == "" {
		return 0
	}
	count := 0
	inWord := false
	for _, ch := range text {
		if ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r' {
			if inWord {
				count++
				inWord = false
			}
		} else {
			inWord = true
		}
	}
	if inWord {
		count++
	}
	return count
}

func formatNumber(n int) string {
	if n >= 1000 {
		return string(rune(n/1000+'0')) + "," + string(rune((n%1000)/100+'0')) + string(rune((n%100)/10+'0')) + string(rune(n%10+'0'))
	}
	return string(rune(n + '0'))
}

func checkDocumentStructure(content string) bool {
	structureKeywords := []string{
		"introduction", "引言", "背景",
		"method", "方法", "methodology",
		"result", "结果", "findings",
		"discussion", "讨论",
		"conclusion", "结论",
	}

	found := 0
	for _, keyword := range structureKeywords {
		if containsIgnoreCase(content, keyword) {
			found++
		}
	}
	return found >= 3
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func generateCheckSummary(score int, checks []FormatCheckItem) string {
	warnings := 0
	for _, check := range checks {
		if check.Status == "warning" {
			warnings++
		}
	}

	if score >= 90 {
		return "格式检查通过，论文格式非常规范"
	} else if score >= 70 {
		return "格式基本合格，有" + formatNumber(warnings) + "个建议改进项"
	} else if score >= 60 {
		return "格式需要改进，有" + formatNumber(warnings) + "个问题需要修复"
	}
	return "格式检查未通过，有" + formatNumber(warnings) + "个严重问题需要修复"
}
