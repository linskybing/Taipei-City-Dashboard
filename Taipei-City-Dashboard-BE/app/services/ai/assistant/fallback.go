package assistant

import "fmt"

func unavailableTool(name string, err error) (string, error) {
	return marshalTool(toolEnvelope{
		Tool: name,
		ConfidenceNotes: []string{
			fmt.Sprintf("工具 %s 暫時無法取得後端資料，請以既有儀表板訊號保守判讀。", name),
		},
		Data: map[string]interface{}{
			"status": "unavailable",
			"error":  err.Error(),
		},
	})
}
