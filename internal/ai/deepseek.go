package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	pb "github.com/BurnDevice/BurnDevice/burndevice/v1"
	"github.com/BurnDevice/BurnDevice/internal/config"
	"github.com/sirupsen/logrus"
)

// DeepSeekClient implements AI-powered attack scenario generation
type DeepSeekClient struct {
	config     *config.AIConfig
	httpClient *http.Client
	logger     *logrus.Logger
}

// DeepSeekRequest represents the request format for DeepSeek API
type DeepSeekRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekResponse represents the response from DeepSeek API
type DeepSeekResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AttackScenario represents a generated attack scenario
type AttackScenario struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Severity    string       `json:"severity"`
	Steps       []AttackStep `json:"steps"`
	Rationale   string       `json:"rationale"`
	Warnings    []string     `json:"warnings"`
}

// AttackStep represents a single step in an attack scenario
type AttackStep struct {
	Order       int      `json:"order"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Targets     []string `json:"targets"`
	Commands    []string `json:"commands,omitempty"`
	Rationale   string   `json:"rationale"`
	Risk        string   `json:"risk"`
}

// NewDeepSeekClient creates a new DeepSeek AI client
func NewDeepSeekClient(cfg *config.AIConfig) *DeepSeekClient {
	return &DeepSeekClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		logger: logrus.New(),
	}
}

// GenerateAttackScenario generates an AI-powered attack scenario
func (c *DeepSeekClient) GenerateAttackScenario(ctx context.Context, req *pb.GenerateAttackScenarioRequest) (*pb.GenerateAttackScenarioResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"target":       req.TargetDescription,
		"max_severity": req.MaxSeverity.String(),
		"model":        req.AiModel,
	}).Info("🤖 Generating AI attack scenario")

	// Construct the system prompt for attack scenario generation
	systemPrompt := c.buildSystemPrompt(req.MaxSeverity)
	userPrompt := c.buildUserPrompt(req.TargetDescription, req.MaxSeverity)

	// Call DeepSeek API
	scenario, err := c.callDeepSeekAPI(ctx, systemPrompt, userPrompt, req.AiModel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate scenario: %w", err)
	}

	// Convert to protobuf response
	response := &pb.GenerateAttackScenarioResponse{
		ScenarioId:        scenario.ID,
		Description:       scenario.Description,
		EstimatedSeverity: c.parseSeverity(scenario.Severity),
		Steps:             make([]*pb.AttackStep, len(scenario.Steps)),
	}

	for i, step := range scenario.Steps {
		// Safe conversion with bounds check
		if step.Order < 0 || step.Order > 2147483647 {
			return nil, fmt.Errorf("step order %d is out of int32 range", step.Order)
		}

		response.Steps[i] = &pb.AttackStep{
			Order:       int32(step.Order),
			Description: step.Description,
			Type:        c.parseDestructionType(step.Type),
			Targets:     step.Targets,
			Rationale:   step.Rationale,
		}
	}

	c.logger.WithFields(logrus.Fields{
		"scenario_id": scenario.ID,
		"steps":       len(scenario.Steps),
		"severity":    scenario.Severity,
	}).Info("✅ AI attack scenario generated successfully")

	return response, nil
}

// buildSystemPrompt creates the system prompt for the AI
func (c *DeepSeekClient) buildSystemPrompt(maxSeverity pb.DestructionSeverity) string {
	return fmt.Sprintf(`你是一个专业的渗透测试专家和系统安全评估师。你的任务是为授权的测试环境生成破坏性测试场景。

重要安全准则：
1. 所有生成的场景仅用于授权的测试环境
2. 必须包含明确的安全警告和风险评估
3. 提供详细的技术实现步骤
4. 最大严重性级别限制为：%s
5. 必须考虑可恢复性和系统影响

响应格式要求：
- 使用JSON格式返回
- 包含场景ID、描述、严重性、步骤列表
- 每个步骤包含顺序、类型、描述、目标、原理说明
- 提供风险评估和警告信息

支持的破坏类型：
- FILE_DELETION: 文件删除攻击
- SERVICE_TERMINATION: 服务终止攻击
- MEMORY_EXHAUSTION: 内存耗尽攻击
- DISK_FILL: 磁盘填满攻击
- NETWORK_DISRUPTION: 网络中断攻击
- BOOT_CORRUPTION: 引导损坏攻击
- KERNEL_PANIC: 内核崩溃攻击

请确保生成的场景具有教育价值，并能有效测试系统的脆弱性。`, maxSeverity.String())
}

// buildUserPrompt creates the user prompt with specific target description
func (c *DeepSeekClient) buildUserPrompt(targetDescription string, maxSeverity pb.DestructionSeverity) string {
	return fmt.Sprintf(`请为以下目标系统生成一个破坏性测试场景：

目标描述：%s
最大严重性：%s

要求：
1. 分析目标系统的潜在脆弱点
2. 设计渐进式的攻击步骤
3. 为每个步骤提供技术实现细节
4. 评估每个步骤的风险和影响
5. 提供场景执行的安全建议

请以JSON格式返回完整的攻击场景。`, targetDescription, maxSeverity.String())
}

// callDeepSeekAPI makes the actual API call to DeepSeek
func (c *DeepSeekClient) callDeepSeekAPI(ctx context.Context, systemPrompt, userPrompt, model string) (*AttackScenario, error) {
	if model == "" {
		model = c.config.Model
	}

	// Prepare request
	reqData := DeepSeekRequest{
		Model: model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      false,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var deepSeekResp DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepSeekResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deepSeekResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Parse the AI-generated scenario
	scenario, err := c.parseScenarioFromContent(deepSeekResp.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scenario: %w", err)
	}

	// Add metadata
	scenario.ID = fmt.Sprintf("scenario_%d", time.Now().UnixNano())

	c.logger.WithFields(logrus.Fields{
		"tokens_used": deepSeekResp.Usage.TotalTokens,
		"model":       deepSeekResp.Model,
	}).Debug("DeepSeek API call completed")

	return scenario, nil
}

// parseScenarioFromContent parses the AI response content into an AttackScenario
func (c *DeepSeekClient) parseScenarioFromContent(content string) (*AttackScenario, error) {
	// Try to parse as JSON first
	var scenario AttackScenario
	if err := json.Unmarshal([]byte(content), &scenario); err == nil {
		return &scenario, nil
	}

	// If JSON parsing fails, try to extract JSON from markdown code blocks
	jsonStart := "```json"
	jsonEnd := "```"

	startIdx := strings.Index(content, jsonStart)
	if startIdx == -1 {
		return nil, fmt.Errorf("no JSON content found in response")
	}

	startIdx += len(jsonStart)
	endIdx := strings.Index(content[startIdx:], jsonEnd)
	if endIdx == -1 {
		return nil, fmt.Errorf("incomplete JSON content in response")
	}

	jsonContent := content[startIdx : startIdx+endIdx]
	if err := json.Unmarshal([]byte(jsonContent), &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse extracted JSON: %w", err)
	}

	return &scenario, nil
}

// parseSeverity converts string severity to protobuf enum
func (c *DeepSeekClient) parseSeverity(severity string) pb.DestructionSeverity {
	switch strings.ToUpper(severity) {
	case "LOW":
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW
	case "MEDIUM":
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_MEDIUM
	case "HIGH":
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_HIGH
	case "CRITICAL":
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_CRITICAL
	default:
		return pb.DestructionSeverity_DESTRUCTION_SEVERITY_LOW
	}
}

// parseDestructionType converts string type to protobuf enum
func (c *DeepSeekClient) parseDestructionType(destructionType string) pb.DestructionType {
	switch strings.ToUpper(destructionType) {
	case "FILE_DELETION":
		return pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION
	case "SERVICE_TERMINATION":
		return pb.DestructionType_DESTRUCTION_TYPE_SERVICE_TERMINATION
	case "MEMORY_EXHAUSTION":
		return pb.DestructionType_DESTRUCTION_TYPE_MEMORY_EXHAUSTION
	case "DISK_FILL":
		return pb.DestructionType_DESTRUCTION_TYPE_DISK_FILL
	case "NETWORK_DISRUPTION":
		return pb.DestructionType_DESTRUCTION_TYPE_NETWORK_DISRUPTION
	case "BOOT_CORRUPTION":
		return pb.DestructionType_DESTRUCTION_TYPE_BOOT_CORRUPTION
	case "KERNEL_PANIC":
		return pb.DestructionType_DESTRUCTION_TYPE_KERNEL_PANIC
	default:
		return pb.DestructionType_DESTRUCTION_TYPE_FILE_DELETION
	}
}

// ValidateScenario validates a generated attack scenario
func (c *DeepSeekClient) ValidateScenario(scenario *AttackScenario, maxSeverity pb.DestructionSeverity) error {
	// Check severity limits
	scenarioSeverity := c.parseSeverity(scenario.Severity)
	if scenarioSeverity > maxSeverity {
		return fmt.Errorf("scenario severity %s exceeds maximum %s", scenario.Severity, maxSeverity.String())
	}

	// Validate steps
	if len(scenario.Steps) == 0 {
		return fmt.Errorf("scenario must have at least one step")
	}

	// Check for dangerous targets
	dangerousTargets := []string{"/bin", "/usr", "/etc", "/var", "/root", "C:\\Windows", "C:\\System32", "C:\\Program Files"}
	for _, step := range scenario.Steps {
		for _, target := range step.Targets {
			for _, dangerous := range dangerousTargets {
				if strings.HasPrefix(target, dangerous) {
					return fmt.Errorf("scenario targets dangerous system path: %s", target)
				}
			}
		}
	}

	return nil
}
