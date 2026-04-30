<script setup>
import { nextTick, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import ChatAssistantControls from "./ChatAssistantControls.vue";
import ChatComponentCanvas from "./ChatComponentCanvas.vue";
import ChatInputBar from "./ChatInputBar.vue";
import ChatMessage from "./ChatMessage.vue";
import ChatStickyNotice from "./ChatStickyNotice.vue";
import ChatWindowHeader from "./ChatWindowHeader.vue";
import { useChatComponentCanvas } from "./useChatComponentCanvas";
import { useAuthStore } from "../../store/authStore";
import { useChatStore } from "../../store/chatStore";
import { useContentStore } from "../../store/contentStore";
import http from "../../router/axios";

defineProps({
	standalone: { type: Boolean, default: false },
});

const chatStore = useChatStore();
const contentStore = useContentStore();
const authStore = useAuthStore();
const { addChatData, addQueryData, saveChatLog, setChatSettings } = chatStore;
const { createDashboard } = contentStore;
const { chatData, chatSettings } = storeToRefs(chatStore);
const { editDashboard } = storeToRefs(contentStore);
const { user } = storeToRefs(authStore);

const userMessage = ref("");
const chatAreaRef = ref(null);
const isStickyOpen = ref(false);
const dashboardCreationLoading = ref(false);
const selectedTheme = ref(chatSettings.value.theme);
const selectedCity = ref(chatSettings.value.city);
const selectedAudience = ref(chatSettings.value.audience);
const { isCanvasOpen, canvasRelations, canvasItems, openCanvas, closeCanvas } =
	useChatComponentCanvas(chatData);

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
	const components = Array.from(new Set((relations || []).map((item) => item.id))).map(
		(id) => ({ id }),
	);
	if (user.value.user_id) {
		editDashboard.value = {
			index: "",
			name: "推薦儀表板",
			icon: "star",
			components,
		};
		await createDashboard();
		saveChatLog("建立儀表板", "使用者成功建立儀表板!");
	} else {
		addChatData({ role: "bot", content: "請先登入會員以使用此功能喔！" });
	}
	dashboardCreationLoading.value = false;
};

const sendBtnHandler = (text) => {
	if (!text.trim()) return;
	addQueryData(
		{ role: "user", content: text },
		{
			theme: selectedTheme.value,
			city: selectedCity.value,
			audience: selectedAudience.value,
		},
	);
	userMessage.value = "";
};

const toggleSticky = () => {
	isStickyOpen.value = !isStickyOpen.value;
};

watch(
	() => chatData.value.length,
	async () => {
		await nextTick();
		const chat = chatAreaRef.value;
		if (chat) chat.scrollTop = chat.scrollHeight - chat.clientHeight;
	},
	{ deep: true },
);

watch([selectedTheme, selectedCity, selectedAudience], () => {
	setChatSettings({
		theme: selectedTheme.value,
		city: selectedCity.value,
		audience: selectedAudience.value,
	});
});
</script>

<template>
  <div
    class="chat-layout"
    :class="{
      'is-standalone': standalone,
      'has-canvas': isCanvasOpen && canvasItems.length,
    }"
  >
    <div
      class="chat-widget"
      :class="{ 'is-standalone': standalone }"
    >
      <ChatWindowHeader
        :standalone="standalone"
      />
      <div
        ref="chatAreaRef"
        class="chat-area scrollbar-custom"
      >
        <ChatStickyNotice
          :open="isStickyOpen"
          @toggle="toggleSticky"
        />
        <ChatMessage
          v-for="chat in chatData"
          :key="chat.id"
          :chat="chat"
          @action="qaBtnHandler"
          @show-components="openCanvas"
        />
      </div>
      <ChatAssistantControls
        v-model:theme="selectedTheme"
        v-model:city="selectedCity"
        v-model:audience="selectedAudience"
      />
      <ChatInputBar
        v-model="userMessage"
        @send="sendBtnHandler"
      />
    </div>
    <ChatComponentCanvas
      :open="isCanvasOpen"
      :items="canvasItems"
      :creating="dashboardCreationLoading"
      @close="closeCanvas"
      @create-dashboard="qaBtnHandler('建立儀表板', canvasRelations)"
    />
  </div>
</template>

<style lang="scss" scoped src="./ChatBoxLayout.scss"></style>
