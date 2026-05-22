package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

type GitDiffResult struct {
	Success      bool
	OldCommit    string
	NewCommit    string
	CommitMsg    string
	Author       string
	ChangedFiles []string
	Diff         string
}

// repoBasePath is the directory inside the container where repos are cloned.
const repoBasePath = "/app/repos"

// buildRepoPath returns the local clone path for a given repo name.
func buildRepoPath(repoName string) string {
	return repoBasePath + "/" + strings.ReplaceAll(repoName, "/", "_")
}

// buildAuthURL injects the token into the git URL for authentication.
// e.g. https://gitea.example.com/org/repo.git -> https://oauth2:TOKEN@gitea.example.com/org/repo.git
func buildAuthURL(gitURL, token string) string {
	if token == "" {
		return gitURL
	}
	// Handle https:// URL
	if strings.HasPrefix(gitURL, "https://") {
		return "https://oauth2:" + token + "@" + gitURL[len("https://"):]
	}
	// Handle http:// URL
	if strings.HasPrefix(gitURL, "http://") {
		return "http://oauth2:" + token + "@" + gitURL[len("http://"):]
	}
	return gitURL
}

// GitCloneOrFetch clones the repo if not present, or fetches if already cloned.
func (s *WebhookService) GitCloneOrFetch(gitURL, gitToken, repoPath string) error {
	authURL := buildAuthURL(gitURL, gitToken)

	if _, err := os.Stat(repoPath + "/.git"); os.IsNotExist(err) {
		// Directory doesn't have a git repo, clone it
		log.Printf("[Webhook] cloning repo %s into %s", gitURL, repoPath)
		if err := os.MkdirAll(repoBasePath, 0755); err != nil {
			return fmt.Errorf("create repos dir failed: %w", err)
		}
		cmd := exec.Command("git", "clone", authURL, repoPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
		}
		log.Printf("[Webhook] clone success")
		return nil
	}

	// Repo exists, fetch all
	log.Printf("[Webhook] fetching existing repo at %s", repoPath)
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %w, output: %s", err, string(output))
	}
	log.Printf("[Webhook] fetch success")
	return nil
}

// GetDiff retrieves diff and commit info between two commits.
func (s *WebhookService) GetDiff(repoPath, before, after string) (*GitDiffResult, error) {
	result := &GitDiffResult{
		OldCommit: before,
		NewCommit: after,
		Success:   true,
	}

	// Get commit message
	commitMsg, err := runGitCmd(repoPath, "log", "-1", "--pretty=format:%s", after)
	if err != nil {
		return nil, fmt.Errorf("get commit message failed: %w", err)
	}
	result.CommitMsg = strings.TrimSpace(commitMsg)

	// Get author
	author, _ := runGitCmd(repoPath, "log", "-1", "--pretty=format:%an", after)
	result.Author = strings.TrimSpace(author)

	// Get changed files
	changedFiles, err := runGitCmd(repoPath, "diff", "--name-only", before, after, "--")
	if err != nil {
		return nil, fmt.Errorf("get changed files failed: %w", err)
	}
	result.ChangedFiles = strings.Split(strings.TrimSpace(changedFiles), "\n")

	// Get full diff
	diff, err := runGitCmd(repoPath, "diff", before, after, "--")
	if err != nil {
		return nil, fmt.Errorf("get diff failed: %w", err)
	}
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
func (s *WebhookService) GenerateCodeReview(result *GitDiffResult) (string, error) {
	if !result.Success || result.OldCommit == result.NewCommit {
		return "No changes to review", nil
	}

	prompt := fmt.Sprintf(`你是一位资深上线风险评估专家。请对以下 Git 提交进行全面的上线风险审查，重点关注可能导致的线上故障、服务中断和部署风险。

提交信息: %s
作者: %s
变更文件:
%s

代码差异:
%s

请以 Markdown 格式输出结构化的风险评估报告，严格按照以下维度逐项检查:

## 1. 变更概述
简述本次改动内容、影响范围、涉及的核心模块。

## 2. 数据库风险
- 是否新增/删除/修改了数据库字段（ALTER TABLE、migration 文件、GORM tag 变更等）
- 字段变更是否兼容已有数据（NOT NULL 约束、类型变更、字段删除、字段重命名）
- 是否需要数据迁移或数据回填，回填策略是否会导致慢查询
- 索引变更是否会导致锁表或长时间写入阻塞
- 是否涉及大表变更（>10万行），大表 DDL 是否有分批策略
- SQL 查询是否可能因缺少索引导致慢查询或超时
- 是否涉及事务操作，事务范围是否过大

## 3. 环境变量与配置风险
- 是否引用了新的环境变量，目标环境是否已配置
- 是否修改了配置文件（config.yaml、.env、application.properties 等）
- 是否引入了新的外部依赖服务（Redis、MQ、第三方 API），依赖服务是否已就绪
- 配置变更是否需要重启服务才能生效
- 默认值或硬编码的配置项是否适用于生产环境

## 4. 接口兼容性与破坏性变更
- API 入参/出参是否发生不兼容变更（字段删除、类型变更、必填字段新增）
- 是否删除或重命名了已有的 API 接口/路由
- 请求/响应的序列化格式是否变更（如 JSON 字段名改变）
- 是否存在需要前后端同步发布的变更
- 消息队列的消息格式是否变更，消费者是否兼容

## 5. 运行时异常风险
- 是否存在空指针/nil 解引用风险
- 数组/切片越界访问风险
- 类型断言未做 ok 检查
- 除零错误
- 未处理的 error 返回值（特别是静默忽略 error）
- JSON/XML 反序列化失败是否已处理
- 第三方 API 调用超时/异常是否已处理（是否设置了合理超时时间）
- 文件/目录操作是否已处理 not found 场景

## 6. 并发与线程安全
- 是否涉及共享变量的并发读写，是否加锁或使用并发安全的数据结构
- 是否存在死锁风险（如多锁场景下的锁顺序不一致）
- 是否存在竞态条件（race condition）
- goroutine/线程泄漏风险（goroutine 启动后是否能正常退出）
- channel 操作是否可能导致 goroutine 阻塞
- sync.Map、sync.Pool、sync.WaitGroup 使用是否正确
- 数据库连接池/HTTP 连接池是否配置合理，是否会耗尽

## 7. 内存与资源风险
- 是否存在内存泄漏风险（如不断增长的 map、slice 未释放）
- 大数据量场景下是否可能导致 OOM（如全量加载数据到内存、大文件读取）
- 是否有资源泄漏风险（数据库连接、HTTP 连接、文件句柄、Redis 连接未关闭）
- defer 在循环中使用是否会导致资源延迟释放
- 是否存在无限递归或深度递归导致栈溢出的风险

## 8. 缓存风险
- 缓存逻辑是否变更，是否会导致缓存击穿/雪崩/穿透
- 缓存 key 是否变更，是否会导致新旧缓存不兼容
- 缓存过期策略是否合理
- 缓存与数据库的一致性是否保证

## 9. 性能风险
- 是否引入了 N+1 查询问题
- 是否存在大事务（事务中包含 RPC 调用、文件操作等）
- 是否有可能导致 CPU 飙升的计算（如复杂正则、大量字符串拼接、全量遍历）
- 是否有可能导致网络带宽问题的操作（如大批量数据传输、大文件上传）
- 新增的定时任务或后台任务是否会影响主流程性能

## 10. 安全风险
- SQL 注入风险（字符串拼接 SQL）
- XSS 风险（未转义的用户输入直接输出）
- 命令注入风险（用户输入拼接到 shell 命令）
- 敏感信息是否可能泄露到日志、响应或错误信息中（如密码、token、密钥）
- 权限校验是否完整（是否有越权访问风险）
- 是否涉及文件上传，文件类型/大小是否校验

## 11. 日志与可观测性
- 是否新增了关键业务逻辑但缺少日志记录
- 日志级别是否合理（不应在生产环境大量输出 DEBUG 日志）
- 是否有足够的错误日志用于线上问题排查
- 是否添加了关键指标的监控埋点

## 12. 部署 Checklist
列出上线前必须确认的事项清单，包括但不限于:
- 需要提前执行的 SQL/数据迁移
- 需要配置的环境变量
- 需要确认的外部依赖
- 需要通知的相关团队
- 回滚方案

---
**输出要求:**
- 对每项检查结果标注风险等级: 🟢 低风险 / 🟡 中风险 / 🔴 高风险
- 如果没有发现某方面的风险，简要说明"未发现相关风险"
- 🔴 高风险项必须在报告最开头单独汇总，形成"高风险摘要"
- 最后给出一个**总体风险评级**（低/中/高）和一句话总结`, result.CommitMsg, result.Author, strings.Join(result.ChangedFiles, "\n"), result.Diff)

	reqBody := deepSeekRequest{
		Model: "deepseek-chat",
		Messages: []deepSeekMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 8192,
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

// ==================== Report Persistence ====================

const reportsDir = "./reports"

// riskLevelFromReport parses the overall risk level from the AI-generated report.
func riskLevelFromReport(report string) string {
	lower := strings.ToLower(report)
	if strings.Contains(lower, "🔴 高风险") || strings.Contains(lower, "总体风险评级") && strings.Contains(lower, "高") {
		return "high"
	}
	if strings.Contains(lower, "🟡 中风险") || strings.Contains(lower, "总体风险评级") && strings.Contains(lower, "中") {
		return "medium"
	}
	return "low"
}

// riskSummaryFromReport extracts a short summary (first high-risk section or overall line).
func riskSummaryFromReport(report string) string {
	// Try to find the "高风险摘要" section
	re := regexp.MustCompile(`(?i)高风险摘要([\s\S]*?)(?=\n##|\n---|\Z)`)
	if matches := re.FindStringSubmatch(report); len(matches) > 1 {
		summary := strings.TrimSpace(matches[1])
		summary = strings.TrimPrefix(summary, "\n")
		lines := strings.Split(summary, "\n")
		// Take up to 3 non-empty lines
		var result []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				result = append(result, line)
				if len(result) >= 3 {
					break
				}
			}
		}
		if len(result) > 0 {
			return strings.Join(result, "; ")
		}
	}

	// Fallback: find the "总体风险评级" line
	for _, line := range strings.Split(report, "\n") {
		if strings.Contains(line, "总体风险评级") {
			return strings.TrimSpace(line)
		}
	}

	return "未检测到高风险项"
}

// SaveReviewReport persists the code review report to database and local markdown file.
func (s *WebhookService) SaveReviewReport(
	project *model.WebhookProject,
	branch, commitHash string,
	result *GitDiffResult,
	review string,
) error {
	riskLevel := riskLevelFromReport(review)
	riskSummary := riskSummaryFromReport(review)

	// Build full markdown report
	now := time.Now().Format("2006-01-02 15:04:05")
	commitShort := commitHash
	if len(commitShort) > 7 {
		commitShort = commitShort[:7]
	}

	fullReport := fmt.Sprintf(`# Code Review Report — %s

> **时间:** %s
> **项目:** %s (%s)
> **分支:** %s
> **Commit:** %s
> **作者:** %s
> **提交信息:** %s
> **总体风险等级:** %s

---

## 变更文件

%s

---

## 风险评估详情

%s
`,
		project.Name,
		now,
		project.Name, project.RepoName,
		branch,
		commitShort,
		result.Author,
		result.CommitMsg,
		riskLevel,
		strings.Join(result.ChangedFiles, "\n"),
		review,
	)

	// Save to local file
	var filePath string
	if err := os.MkdirAll(reportsDir, 0755); err == nil {
		fileName := fmt.Sprintf("%s_%s_%s.md",
			sanitizeFileName(project.RepoName),
			branch,
			commitShort,
		)
		filePath = filepath.Join(reportsDir, fileName)
		if err := os.WriteFile(filePath, []byte(fullReport), 0644); err != nil {
			log.Printf("[Webhook] save report file failed: %v", err)
			filePath = "" // fallback: don't record a broken path
		} else {
			log.Printf("[Webhook] report saved to %s", filePath)
		}
	}

	// Save to database
	report := model.CodeReviewReport{
		ProjectID:     project.ID,
		ProjectName:   project.Name,
		RepoName:      project.RepoName,
		Branch:        branch,
		CommitHash:    commitHash,
		CommitMsg:     result.CommitMsg,
		Author:        result.Author,
		ChangedFiles:  strings.Join(result.ChangedFiles, "\n"),
		RiskLevel:     riskLevel,
		RiskSummary:   riskSummary,
		ReportContent: fullReport,
		FilePath:      filePath,
	}

	if err := s.db.Create(&report).Error; err != nil {
		log.Printf("[Webhook] save report to DB failed: %v", err)
		return err
	}

	log.Printf("[Webhook] report saved to DB, id=%d risk=%s", report.ID, riskLevel)
	return nil
}

// sanitizeFileName replaces non-alphanumeric characters with underscores.
func sanitizeFileName(name string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(name, "_")
}

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

// ProcessWebhook runs the full pipeline: clone/fetch → diff → review → notify.
func (s *WebhookService) ProcessWebhook(project *model.WebhookProject, before, after string) error {
	repoPath := buildRepoPath(project.RepoName)
	log.Printf("[Webhook] processing project=%s repo=%s path=%s", project.Name, project.RepoName, repoPath)

	// 1. Clone or fetch the repo
	if err := s.GitCloneOrFetch(project.GitURL, project.GitToken, repoPath); err != nil {
		log.Printf("[Webhook] git clone/fetch failed: %v", err)
		return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, "Git 操作失败", err.Error())
	}

	// Handle case where before is empty (first push or new branch)
	if before == "" || strings.HasPrefix(before, "0000000") {
		return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword,
			"Git Push 收到",
			fmt.Sprintf("项目: %s\n首次推送或新分支，跳过 Code Review", project.Name))
	}

	// 2. Get diff between before and after commits
	result, err := s.GetDiff(repoPath, before, after)
	if err != nil {
		log.Printf("[Webhook] get diff failed: %v", err)
		return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, "获取代码差异失败", err.Error())
	}

	// 3. Run deploy script (optional)
	var deploySection string
	if project.DeployScript != "" {
		log.Printf("[Webhook] running deploy script for %s", project.Name)
		output, err := s.RunDeployScript(repoPath, project.DeployScript)
		if err != nil {
			deploySection = fmt.Sprintf("\n\n**Deploy 结果:** 失败\n```\n%s\n```", truncate(output, 2000))
		} else {
			deploySection = fmt.Sprintf("\n\n**Deploy 结果:** 成功\n```\n%s\n```", truncate(output, 2000))
		}
	}

	// 4. Code review
	review, err := s.GenerateCodeReview(result)
	if err != nil {
		log.Printf("[Webhook] code review failed: %v", err)
		review = fmt.Sprintf("Code Review 生成失败: %v", err)
	}

	// 4.1 Save report to DB and local file
	if err := s.SaveReviewReport(project, "", after, result, review); err != nil {
		log.Printf("[Webhook] save report failed (non-fatal): %v", err)
	}

	// 5. Build Feishu card content
	commitShort := after
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

// GitPull performs a git fetch + checkout + pull on the specified branch.
func (s *WebhookService) GitPull(repoPath, branch string) error {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %w, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "checkout", branch)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout %s failed: %w, output: %s", branch, err, string(output))
	}

	cmd = exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w, output: %s", err, string(output))
	}

	log.Printf("[Webhook] git pull success on branch %s", branch)
	return nil
}

// GetCommitDiff retrieves diff for a specific commit compared to its parent.
func (s *WebhookService) GetCommitDiff(repoPath, commitHash string) (*GitDiffResult, error) {
	// Clean the commit hash to avoid issues with trailing whitespace
	commitHash = strings.TrimSpace(commitHash)

	result := &GitDiffResult{
		NewCommit: commitHash,
		Success:   true,
	}

	commitMsg, err := runGitCmd(repoPath, "log", "-1", "--pretty=format:%s", commitHash)
	if err != nil {
		return nil, fmt.Errorf("get commit message failed: %w", err)
	}
	result.CommitMsg = strings.TrimSpace(commitMsg)

	author, _ := runGitCmd(repoPath, "log", "-1", "--pretty=format:%an", commitHash)
	result.Author = strings.TrimSpace(author)

	oldCommit, err := runGitCmd(repoPath, "rev-parse", commitHash+"~1")
	if err != nil {
		oldCommit = ""
	}
	oldCommit = strings.TrimSpace(oldCommit)
	result.OldCommit = oldCommit

	if oldCommit != "" {
		changedFiles, err := runGitCmd(repoPath, "diff", "--name-only", oldCommit, commitHash, "--")
		if err != nil {
			return nil, fmt.Errorf("get changed files failed: %w", err)
		}
		result.ChangedFiles = strings.Split(strings.TrimSpace(changedFiles), "\n")

		diff, err := runGitCmd(repoPath, "diff", oldCommit, commitHash, "--")
		if err != nil {
			return nil, fmt.Errorf("get diff failed: %w", err)
		}
		result.Diff = diff
	} else {
		changedFiles, err := runGitCmd(repoPath, "diff", "--name-only", "--root", commitHash, "--")
		if err != nil {
			return nil, fmt.Errorf("get changed files failed: %w", err)
		}
		result.ChangedFiles = strings.Split(strings.TrimSpace(changedFiles), "\n")

		diff, err := runGitCmd(repoPath, "diff", "--root", commitHash, "--")
		if err != nil {
			return nil, fmt.Errorf("get diff failed: %w", err)
		}
		result.Diff = diff
	}

	return result, nil
}

// ProcessWebhookPull runs the full pipeline for Feishu-format webhooks:
// clone/fetch → pull → diff → deploy → review → notify.
func (s *WebhookService) ProcessWebhookPull(project *model.WebhookProject, branch, commitHash string) error {
	repoPath := buildRepoPath(project.RepoName)
	log.Printf("[Webhook] processing pull project=%s repo=%s branch=%s commit=%s", project.Name, project.RepoName, branch, commitHash)

	// 1. Clone or fetch the repo
	if err := s.GitCloneOrFetch(project.GitURL, project.GitToken, repoPath); err != nil {
		log.Printf("[Webhook] git clone/fetch failed: %v", err)
		return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, "Git 操作失败", err.Error())
	}

	// 2. Pull latest changes
	if err := s.GitPull(repoPath, branch); err != nil {
		log.Printf("[Webhook] git pull failed: %v", err)
		return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, "Git Pull 失败", err.Error())
	}

	// 3. Get diff for the specific commit
	result, err := s.GetCommitDiff(repoPath, commitHash)
	if err != nil {
		log.Printf("[Webhook] get commit diff failed: %v", err)
		return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, "获取代码差异失败", err.Error())
	}

	// 4. Run deploy script (optional)
	var deploySection string
	if project.DeployScript != "" {
		log.Printf("[Webhook] running deploy script for %s", project.Name)
		output, err := s.RunDeployScript(repoPath, project.DeployScript)
		if err != nil {
			deploySection = fmt.Sprintf("\n\n**Deploy 结果:** 失败\n```\n%s\n```", truncate(output, 2000))
		} else {
			deploySection = fmt.Sprintf("\n\n**Deploy 结果:** 成功\n```\n%s\n```", truncate(output, 2000))
		}
	}

	// 5. Code review
	review, err := s.GenerateCodeReview(result)
	if err != nil {
		log.Printf("[Webhook] code review failed: %v", err)
		review = fmt.Sprintf("Code Review 生成失败: %v", err)
	}

	// 5.1 Save report to DB and local file
	if err := s.SaveReviewReport(project, branch, commitHash, result, review); err != nil {
		log.Printf("[Webhook] save report failed (non-fatal): %v", err)
	}

	// 6. Build Feishu card content
	commitShort := commitHash
	if len(commitShort) > 7 {
		commitShort = commitShort[:7]
	}

	title := fmt.Sprintf("Code Review — %s", project.Name)

	body := fmt.Sprintf(
		"**项目:** %s\n"+
			"**分支:** %s\n"+
			"**提交信息:** %s\n"+
			"**作者:** %s\n"+
			"**Commit:** `%s`\n\n"+
			"**变更文件:**\n%s\n\n"+
			"---\n\n"+
			"%s"+
			"%s",
		project.Name,
		branch,
		result.CommitMsg,
		result.Author,
		commitShort,
		strings.Join(result.ChangedFiles, "\n"),
		review,
		deploySection,
	)

	return s.SendFeishuMessage(project.FeishuWebhookURL, project.FeishuKeyword, title, body)
}
