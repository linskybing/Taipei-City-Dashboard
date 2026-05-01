<script setup>
import { nextTick, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import ChatAssistantControls from "./ChatAssistantControls.vue";
import ChatComponentCanvas from "./ChatComponentCanvas.vue";
import ChatInputBar from "./ChatInputBar.vue";
import ChatMessage from "./ChatMessage.vue";
import ChatSessionToolbar from "./ChatSessionToolbar.vue";
import ChatStickyNotice from "./ChatStickyNotice.vue";
import ChatWindowHeader from "./ChatWindowHeader.vue";
import { useChatDashboardActions } from "./useChatDashboardActions";
import { useChatSessionControls } from "./useChatSessionControls";
import { useChatComponentCanvas } from "./useChatComponentCanvas";
import { useAuthStore } from "../../store/authStore";
import { useChatStore } from "../../store/chatStore";
import { useContentStore } from "../../store/contentStore";

defineProps({
	standalone: { type: Boolean, default: false },
});

const chatStore = useChatStore();
const contentStore = useContentStore();
const authStore = useAuthStore();
const {
	addChatData,
	addQueryData,
	bootstrapSessions,
	createSession,
	deleteCurrentSession,
	renameCurrentSession,
	selectSession,
	setChatSettings,
} = chatStore;
const { createDashboard } = contentStore;
const {
	chatData,
	chatSettings,
	currentSessionId,
	currentSessionTitle,
	historyLoading,
	sessionBusy,
	sessionList,
	sessionsLoading,
} = storeToRefs(chatStore);
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
const {
	handleCreateSession,
	handleDeleteSession,
	handleRenameSession,
	handleSessionSelect,
	inputPlaceholder,
	isInputDisabled,
	isSessionToolbarBusy,
} = useChatSessionControls({
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
});
const { qaBtnHandler } = useChatDashboardActions({
	addChatData,
	createDashboard,
	dashboardCreationLoading,
	editDashboard,
	user,
});

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
      <ChatSessionToolbar
        :sessions="sessionList"
        :current-session-id="currentSessionId"
        :loading="sessionsLoading"
        :busy="isSessionToolbarBusy"
        @select="handleSessionSelect"
        @create="handleCreateSession"
        @rename="handleRenameSession"
        @delete="handleDeleteSession"
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
        :disabled="isInputDisabled"
        :placeholder="inputPlaceholder"
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
