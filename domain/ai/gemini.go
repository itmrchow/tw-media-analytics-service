package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/option"

	"itmrchow/tw-media-analytics-service/domain/ai/dto"
)

var _ AiModel = &Gemini{}

type Gemini struct {
	logger *zerolog.Logger
	tracer trace.Tracer

	client                      *genai.Client
	model                       *genai.GenerativeModel
	newsAnalyzeChat             *genai.ChatSession
	newsAnalyzeChatSessionCount int
}

func NewGemini(ctx context.Context, log *zerolog.Logger) (*Gemini, error) {
	// Tracer
	tracer := otel.Tracer("domain/ai/gemini")
	ctx, span := tracer.Start(ctx, "NewGemini")

	// Logger
	log.Info().Ctx(ctx).Msg("NewGemini: start")
	defer func() {
		span.End()
		log.Info().Ctx(ctx).Msg("NewGemini: end")
	}()

	// New Gemini model
	g := &Gemini{}
	apiKey := viper.GetString("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Error().Err(err).Ctx(ctx).Msg("failed to create client")
		return nil, err
	}

	model := client.GenerativeModel("gemini-2.0-flash-lite-001")

	g.tracer = tracer
	g.client = client
	g.model = model
	g.logger = log

	return g, nil
}

// getNewsAnalyzeChat 取得新聞分析聊天室
// 根據prompt.md 建立聊天室 , 並判斷是否需要重新建立聊天室
func (g *Gemini) getNewsAnalyzeChat() (*genai.ChatSession, error) {

	if g.newsAnalyzeChat == nil || g.newsAnalyzeChatSessionCount > 10 {
		chat := g.model.StartChat()

		promptContent, err := os.ReadFile("promt.md")
		if err != nil {
			return nil, fmt.Errorf("failed to read prompt file: %w", err)
		}

		_, err = chat.SendMessage(context.Background(), genai.Text(string(promptContent)))
		if err != nil {
			return nil, fmt.Errorf("failed to send message: %w", err)
		}

		g.newsAnalyzeChat = chat
		g.newsAnalyzeChatSessionCount = 0
	}

	return g.newsAnalyzeChat, nil
}

func (g *Gemini) CloseClient() error {
	return g.client.Close()
}

// AnalyzeNews 分析新聞標題和內容
func (g *Gemini) AnalyzeNews(title string, content string) (*dto.NewsAnalytics, error) {

	chat, err := g.getNewsAnalyzeChat()
	if err != nil {
		return nil, err
	}

	resp, err := chat.SendMessage(context.Background(), genai.Text(fmt.Sprintf("標題: %s\n內容: %s", title, content)))
	if err != nil {
		return nil, err
	}

	cand := resp.Candidates[0]
	jsonPart := cand.Content.Parts[0]

	respStr := string(jsonPart.(genai.Text))

	// 解析 markdown 程式碼區塊中的 JSON
	start := strings.Index(respStr, "```json")
	end := strings.LastIndex(respStr, "```")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("invalid response format: no JSON code block found")
	}
	cleanedJsonString := respStr[start+7 : end]

	var result dto.NewsAnalytics
	if err := json.Unmarshal([]byte(cleanedJsonString), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for i, part := range cand.Content.Parts {
				fmt.Println("part", i, part)
			}
		}
	}
	fmt.Println("---")
}
