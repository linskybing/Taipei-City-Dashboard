<script setup>
import BotLogo from "../icons/BotLogo.vue";
import UserLogo from "../icons/UserLogo.vue";
import ChatAssistantDetails from "./ChatAssistantDetails.vue";
import ChatRelationTable from "./ChatRelationTable.vue";

defineProps({
	chat: { type: Object, required: true },
});

const emit = defineEmits(["action"]);
</script>

<template>
  <div class="message">
    <div
      v-if="chat.role === 'bot'"
      class="bot"
    >
      <div class="avatar">
        <BotLogo />
      </div>
      <div class="content">
        <div
          v-if="chat.content"
          class="message--bubble"
        >
          <p>{{ chat.content }}</p>
        </div>
        <ChatAssistantDetails
          :actions="chat.actions"
          :sources="chat.sources"
          :confidence-notes="chat.confidenceNotes"
        />
        <ChatRelationTable :relations="chat.relations" />
        <div
          v-if="chat.button"
          v-horizontal-wheel
          class="message--button scrollbar-x-hide"
        >
          <button
            v-for="btn in chat.button"
            :key="btn.id"
            @click="emit('action', btn.text, chat.relations)"
          >
            {{ btn.text }}
          </button>
        </div>
      </div>
    </div>
    <div
      v-else
      class="user"
    >
      <div class="avatar">
        <UserLogo />
      </div>
      <div
        v-if="chat.content"
        class="content"
      >
        <div class="message--bubble">
          <p>{{ chat.content }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.scrollbar-x-hide {
	scrollbar-width: none;

	&::-webkit-scrollbar {
		display: none;
	}
}

.message {
	padding: 8px;
}

.bot,
.user {
	display: flex;
	gap: 0.5rem;
	align-items: flex-start;
}

.user {
	flex-direction: row-reverse;
}

.avatar {
	width: 40px;
	height: 40px;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

.avatar svg {
	width: 100%;
	height: auto;
}

.content {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.message--bubble {
	border: 1px solid #ffffff;
	border-radius: 10px;
	background: #282a2c;
	color: #ffffff;
}

.message--bubble p {
	white-space: pre-line;
	margin: 0;
	padding: 8px 16px;
	font-size: 16px;
}

.message--button {
	display: flex;
	gap: 0.5rem;
	overflow-x: auto;
}

.message--button button {
	flex-shrink: 0;
	background: #494b4e;
	color: #ffffff;
	font-size: 14px;
	padding: 0.5rem 1rem;
	border-radius: 15px;
	border: none;
	cursor: pointer;
	white-space: nowrap;
}

.message--button button:hover {
	filter: brightness(0.5);
}
</style>
