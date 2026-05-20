package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type WebhookService struct {
	feishuWebhookURL string
	deepseekAPIKey   string
	deepseekBaseURL  string
	repoPath         string
	keyword          string
}

func NewWebhookService(feishuWebhookURL, deepseekAPIKey, deepseekBaseURL, repoPath, keyword string) *WebhookService {
	return &WebhookService{
		feishuWebhookURL: feishuWebhookURL,
		deepseekAPIKey:   deepseekAPIKey,
		deepseekBaseURL:  deepseekBaseURL,
		repoPath:         repoPath,
		keyword:          keyword,
	}
}

type GitPullResult struct {
	Success      bool
	OldCommit    string
	NewCommit    string
	CommitMsg    string
	Author       string
	ChangedFiles []string
	Diff         string
}

// GitPull executes git pull and returns the result
func (s *WebhookService) GitPull() (*GitPullResult, error) {
	result := &GitPullResult{}

	// Get current commit
	oldCommit, err := s.runGitCommand("rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("get current commit failed: %w", err)
	}
	result.OldCommit = strings.TrimSpace(oldCommit)

	// Git pull
	_, err = s.runGitCommand("pull", "-p")
	if err != nil {
		return nil, fmt.Errorf("git pull failed: %w", err)
	}

	// Get new commit
	newCommit, err := s.runGitCommand("rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("get new commit failed: %w", err)
	}
	result.NewCommit = strings.TrimSpace(newCommit)

	// Check if there are updates
	if result.OldCommit == result.NewCommit {
		result.Success = true
		result.CommitMsg = "No updates"
		return result, nil
	}

	result.Success = true

	// Get commit message
	commitMsg, _ := s.runGitCommand("log", "-1", "--pretty=format:%s")
	result.CommitMsg = strings.TrimSpace(commitMsg)

	// Get author
	author, _ := s.runGitCommand("log", "-1", "--pretty=format:%an")
	result.Author = strings.TrimSpace(author)

	// Get changed files
	changedFiles, _ := s.runGitCommand("diff", "--name-only", result.OldCommit, result.NewCommit)
	result.ChangedFiles = strings.Split(strings.TrimSpace(changedFiles), "\n")

	// Get diff
	diff, _ := s.runGitCommand("diff", result.OldCommit, result.NewCommit)
	result.Diff = diff

	return result, nil
}

func (s *WebhookService) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %w, output: %s", strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

// ==================== DeepSeek API ====================

type deepSeekRequest struct {
	Model     string            `json:"model"`
	Messages  []deepSeekMessage `json:"messages"`
	MaxTokens int               `json:"max_tokens"`
}

type deepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// GenerateCodeReview calls DeepSeek API to review the changes
func (s *WebhookService) GenerateCodeReview(result *GitPullResult) (string, error) {
	if !result.Success || result.OldCommit == result.NewCommit {
		return "No changes to review", nil
	}

	prompt := fmt.Sprintf(`请对以下 Git 提交进行 Code Review，检查代码质量、潜在问题和改进建议。

提交信息: %s
作者: %s
变更文件:
%s

代码差异:
%s

请以 Markdown 格式输出 Code Review 报告，包含:
1. 变更概述
2. 代码质量评估
3. 潜在问题
4. 改进建议
5. 总体评价`, result.CommitMsg, result.Author, strings.Join(result.ChangedFiles, "\n"), result.Diff)

	reqBody := deepSeekRequest{
		Model: "deepseek-chat",
		Messages: []deepSeekMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 4096,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	apiURL := strings.TrimRight(s.deepseekBaseURL, "/") + "/v1/chat/completions"

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.deepseekAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call DeepSeek API failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody bytes.Buffer
		errBody.ReadFrom(resp.Body)
		return "", fmt.Errorf("DeepSeek API returned status %d: %s", resp.StatusCode, errBody.String())
	}

	var respBody deepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("decode response failed: %w", err)
	}

	if len(respBody.Choices) == 0 {
		return "", fmt.Errorf("empty response from DeepSeek")
	}

	return respBody.Choices[0].Message.Content, nil
}

// ==================== Feishu Interactive Card ====================

type feishuCardMessage struct {
	MsgType string     `json:"msg_type"`
	Card    feishuCard `json:"card"`
}

type feishuCard struct {
	Config   feishuCardConfig `json:"config"`
	Header   feishuCardHeader `json:"header"`
	Elements []interface{}    `json:"elements"`
}

type feishuCardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode"`
}

type feishuCardHeader struct {
	Title    feishuPlainText `json:"title"`
	Template string          `json:"template"`
}

type feishuPlainText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type feishuMarkdownBlock struct {
	Tag  string         `json:"tag"`
	Text feishuMarkdown `json:"text"`
}

type feishuMarkdown struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type feishuDivider struct {
	Tag string `json:"tag"`
}

type feishuNote struct {
	Tag      string            `json:"tag"`
	Elements []feishuPlainText `json:"elements"`
}

// SendFeishuMessage sends a nicely formatted interactive card to Feishu webhook
func (s *WebhookService) SendFeishuMessage(title, content string) error {
	// Truncate content if too long (Feishu card has size limits)
	if len(content) > 18000 {
		content = content[:18000] + "\n\n... *(内容过长已截断)*"
	}

	// Embed Feishu bot keyword into the card (required by Feishu security settings)
	noteText := "Powered by DeepSeek · NutriLens Auto Review"
	if s.keyword != "" {
		noteText = s.keyword + " | " + noteText
	}

	card := feishuCardMessage{
		MsgType: "interactive",
		Card: feishuCard{
			Config: feishuCardConfig{
				WideScreenMode: true,
			},
			Header: feishuCardHeader{
				Title: feishuPlainText{
					Tag:     "plain_text",
					Content: title,
				},
				Template: "blue",
			},
			Elements: []interface{}{
				feishuMarkdownBlock{
					Tag: "div",
					Text: feishuMarkdown{
						Tag:     "lark_md",
						Content: content,
					},
				},
				feishuDivider{Tag: "hr"},
				feishuNote{
					Tag: "note",
					Elements: []feishuPlainText{
						{Tag: "plain_text", Content: noteText},
					},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	resp, err := http.Post(s.feishuWebhookURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("send to Feishu failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody bytes.Buffer
		errBody.ReadFrom(resp.Body)
		return fmt.Errorf("Feishu returned status %d: %s", resp.StatusCode, errBody.String())
	}

	log.Printf("[Feishu] interactive card sent successfully")
	return nil
}

// ProcessWebhook handles the complete webhook flow
func (s *WebhookService) ProcessWebhook() error {
	// 1. Git pull
	result, err := s.GitPull()
	if err != nil {
		log.Printf("[Webhook] git pull failed: %v", err)
		return s.SendFeishuMessage("Git Pull 失败", err.Error())
	}

	if result.OldCommit == result.NewCommit {
		return s.SendFeishuMessage("Git Pull 完成", "没有新的更新")
	}

	// 2. Generate code review
	review, err := s.GenerateCodeReview(result)
	if err != nil {
		log.Printf("[Webhook] code review failed: %v", err)
		review = fmt.Sprintf("Code Review 生成失败: %v", err)
	}

	// 3. Send to Feishu
	commitShort := result.NewCommit
	if len(commitShort) > 7 {
		commitShort = commitShort[:7]
	}

	title := fmt.Sprintf("Git Pull & Code Review — %s", commitShort)

	body := fmt.Sprintf(
		"**提交信息:** %s\n"+
			"**作者:** %s\n"+
			"**Commit:** `%s`\n\n"+
			"**变更文件:**\n%s\n\n"+
			"---\n\n"+
			"%s",
		result.CommitMsg,
		result.Author,
		commitShort,
		strings.Join(result.ChangedFiles, "\n"),
		review,
	)

	return s.SendFeishuMessage(title, body)
}
