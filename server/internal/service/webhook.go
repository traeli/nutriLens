package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"nutrilens/internal/model"

	"gorm.io/gorm"
)

type WebhookService struct {
	deepseekAPIKey  string
	deepseekBaseURL string
	db              *gorm.DB
}

func NewWebhookService(deepseekAPIKey, deepseekBaseURL string, db *gorm.DB) *WebhookService {
	return &WebhookService{
		deepseekAPIKey:  deepseekAPIKey,
		deepseekBaseURL: deepseekBaseURL,
		db:              db,
	}
}

// FindProjectByRepoName looks up an enabled webhook project by Gitea repository full name.
func (s *WebhookService) FindProjectByRepoName(repoName string) (*model.WebhookProject, error) {
	var project model.WebhookProject
	if err := s.db.Where("repo_name = ? AND enabled = ?", repoName, true).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
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

// GitPull executes git pull in the given repo path.
func (s *WebhookService) GitPull(repoPath string) (*GitPullResult, error) {
	result := &GitPullResult{}

	oldCommit, err := runGitCmd(repoPath, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("get current commit failed: %w", err)
	}
	result.OldCommit = strings.TrimSpace(oldCommit)

	_, err = runGitCmd(repoPath, "pull", "-p")
	if err != nil {
		return nil, fmt.Errorf("git pull failed: %w", err)
	}

	newCommit, err := runGitCmd(repoPath, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("get new commit failed: %w", err)
	}
	result.NewCommit = strings.TrimSpace(newCommit)

	if result.OldCommit == result.NewCommit {
		result.Success = true
		result.CommitMsg = "No updates"
		return result, nil
	}

	result.Success = true

	commitMsg, _ := runGitCmd(repoPath, "log", "-1", "--pretty=format:%s")
	result.CommitMsg = strings.TrimSpace(commitMsg)

	author, _ := runGitCmd(repoPath, "log", "-1", "--pretty=format:%an")
	result.Author = strings.TrimSpace(author)

	changedFiles, _ := runGitCmd(repoPath, "diff", "--name-only", result.OldCommit, result.NewCommit)
	result.ChangedFiles = strings.Split(strings.TrimSpace(changedFiles), "\n")

	diff, _ := runGitCmd(repoPath, "diff", result.OldCommit, result.NewCommit)
	result.Diff = diff

	return result, nil
}

func runGitCmd(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w, output: %s", strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

// RunDeployScript writes the script to a temp file and executes it with bash.
func (s *WebhookService) RunDeployScript(repoPath, script string) (string, error) {
	tmpFile, err := os.CreateTemp("", "deploy-*.sh")
	if err != nil {
		return "", fmt.Errorf("create temp file failed: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("write script failed: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return "", fmt.Errorf("chmod script failed: %w", err)
	}

	cmd := exec.Command("bash", tmpFile.Name())
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("script failed: %w, output: %s", err, string(output))
	}

	log.Printf("[Webhook] deploy script success, output: %s", string(output))
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

// GenerateCodeReview calls DeepSeek API to review the changes.
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

// SendFeishuMessage sends an interactive card to the given Feishu webhook URL.
func (s *WebhookService) SendFeishuMessage(webhookURL, keyword, title, content string) error {
	if webhookURL == "" {
		log.Printf("[Feishu] no webhook URL configured, skip sending")
		return nil
	}

	if len(content) > 18000 {
		content = content[:18000] + "\n\n... *(内容过长已截断)*"
	}

	noteText := "Powered by DeepSeek · NutriLens Auto Review"
	if keyword != "" {
		noteText = keyword + " | " + noteText
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

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("send to Feishu failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody bytes.Buffer
		errBody.ReadFrom(resp.Body)
		return fmt.Errorf("Feishu returned status %d: %s", resp.StatusCode, errBody.String())
	}

	log.Printf("[Feishu] card sent successfully")
	return nil
}

// ==================== Full Pipeline ====================

// ProcessWebhook runs the full pipeline for a given project: git pull → deploy → review → notify.
func (s *WebhookService) ProcessWebhook(project *model.WebhookProject) error {
	log.Printf("[Webhook] processing project=%s repo=%s", project.Name, project.RepoPath)

	// 1. Git pull
	result, err := s.GitPull(project.RepoPath)
	if err != nil {
		log.Printf("[Webhook] git pull failed: %v", err)
		return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, "Git Pull 失败", err.Error())
	}

	if result.OldCommit == result.NewCommit {
		return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, "Git Pull 完成", "没有新的更新")
	}

	// 2. Run deploy script
	var deploySection string
	if project.DeployScript != "" {
		log.Printf("[Webhook] running deploy script for %s", project.Name)
		output, err := s.RunDeployScript(project.RepoPath, project.DeployScript)
		if err != nil {
			deploySection = fmt.Sprintf("\n\n**Deploy 结果:** 失败\n```\n%s\n```", truncate(output, 2000))
		} else {
			deploySection = fmt.Sprintf("\n\n**Deploy 结果:** 成功\n```\n%s\n```", truncate(output, 2000))
		}
	}

	// 3. Code review
	review, err := s.GenerateCodeReview(result)
	if err != nil {
		log.Printf("[Webhook] code review failed: %v", err)
		review = fmt.Sprintf("Code Review 生成失败: %v", err)
	}

	// 4. Build Feishu card content
	commitShort := result.NewCommit
	if len(commitShort) > 7 {
		commitShort = commitShort[:7]
	}

	title := fmt.Sprintf("Code Review — %s", project.Name)

	body := fmt.Sprintf(
		"**项目:** %s\n"+
			"**提交信息:** %s\n"+
			"**作者:** %s\n"+
			"**Commit:** `%s`\n\n"+
			"**变更文件:**\n%s\n\n"+
			"---\n\n"+
			"%s"+
			"%s",
		project.Name,
		result.CommitMsg,
		result.Author,
		commitShort,
		strings.Join(result.ChangedFiles, "\n"),
		review,
		deploySection,
	)

	return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, title, body)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "\n... (截断)"
	}
	return s
}
