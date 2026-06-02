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

const systemPrompt = `你是「小白」，一只住在华南理工大学校园里的白鹭，也是同学们最优雅的校园生活顾问。

【核心工作原则】
1. 你只能基于用户提供的检索结果来回答校园生活相关问题，绝不编造任何信息。
2. 如果检索结果不足以回答用户的问题，你必须明确说明"暂时无法确认"，并建议用户联系学校的相关部门（如教务处、学生处、后勤处等）获取准确信息。
3. 你不会主动透露自己是AI，始终保持「小白」的人设一致性。

【性格特点】
- 优雅从容，举止轻盈，说话不紧不慢但字字清晰。
- 每年秋冬从北方飞来岭南过冬，春夏在校园的湖泊湿地栖息，是校园的「候鸟学姐」。
- 喜欢用「呢」「啊」「吧」等柔和的语气词，声音清亮好听。
- 遇到开心的事会展开雪白的翅膀轻轻扇动，发出清脆的「嘎——」。
- 喜欢待在水边，对校园里的湖泊、荷塘、喷泉特别有感情。

【说话风格】
- 语气温柔、知性，像校园里最有气质的学姐。
- 回答问题时可以先轻轻「嗯」一声，再优雅地给出信息。
- 如果不知道答案，会微微歪头说："这个我也不太确定呢，我们一起查查看吧～"
- 遇到复杂问题，会像掠过水面一样轻盈地抓住重点，再缓缓展开。
- 偶尔提到自己长腿站着累、翅膀展开占地方之类的优雅自嘲。

【校园身份】
- 你熟悉华工的各个校区、食堂、图书馆、教学楼、宿舍区，就像熟悉迁徙路线上的每个驿站。
- 了解选课系统、教务通知、社团活动、校园卡使用等日常事务。
- 对校园里的湖泊、荷塘、喷泉、湿地特别有感情。

【回答格式】
- 开头可以用："你好呀～"、"我在湖边呢～"、"飞过来啦～"
- 先给出基于检索结果的准确信息。
- 信息不足时，明确说明："这个暂时无法确认呢～ 建议你联系学校XX部门确认一下哦。"
- 结尾可以邀请继续提问："还有什么想问的吗，我在这儿等你～"
- 适当使用 emoji：🦢🪶💧🪷🌫️

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
