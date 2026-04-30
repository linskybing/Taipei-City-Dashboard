import { ref, watch } from "vue";
import { defineStore } from "pinia";
import http from "../router/axios";
import { useAuthStore } from "./authStore";

const defaultChatData = [
	{
		id: 1,
		role: "bot",
		isDefault: true,
		content:
			"您好，我是臺北城市儀表板 AI 決策助理。請輸入城市議題或 component_id，我會檢索儀表板組件；若要求資料品質、描述統計、趨勢、季節、異常、檢定或預測，會使用對應統計工具整理分析。",
	},
];

const defaultSettings = {
	theme: "auto",
	city: "metrotaipei",
	audience: "government",
	dashboardIndex: "",
};

export const useChatStore = defineStore("chat", () => {
	const savedChatData = JSON.parse(sessionStorage.getItem("chatData")) || [];
	const chatData = ref([...defaultChatData, ...savedChatData]);
	const chatSettings = ref({ ...defaultSettings });
	const authStore = useAuthStore();

	watch(
		chatData,
		(newVal) => {
			const userBotMessages = newVal.filter((item) => !item.isDefault);
			sessionStorage.setItem("chatData", JSON.stringify(userBotMessages));
		},
		{ deep: true },
	);

	const addChatData = (newChatData) => {
		const id = chatData.value.length + 1;
		chatData.value.push({ id, isDefault: false, ...newChatData });
		return id;
	};

	const setChatSettings = (settings) => {
		chatSettings.value = { ...chatSettings.value, ...settings };
	};

	const addQueryData = async (newChatData, settings = {}) => {
		const mergedSettings = { ...chatSettings.value, ...settings };
		setChatSettings(mergedSettings);
		addChatData({ role: "user", content: newChatData.content });

		if (!authStore.token && !localStorage.getItem("token")) {
			addChatData({
				role: "bot",
				content: "請先登入會員後再使用 AI 決策助理。",
				error: true,
			});
			return;
		}

		const loadingId = addChatData({ role: "bot", content: "正在整理儀表板訊號與資料來源..." });

		try {
			const response = await http.post("/ai/chat/twai", {
				session: getSessionID(),
				stream: false,
				theme: mergedSettings.theme,
				city: mergedSettings.city,
				audience: mergedSettings.audience,
				dashboard_index: mergedSettings.dashboardIndex,
				messages: [{ role: "user", content: newChatData.content }],
			});
			updateBotMessage(loadingId, buildAssistantMessage(response.data?.data));
		} catch (error) {
			console.error("AIChatError:", error);
			updateBotMessage(loadingId, {
				role: "bot",
				content: buildAssistantErrorMessage(error),
				error: true,
			});
		}
	};

	const updateBotMessage = (id, payload) => {
		const index = chatData.value.findIndex((item) => item.id === id);
		if (index < 0) return;
		chatData.value[index] = { id, isDefault: false, ...payload };
	};

	const saveChatLog = async (question, answer) => {
		try {
			const formData = new FormData();
			formData.append("session", getSessionID());
			formData.append("question", question);
			formData.append("answer", JSON.stringify(answer));
			await http.post("/chatlog/", formData, {
				headers: { "Content-Type": "multipart/form-data" },
			});
		} catch (error) {
			console.error("saveChatLog error:", error);
		}
	};

	return {
		chatData,
		chatSettings,
		addChatData,
		addQueryData,
		saveChatLog,
		setChatSettings,
	};
});

function buildAssistantMessage(data = {}) {
	const relations = data.related_components || [];
	return {
		role: "bot",
		content: data.content || "目前沒有足夠資料形成判讀。",
		relations,
		sources: data.sources || [],
		actions: data.recommended_actions || [],
		confidenceNotes: data.confidence_notes || [],
		analysisCards: data.analysis_cards || [],
		visualizations: data.visualization_refs || [],
		button: relations.length > 0 ? [{ id: 1, text: "建立儀表板" }] : null,
	};
}

function buildAssistantErrorMessage(error) {
	const status = error?.response?.status;
	if (status === 401 || status === 403) {
		return "登入狀態已失效，請重新登入後再使用 AI 決策助理。";
	}
	return "目前無法取得 AI 決策助理回覆，請稍後再試或改以更具體的議題描述查詢。";
}

function getSessionID() {
	const d = new Date();
	const todayId =
		d.getFullYear() +
		String(d.getMonth() + 1).padStart(2, "0") +
		String(d.getDate()).padStart(2, "0");
	return `session_${todayId}`;
}
