<script setup>
import BotLogo from "../icons/BotLogo.vue";
import UserLogo from "../icons/UserLogo.vue";
import ChatAssistantDetails from "./ChatAssistantDetails.vue";
import ChatRelationTable from "./ChatRelationTable.vue";
import ChatVisualizationRefs from "./ChatVisualizationRefs.vue";

defineProps({
	chat: { type: Object, required: true },
});

const emit = defineEmits(["action", "show-components"]);
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
          :analysis-cards="chat.analysisCards"
          :sources="chat.sources"
          :confidence-notes="chat.confidenceNotes"
        />
        <ChatVisualizationRefs :items="chat.visualizations" />
        <ChatRelationTable
          :relations="chat.relations"
          @show="emit('show-components', $event)"
        />
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
	padding: 0.45rem 0;
}

.bot,
.user {
	display: flex;
	gap: 0.7rem;
	align-items: flex-start;
}

.user {
	flex-direction: row-reverse;
}

.avatar {
	width: 34px;
	height: 34px;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
}

.avatar svg {
	width: 24px;
	height: auto;
}

.content {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.bot .content {
	flex: 1;
	min-width: 0;
}

.user .content {
	max-width: min(70%, 720px);
}

.message--bubble {
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	color: var(--color-normal-text);
}

.message--bubble p {
	white-space: pre-line;
	margin: 0;
	padding: 0.75rem 0.9rem;
	font-size: 15px;
	line-height: 1.6;
}

.user .message--bubble {
	border-color: #2d6fd2;
	background: #163d75;
}

.user .message--bubble p {
	color: #ffffff;
}

.message--button {
	display: flex;
	gap: 0.5rem;
	overflow-x: auto;
}

.message--button button {
	flex-shrink: 0;
	border: 1px solid #3d4651;
	background: #20252b;
	color: #ffffff;
	font-size: 14px;
	padding: 0.5rem 1rem;
	border-radius: 8px;
	cursor: pointer;
	white-space: nowrap;
	transition: border-color 0.2s, color 0.2s;
}

.message--button button:hover,
.message--button button:focus-visible {
	border-color: var(--color-highlight);
	color: var(--color-highlight);
	outline: none;
}
</style>
