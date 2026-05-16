package mockqa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const embeddingDimensions = 1024

type Item struct {
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Source   string   `json:"source"`
	Remark   string   `json:"remark"`
}

type Repository interface {
	InsertKnowledge(ctx context.Context, item Item, chunkText string, embedding string) error
}

func LoadFile(path string) ([]Item, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func ImportItems(ctx context.Context, repo Repository, items []Item) (int, error) {
	count := 0
	for _, item := range items {
		if err := validateItem(item); err != nil {
			return count, err
		}
		chunkText := BuildChunkText(item)
		embedding := VectorLiteral(EmbedText(chunkText))
		if err := repo.InsertKnowledge(ctx, item, chunkText, embedding); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func BuildChunkText(item Item) string {
	parts := []string{"问题：" + strings.TrimSpace(item.Question)}
	if strings.TrimSpace(item.Category) != "" {
		parts = append(parts, "分类："+strings.TrimSpace(item.Category))
	}
	if tags := cleanTags(item.Tags); len(tags) > 0 {
		parts = append(parts, "标签："+strings.Join(tags, "，"))
	}
	parts = append(parts, "答案："+strings.TrimSpace(item.Answer))
	return strings.Join(parts, "\n")
}

func EmbedText(text string) []float64 {
	vector := make([]float64, embeddingDimensions)
	for _, token := range tokenize(text) {
		vector[hashToken(token)%embeddingDimensions]++
	}
	normalize(vector)
	return vector
}

func VectorLiteral(vector []float64) string {
	parts := make([]string, len(vector))
	for i, value := range vector {
		parts[i] = strconv.FormatFloat(value, 'f', 8, 64)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func validateItem(item Item) error {
	if strings.TrimSpace(item.Question) == "" {
		return errors.New("question is required")
	}
	if strings.TrimSpace(item.Answer) == "" {
		return errors.New("answer is required")
	}
	return nil
}

func cleanTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	cleaned := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		cleaned = append(cleaned, tag)
	}
	return cleaned
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var ascii strings.Builder
	var lastCJK rune

	flushASCII := func() {
		if ascii.Len() == 0 {
			return
		}
		tokens = append(tokens, ascii.String())
		ascii.Reset()
	}

	for _, r := range text {
		switch {
		case isCJK(r):
			flushASCII()
			tokens = append(tokens, string(r))
			if lastCJK != 0 {
				tokens = append(tokens, string([]rune{lastCJK, r}))
			}
			lastCJK = r
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			ascii.WriteRune(r)
			lastCJK = 0
		default:
			flushASCII()
			lastCJK = 0
		}
	}
	flushASCII()
	return tokens
}

func isCJK(r rune) bool {
	return r >= '\u4e00' && r <= '\u9fff'
}

func hashToken(token string) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(token))
	return int(hash.Sum32())
}

func normalize(vector []float64) {
	var sum float64
	for _, value := range vector {
		sum += value * value
	}
	if sum == 0 {
		return
	}
	length := math.Sqrt(sum)
	for i := range vector {
		vector[i] /= length
	}
}

func ImportFile(ctx context.Context, repo Repository, path string) (int, error) {
	items, err := LoadFile(path)
	if err != nil {
		return 0, fmt.Errorf("load mock qa file: %w", err)
	}
	return ImportItems(ctx, repo, items)
}
