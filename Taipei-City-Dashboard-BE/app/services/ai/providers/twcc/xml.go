package twcc

import (
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

func extractXMLToolCalls(text string) ([]llms.ToolCall, string) {
	var toolCalls []llms.ToolCall
	remaining := text

	for {
		startTag, startIdx := findToolStart(remaining)
		if startIdx == -1 {
			break
		}

		nameEndIdx := strings.Index(remaining[startIdx:], ">")
		if nameEndIdx == -1 {
			break
		}
		nameEndIdx += startIdx

		endTag := "</function>"
		endIdx := strings.Index(remaining[nameEndIdx:], endTag)
		if endIdx == -1 {
			break
		}
		endIdx += nameEndIdx

		funcName := remaining[startIdx+len(startTag) : nameEndIdx]
		args := remaining[nameEndIdx+1 : endIdx]
		toolCalls = append(toolCalls, llms.ToolCall{
			ID:   fmt.Sprintf("call_%d", time.Now().UnixNano()),
			Type: "function",
			FunctionCall: &llms.FunctionCall{
				Name:      funcName,
				Arguments: args,
			},
		})
		remaining = remaining[:startIdx] + remaining[endIdx+len(endTag):]
	}
	return toolCalls, strings.TrimSpace(remaining)
}

func cleanXML(text string) string {
	for {
		_, startIdx := findToolStart(text)
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(text[startIdx:], ">")
		if endIdx == -1 {
			break
		}
		text = text[:startIdx] + text[startIdx+endIdx+1:]
	}
	return strings.ReplaceAll(text, "</function>", "")
}

func findToolStart(text string) (string, int) {
	startTag := "<function="
	startIdx := strings.Index(text, startTag)
	if startIdx != -1 {
		return startTag, startIdx
	}
	startTag = "tool<function="
	return startTag, strings.Index(text, startTag)
}
