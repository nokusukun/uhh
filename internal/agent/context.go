package agent

import (
	"github.com/tmc/langchaingo/llms"
)

// Message represents a conversation message
type Message struct {
	Role      string
	Content   string
	ToolID    string           // For tool responses
	ToolCalls []llms.ToolCall  // For assistant messages with tool calls
}

// Context manages conversation history for the agent
type Context struct {
	SystemPrompt string
	Messages     []Message
	MaxMessages  int
}

// NewContext creates a new conversation context
func NewContext(systemPrompt string) *Context {
	return &Context{
		SystemPrompt: systemPrompt,
		Messages:     make([]Message, 0),
		MaxMessages:  50, // Keep last 50 messages
	}
}

// AddUserMessage adds a user message to the context
func (c *Context) AddUserMessage(content string) {
	c.Messages = append(c.Messages, Message{
		Role:    "user",
		Content: content,
	})
	c.truncate()
}

// AddAssistantMessage adds an assistant message to the context
func (c *Context) AddAssistantMessage(content string) {
	c.Messages = append(c.Messages, Message{
		Role:    "assistant",
		Content: content,
	})
	c.truncate()
}

// AddAssistantMessageWithToolCalls adds an assistant message with tool calls
func (c *Context) AddAssistantMessageWithToolCalls(content string, toolCalls []llms.ToolCall) {
	c.Messages = append(c.Messages, Message{
		Role:      "assistant",
		Content:   content,
		ToolCalls: toolCalls,
	})
	c.truncate()
}

// AddToolResult adds a tool result to the context
func (c *Context) AddToolResult(toolID, result string) {
	c.Messages = append(c.Messages, Message{
		Role:    "tool",
		Content: result,
		ToolID:  toolID,
	})
	c.truncate()
}

// truncate removes old messages if we exceed the limit
func (c *Context) truncate() {
	if len(c.Messages) > c.MaxMessages {
		// Keep the most recent messages
		c.Messages = c.Messages[len(c.Messages)-c.MaxMessages:]
	}
}

// ToLangchainMessages converts the context to langchaingo message format
func (c *Context) ToLangchainMessages() []llms.MessageContent {
	messages := make([]llms.MessageContent, 0, len(c.Messages)+1)

	// Add system message first
	if c.SystemPrompt != "" {
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeSystem, c.SystemPrompt))
	}

	// Add conversation messages
	for _, msg := range c.Messages {
		switch msg.Role {
		case "user":
			messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, msg.Content))
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				parts := make([]llms.ContentPart, 0, len(msg.ToolCalls)+1)
				if msg.Content != "" {
					parts = append(parts, llms.TextContent{Text: msg.Content})
				}
				for _, tc := range msg.ToolCalls {
					parts = append(parts, tc)
				}
				messages = append(messages, llms.MessageContent{
					Role:  llms.ChatMessageTypeAI,
					Parts: parts,
				})
			} else {
				messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, msg.Content))
			}
		case "tool":
			// Tool results are added as part of the conversation
			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: msg.ToolID,
						Content:    msg.Content,
					},
				},
			})
		}
	}

	return messages
}

// Clear clears all messages from the context
func (c *Context) Clear() {
	c.Messages = make([]Message, 0)
}

// LastMessage returns the last message in the context
func (c *Context) LastMessage() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return &c.Messages[len(c.Messages)-1]
}

// MessageCount returns the number of messages in the context
func (c *Context) MessageCount() int {
	return len(c.Messages)
}
