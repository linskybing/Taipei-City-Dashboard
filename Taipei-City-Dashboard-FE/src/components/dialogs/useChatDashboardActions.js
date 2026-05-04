import http from "../../router/axios";

export function useChatDashboardActions({
	addChatData,
	createDashboard,
	dashboardCreationLoading,
	editDashboard,
	user,
}) {
	const qaBtnHandler = async (text, relations) => {
		if (text !== "建立儀表板" || dashboardCreationLoading.value) return;
		dashboardCreationLoading.value = true;
		const response = await http.get("/dashboard/");
		if (response.data?.data?.personal?.length > 20) {
			addChatData({
				role: "bot",
				content: "您的個人儀表板已超出限制 20 個，請先移除既有儀表板後，重新執行本功能！",
			});
			dashboardCreationLoading.value = false;
			return;
		}
		const components = Array.from(new Set((relations || []).map((item) => item.id))).map((id) => ({ id }));
		if (user.value.user_id) {
			editDashboard.value = { index: "", name: "推薦儀表板", icon: "star", components };
			await createDashboard();
		} else {
			addChatData({ role: "bot", content: "請先登入會員以使用此功能喔！" });
		}
		dashboardCreationLoading.value = false;
	};

	return { qaBtnHandler };
}