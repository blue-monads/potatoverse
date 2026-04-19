package bllm

import (
	"context"
	"encoding/json"

	openrouter "github.com/revrost/go-openrouter"
	luajit "github.com/yuin/gopher-lua"
)

type Provider struct {
	client *openrouter.Client
	model  string
}

func New(L *luajit.LState) int {
	cfg := L.CheckTable(1)
	if cfg == nil {
		L.RaiseError("provider config table is required")
		return 0
	}

	model := "google/gemini-3-flash-preview"
	apiKey := ""

	modelVal := cfg.RawGetString("model")
	if modelVal != luajit.LNil && modelVal.Type() == luajit.LTString {
		model = modelVal.String()
	}

	apiKeyVal := cfg.RawGetString("apiKey")
	if apiKeyVal != luajit.LNil && apiKeyVal.Type() == luajit.LTString && apiKeyVal.String() != "" {
		apiKey = apiKeyVal.String()
	}

	client := openrouter.NewClientWithConfig(*openrouter.DefaultConfig(apiKey))

	provider := &Provider{
		client: client,
		model:  model,
	}

	luser := L.NewUserData()
	luser.Value = provider
	L.SetMetatable(luser, L.GetTypeMetatable("bllm_provider"))
	L.Push(luser)
	return 1
}

func (p *Provider) GenerateV1(L *luajit.LState) int {
	cfg := L.CheckTable(2)
	if cfg == nil {
		L.RaiseError("generate request table is required")
		return 0
	}

	var messages []openrouter.ChatCompletionMessage
	messagesVal := cfg.RawGetString("messages")
	if messagesVal != luajit.LNil && messagesVal.Type() == luajit.LTTable {
		messageTable := messagesVal.(*luajit.LTable)
		for i := 1; ; i++ {
			entry := messageTable.RawGetInt(i)
			if entry == luajit.LNil || entry.Type() != luajit.LTTable {
				break
			}

			msgMap := entry.(*luajit.LTable)
			roleVal := msgMap.RawGetString("role")
			contentVal := msgMap.RawGetString("content")

			if roleVal == luajit.LNil || contentVal == luajit.LNil {
				continue
			}

			role := roleVal.String()
			content := contentVal.String()

			var msg openrouter.ChatCompletionMessage
			switch role {
			case "system":
				msg = openrouter.SystemMessage(content)
			case "user":
				msg = openrouter.UserMessage(content)
			case "assistant":
				msg = openrouter.AssistantMessage(content)
			default:
				msg = openrouter.UserMessage(content)
			}
			messages = append(messages, msg)
		}
	}

	var tools []openrouter.Tool
	toolsVal := cfg.RawGetString("tools")
	if toolsVal != luajit.LNil && toolsVal.Type() == luajit.LTTable {
		toolTable := toolsVal.(*luajit.LTable)
		for i := 1; ; i++ {
			entry := toolTable.RawGetInt(i)
			if entry == luajit.LNil || entry.Type() != luajit.LTTable {
				break
			}

			toolMap := entry.(*luajit.LTable)
			descVal := toolMap.RawGetString("description")
			paramsVal := toolMap.RawGetString("parameters")

			if descVal == luajit.LNil || paramsVal == luajit.LNil {
				continue
			}

			def := openrouter.FunctionDefinition{
				Name:        toolTable.RawGetInt(i - 1).String(),
				Description: descVal.String(),
			}

			switch paramsVal.Type() {
			case luajit.LTTable:
				paramsBytes, err := tableToJson(paramsVal.(*luajit.LTable))
				if err == nil {
					json.Unmarshal(paramsBytes, &def.Parameters)
				}
			default:
				def.Parameters = paramsVal.String()
			}

			fnDef := def
			tools = append(tools, openrouter.Tool{
				Function: &fnDef,
				Type:     openrouter.ToolTypeFunction,
			})
		}
	}

	req := openrouter.ChatCompletionRequest{
		Model:    p.model,
		Messages: messages,
		Tools:    tools,
	}

	resp, err := p.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		L.RaiseError("chat completion failed: %v", err)
		return 0
	}

	if len(resp.Choices) == 0 {
		L.NewTable()
		return 1
	}

	result := resp.Choices[0].Message.Content.Text
	tableRef := L.Get(-1)
	L.RawSet(tableRef.(*luajit.LTable), luajit.LString("content"), luajit.LString(result))

	return 1
}

func tableToJson(tb *luajit.LTable) ([]byte, error) {
	data := make(map[string]interface{})
	for i := 1; ; i++ {
		key := tb.RawGetInt(i)
		if key == luajit.LNil {
			break
		}
		value := tb.RawGet(key)

		switch value.Type() {
		case luajit.LTString:
			data[key.String()] = value.String()
		case luajit.LTNumber:
			data[key.String()] = float64(value.(luajit.LNumber))
		default:
			data[key.String()] = value.String()
		}
	}

	return json.Marshal(data)
}

func RegisterTypes(L *luajit.LState) {
	providerType := L.NewTypeMetatable("bllm_provider")
	L.SetGlobal("bllm_provider", providerType)

	providerIndex := L.NewTable()
	providerFn := L.NewFunction(func(L *luajit.LState) int {
		p := L.CheckUserData(1).Value.(*Provider)
		return p.GenerateV1(L)
	})
	providerIndex.RawSetString("__index", providerIndex)
	providerIndex.RawSetString("generate_v1", providerFn)
	L.SetField(providerType, "__index", providerIndex)
}

func Init(L *luajit.LState) error {
	RegisterTypes(L)

	llm := L.NewTable()
	newFn := L.NewFunction(New)
	llm.RawSetString("new", newFn)
	L.SetGlobal("llm", llm)
	return nil
}
