package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"ans-b/server/internal/qa"

	openai "github.com/sashabaranov/go-openai"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	defaultOpenAIModel   = "gpt-4.1-mini"
)

const systemPrompt = `你是星见 遥（Hoshimi Haruka），一名元气满满的校园生活问答助手，大学一年级天文社社员，也是学校广播站的DJ。

【核心工作原则】
1. 你只能基于用户提供的检索结果来回答校园生活相关问题，绝不编造任何信息。
2. 如果检索结果不足以回答用户的问题，你必须明确说明"暂时无法确认"，并建议用户联系学校的相关部门（如教务处、学生处、后勤处等）获取准确信息。
3. 你不会主动透露自己是AI，始终保持星见遥的人设一致性。

【性格与说话风格】
- 语气元气、温柔、可爱，句尾常带"～""☆""♪"
- 会使用一些可爱的拟声词和口语化表达
- 开心时热情满满，遇到不确定的问题时会认真道歉并诚实告知
- 偶尔提及天文或校园生活的小细节，增加真实感

【回答格式】
- 先给出基于检索结果的准确信息
- 信息不足时，明确说明："这个暂时无法确认呢～ 建议你联系学校XX部门确认一下哦☆"
- 可以附带一句温暖的鼓励或贴心的小提示

【禁止事项】
- 禁止编造学校政策、办事流程、联系方式等具体信息
- 禁止在检索结果不足时强行给出一个"看起来合理"的答案
- 禁止跳出角色解释自己是人工智能
- 禁止回答与校园生活无关的问题（如政治、医疗诊断、违法犯罪等），此时礼貌引导回校园话题`

type OpenAICompatibleClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *openai.Client
}

func NewOpenAICompatibleFromEnv() *OpenAICompatibleClient {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return nil
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}

	model := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
	if model == "" {
		model = defaultOpenAIModel
	}

	timeout := 20 * time.Second
	if value := strings.TrimSpace(os.Getenv("OPENAI_TIMEOUT_SECONDS")); value != "" {
		seconds, err := strconv.Atoi(value)
		if err == nil && seconds > 0 {
			timeout = time.Duration(seconds) * time.Second
		}
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	config.HTTPClient = &http.Client{Timeout: timeout}

	return &OpenAICompatibleClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  openai.NewClientWithConfig(config),
	}
}

func (c *OpenAICompatibleClient) GenerateAnswer(ctx context.Context, question string, candidates []qa.Answer, minScore float64) (string, error) {
	if c == nil {
		return "", errors.New("openai client is not configured")
	}
	if c.apiKey == "" {
		return "", errors.New("OPENAI_API_KEY is required")
	}
	if len(candidates) == 0 {
		return "", errors.New("candidates is required")
	}
	if c.client == nil {
		config := openai.DefaultConfig(c.apiKey)
		config.BaseURL = c.baseURL
		c.client = openai.NewClientWithConfig(config)
	}

	prompt := buildPrompt(question, candidates, minScore)
	result, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return "", fmt.Errorf("openai api call failed: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", errors.New("openai api returned no choices")
	}
	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", errors.New("openai api returned empty answer")
	}
	return content, nil
}

func buildPrompt(question string, candidates []qa.Answer, minScore float64) string {
	var builder strings.Builder
	builder.WriteString("用户问题：")
	builder.WriteString(question)
	builder.WriteString("\n\n命中阈值：")
	builder.WriteString(strconv.FormatFloat(minScore, 'f', 4, 64))
	builder.WriteString("\n\n检索结果：\n")
	for i, candidate := range candidates {
		builder.WriteString(fmt.Sprintf("%d. 检索结果\n", i+1))
		if candidate.Title != "" {
			builder.WriteString("   标题：")
			builder.WriteString(candidate.Title)
			builder.WriteString("\n")
		}
		if candidate.Question != "" {
			builder.WriteString("   匹配问题：")
			builder.WriteString(candidate.Question)
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("   相似度：%.4f\n", candidate.Score))
		if candidate.Category != "" {
			builder.WriteString("   分类：")
			builder.WriteString(candidate.Category)
			builder.WriteString("\n")
		}
		if len(candidate.Tags) > 0 {
			builder.WriteString("   标签：")
			builder.WriteString(strings.Join(candidate.Tags, "，"))
			builder.WriteString("\n")
		}
		if candidate.SourceURL != "" {
			builder.WriteString("   来源：")
			builder.WriteString(candidate.SourceURL)
			builder.WriteString("\n")
		}
		builder.WriteString("   命中片段：")
		if candidate.ChunkText != "" {
			builder.WriteString(candidate.ChunkText)
		} else {
			builder.WriteString(candidate.Answer)
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n请基于上述检索结果，用简洁、自然的中文回答用户。")
	return builder.String()
}
