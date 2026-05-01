import { computed, onMounted, watch } from "vue";

export function useChatSessionControls({
	authStore,
	addChatData,
	bootstrapSessions,
	createSession,
	deleteCurrentSession,
	renameCurrentSession,
	selectSession,
	currentSessionId,
	currentSessionTitle,
	sessionsLoading,
	historyLoading,
	sessionBusy,
	userMessage,
}) {
	const isInputDisabled = computed(
		() => !currentSessionId.value || sessionsLoading.value || historyLoading.value || sessionBusy.value,
	);
	const inputPlaceholder = computed(() =>
		currentSessionId.value ? "輸入城市議題或 component_id..." : "請先建立或選擇一個對話 session...",
	);
	const isSessionToolbarBusy = computed(
		() => sessionsLoading.value || historyLoading.value || sessionBusy.value,
	);

	const handleCreateSession = async () => {
		if (!authStore.token && !localStorage.getItem("token")) {
			addChatData({ role: "bot", content: "請先登入會員後再建立 AI 對話。", error: true });
			return;
		}
		await createSession();
		userMessage.value = "";
	};

	const handleRenameSession = async () => {
		if (!currentSessionId.value) return;
		const nextTitle = window.prompt("請輸入新的對話標題", currentSessionTitle.value || "");
		if (nextTitle === null) return;
		await renameCurrentSession(nextTitle);
		userMessage.value = "";
	};

	const handleDeleteSession = async () => {
		if (!currentSessionId.value) return;
		const targetTitle = currentSessionTitle.value || "這個對話";
		if (!window.confirm(`確定刪除「${targetTitle}」？`)) return;
		await deleteCurrentSession();
		userMessage.value = "";
	};

	const handleSessionSelect = async (sessionId) => {
		if (!sessionId || sessionId === currentSessionId.value) return;
		await selectSession(sessionId);
		userMessage.value = "";
	};

	onMounted(() => {
		bootstrapSessions();
	});

	watch(
		() => authStore.token,
		() => {
			bootstrapSessions();
		},
	);

	return {
		handleCreateSession,
		handleDeleteSession,
		handleRenameSession,
		handleSessionSelect,
		inputPlaceholder,
		isInputDisabled,
		isSessionToolbarBusy,
	};
}