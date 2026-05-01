export function createChatSessionActions({
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
}) {
	const bootstrapSessions = async () => {
		if (!hasAIChatAccess()) {
			sessionList.value = [];
			resetCurrentSession();
			return false;
		}
		sessionsLoading.value = true;
		try {
			const response = await http.get("/ai/sessions", { skipGlobalLoading: true, skipErrorNotify: true });
			sessionList.value = response.data?.data || [];
			if (!sessionList.value.length) {
				resetCurrentSession();
				return true;
			}
			const targetSession =
				sessionList.value.find((item) => item.session === currentSessionId.value) || sessionList.value[0];
			return selectSession(targetSession.session, true);
		} catch {
			sessionList.value = [];
			resetCurrentSession();
			return false;
		} finally {
			sessionsLoading.value = false;
		}
	};

	const selectSession = async (sessionId, skipSessionRefresh = false) => {
		if (!hasAIChatAccess() || !sessionId) {
			resetCurrentSession();
			return false;
		}
		historyLoading.value = true;
		try {
			const response = await http.get(`/ai/sessions/${encodeURIComponent(sessionId)}`, {
				skipGlobalLoading: true,
				skipErrorNotify: true,
			});
			const session = response.data?.data?.session || {};
			setCurrentSession(session);
			replaceChatHistory(response.data?.data?.messages || []);
			if (!skipSessionRefresh) {
				sessionList.value = sessionList.value.map((item) =>
					item.session === session.session ? { ...item, ...session } : item,
				);
			}
			return true;
		} catch (error) {
			if (error?.response?.status === 404 || error?.response?.status === 410) {
				await bootstrapSessions();
			}
			return false;
		} finally {
			historyLoading.value = false;
		}
	};

	const createSession = async (title = "") => {
		if (!hasAIChatAccess()) return false;
		sessionBusy.value = true;
		try {
			const response = await http.post("/ai/sessions", title ? { title } : {}, { skipGlobalLoading: true });
			const session = response.data?.data || {};
			sessionList.value = [session, ...sessionList.value.filter((item) => item.session !== session.session)];
			setCurrentSession(session);
			chatData.value = [...defaultChatData];
			return true;
		} catch {
			return false;
		} finally {
			sessionBusy.value = false;
		}
	};

	const renameCurrentSession = async (title) => {
		if (!hasAIChatAccess() || !currentSessionId.value) return false;
		sessionBusy.value = true;
		try {
			const response = await http.patch(
				`/ai/sessions/${encodeURIComponent(currentSessionId.value)}`,
				{ title },
				{ skipGlobalLoading: true },
			);
			const session = response.data?.data || {};
			setCurrentSession(session);
			sessionList.value = sessionList.value.map((item) =>
				item.session === session.session ? { ...item, ...session } : item,
			);
			return true;
		} catch {
			return false;
		} finally {
			sessionBusy.value = false;
		}
	};

	const deleteCurrentSession = async () => {
		if (!hasAIChatAccess() || !currentSessionId.value) return false;
		sessionBusy.value = true;
		try {
			await http.delete(`/ai/sessions/${encodeURIComponent(currentSessionId.value)}`, {
				skipGlobalLoading: true,
			});
			sessionList.value = sessionList.value.filter((item) => item.session !== currentSessionId.value);
			const nextSession = sessionList.value[0];
			if (!nextSession) {
				resetCurrentSession();
				return true;
			}
			return selectSession(nextSession.session, true);
		} catch {
			return false;
		} finally {
			sessionBusy.value = false;
		}
	};

	return {
		bootstrapSessions,
		selectSession,
		createSession,
		renameCurrentSession,
		deleteCurrentSession,
	};
}