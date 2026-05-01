export const defaultChatData = [
	{
		id: "default-welcome",
		role: "bot",
		isDefault: true,
		content:
			"您好，我是臺北城市儀表板 AI 決策助理。請先建立或選擇一個對話 session，再輸入城市議題或 component_id。",
	},
];

export const defaultSettings = {
	theme: "auto",
	city: "metrotaipei",
	audience: "government",
	dashboardIndex: "",
};

export function buildAssistantMessage(data = {}) {
	const relations = data.related_components || data.relations || [];
	return {
		role: "bot",
		content: data.content || "目前沒有足夠資料形成判讀。",
		relations,
		sources: data.sources || [],
		actions: data.recommended_actions || data.actions || [],
		confidenceNotes: data.confidence_notes || data.confidenceNotes || [],
		analysisCards: data.analysis_cards || data.analysisCards || [],
		visualizations: data.visualization_refs || data.visualizations || [],
		button: relations.length > 0 ? [{ id: `${data.id || "assistant"}:build`, text: "建立儀表板" }] : null,
	};
}

export function buildHistoryChatMessage(message = {}) {
	if (message.role === "assistant") {
		return {
			id: message.id,
			isDefault: false,
			...buildAssistantMessage(message),
			createdAt: message.created_at,
		};
	}
	return {
		id: message.id,
		isDefault: false,
		role: "user",
		content: message.content || "",
		createdAt: message.created_at,
	};
}

export function buildAssistantErrorMessage(error) {
	const status = error?.response?.status;
	if (status === 401 || status === 403) {
		return "登入狀態已失效，請重新登入後再使用 AI 決策助理。";
	}
	if (status === 410) {
		return "此對話 session 已經刪除，請重新建立或切換其他對話。";
	}
	return "目前無法取得 AI 決策助理回覆，請稍後再試或改以更具體的議題描述查詢。";
}