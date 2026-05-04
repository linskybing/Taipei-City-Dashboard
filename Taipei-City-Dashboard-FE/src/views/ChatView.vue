<script setup>
import { onMounted, onUnmounted } from "vue";
import ChatBox from "../components/dialogs/ChatBox.vue";

onMounted(() => {
	document.body.classList.add("chat-route-active");
});

onUnmounted(() => {
	document.body.classList.remove("chat-route-active");
});
</script>

<template>
  <main class="chatview">
    <ChatBox
      standalone
      class="chatview-box"
    />
  </main>
</template>

<style lang="scss" scoped>
.chatview {
	width: 100vw;
	min-height: 100vh;
	min-height: calc(var(--vh) * 100);
	background: #050607;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 1rem;
}

.chatview-box {
	width: min(960px, calc(100vw - 2rem));
	height: min(860px, calc(100vh - 2rem));
	height: min(860px, calc(var(--vh) * 100 - 2rem));
}

:global(body.chat-route-active .chatbot-container) {
	display: none;
}

:global(body.chat-route-active .navbar) {
	display: none;
}
</style>
