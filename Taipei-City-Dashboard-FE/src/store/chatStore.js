import { ref } from "vue";
import { defineStore } from "pinia";
import http from "../router/axios";
import { useAuthStore } from "./authStore";
import {
	buildAssistantErrorMessage,
	buildAssistantMessage,
	buildHistoryChatMessage,
	defaultChatData,
	defaultSettings,
} from "./chatSessionHelpers";
import { createChatSessionActions } from "./chatSessionActions";

const currentSessionStorageKey = "chatCurrentSessionId";

let localMessageSequence = 0;

function nextMessageId(prefix = "chat") {
	localMessageSequence += 1;
	return `${prefix}_${Date.now()}_${localMessageSequence}`;
}

export const useChatStore = defineStore("chat", () => {
	const chatData = ref([...defaultChatData]);
	const chatSettings = ref({ ...defaultSettings });
	const sessionList = ref([]);
	const currentSessionId = ref(sessionStorage.getItem(currentSessionStorageKey) || "");
	const currentSessionTitle = ref("");
	const sessionsLoading = ref(false);
	const sessionBusy = ref(false);
	const historyLoading = ref(false);
	const authStore = useAuthStore();

	const hasAIChatAccess = () => Boolean(authStore.token || localStorage.getItem("token"));

	const resetCurrentSession = () => {
		currentSessionId.value = "";
		currentSessionTitle.value = "";
		sessionStorage.removeItem(currentSessionStorageKey);
		chatData.value = [...defaultChatData];
	};

	const setCurrentSession = (session = {}) => {
		currentSessionId.value = session.session || "";
		currentSessionTitle.value = session.title || "";
		if (currentSessionId.value) {
			sessionStorage.setItem(currentSessionStorageKey, currentSessionId.value);
			return;
		}
		sessionStorage.removeItem(currentSessionStorageKey);
	};

	const replaceChatHistory = (messages = []) => {
		chatData.value = [
			...defaultChatData,
			...messages.map((message) => ({
				...buildHistoryChatMessage(message),
				id: message.id || nextMessageId(message.role || "history"),
			})),
		];
	};

	const addChatData = (newChatData) => {
		const id = newChatData.id || nextMessageId(newChatData.role || "chat");
		chatData.value = [...chatData.value, { id, isDefault: false, ...newChatData }];
		return id;
	};

	const setChatSettings = (settings) => {
		chatSettings.value = { ...chatSettings.value, ...settings };
	};

	const {
		bootstrapSessions,
		selectSession,
		createSession,
		renameCurrentSession,
		deleteCurrentSession,
	} = createChatSessionActions({
		http,
		hasAIChatAccess,
		sessionList,
		currentSessionId,
		sessionsLoading,
		sessionBusy,
		historyLoading,
		chatData,
		defaultChatData,
		resetCurrentSession,
		setCurrentSession,
		replaceChatHistory,
	});

	const addQueryData = async (newChatData, settings = {}) => {
		const mergedSettings = { ...chatSettings.value, ...settings };
		setChatSettings(mergedSettings);

		if (!hasAIChatAccess()) {
			addChatData({
				role: "bot",
				content: "請先登入會員後再使用 AI 決策助理。",
				error: true,
			});
			return false;
		}
		if (!currentSessionId.value) {
			addChatData({
				role: "bot",
				content: "請先建立一個新的對話 session，再開始提問。",
				error: true,
			});
			return false;
		}
		addChatData({ role: "user", content: newChatData.content });

		const loadingId = addChatData({ role: "bot", content: "正在整理儀表板訊號與資料來源..." });

		try {
			const response = await http.post("/ai/chat/twai", {
				session: currentSessionId.value,
				stream: false,
				theme: mergedSettings.theme,
				city: mergedSettings.city,
				audience: mergedSettings.audience,
				dashboard_index: mergedSettings.dashboardIndex,
				messages: [{ role: "user", content: newChatData.content }],
			});
			updateBotMessage(loadingId, { id: loadingId, ...buildAssistantMessage(response.data?.data) });
			const currentSummary = sessionList.value.find((item) => item.session === currentSessionId.value) || {};
			const refreshedSession = {
				...currentSummary,
				session: currentSessionId.value,
				title: currentSessionTitle.value || currentSummary.title || "新對話",
				status: "active",
				last_activity_at: new Date().toISOString(),
			};
			sessionList.value = [
				refreshedSession,
				...sessionList.value.filter((item) => item.session !== currentSessionId.value),
			];
			return true;
		} catch (error) {
			updateBotMessage(loadingId, {
				role: "bot",
				content: buildAssistantErrorMessage(error),
				error: true,
			});
			if (error?.response?.status === 404 || error?.response?.status === 410) {
				await bootstrapSessions();
			}
			return false;
		}
	};

	const updateBotMessage = (id, payload) => {
		const index = chatData.value.findIndex((item) => item.id === id);
		if (index < 0) return;
		chatData.value = chatData.value.map((item) =>
			item.id === id ? { id, isDefault: false, ...payload } : item,
		);
	};

	return {
		chatData,
		chatSettings,
		sessionList,
		currentSessionId,
		currentSessionTitle,
		sessionsLoading,
		sessionBusy,
		historyLoading,
		addChatData,
		addQueryData,
		bootstrapSessions,
		selectSession,
		createSession,
		renameCurrentSession,
		deleteCurrentSession,
		setChatSettings,
	};
});
