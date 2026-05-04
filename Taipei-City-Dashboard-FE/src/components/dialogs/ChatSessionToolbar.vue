<script setup>
defineProps({
	sessions: { type: Array, required: true },
	currentSessionId: { type: String, default: "" },
	loading: { type: Boolean, default: false },
	busy: { type: Boolean, default: false },
});

const emit = defineEmits(["select", "create", "rename", "delete"]);
</script>

<template>
  <div class="chatsessiontoolbar">
    <div class="chatsessiontoolbar-selectwrap">
      <label
        class="chatsessiontoolbar-label"
        for="chat-session-select"
      >對話 session</label>
      <select
        id="chat-session-select"
        :value="currentSessionId"
        :disabled="loading || busy || !sessions.length"
        @change="emit('select', $event.target.value)"
      >
        <option
          v-if="!sessions.length"
          value=""
        >
          尚未建立對話
        </option>
        <option
          v-for="session in sessions"
          :key="session.session"
          :value="session.session"
        >
          {{ session.title }}
        </option>
      </select>
    </div>
    <div class="chatsessiontoolbar-actions">
      <button
        type="button"
        :disabled="loading || busy"
        title="新增對話"
        aria-label="新增對話"
        @click="emit('create')"
      >
        <span class="material-icons-round">add</span>
      </button>
      <button
        type="button"
        :disabled="loading || busy || !currentSessionId"
        title="重新命名"
        aria-label="重新命名"
        @click="emit('rename')"
      >
        <span class="material-icons-round">edit</span>
      </button>
      <button
        type="button"
        :disabled="loading || busy || !currentSessionId"
        title="刪除對話"
        aria-label="刪除對話"
        @click="emit('delete')"
      >
        <span class="material-icons-round">delete</span>
      </button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.chatsessiontoolbar {
	padding: 0.8rem 1rem 0;
	background: #151719;
	display: grid;
	grid-template-columns: minmax(0, 1fr) auto;
	gap: 0.75rem;
	align-items: end;
}

.chatsessiontoolbar-selectwrap {
	min-width: 0;
	display: flex;
	flex-direction: column;
	gap: 0.3rem;
}

.chatsessiontoolbar-label {
	font-size: 12px;
	font-weight: 700;
	color: #a8b0ba;
}

.chatsessiontoolbar select {
	height: 38px;
	width: 100%;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	color: var(--color-normal-text);
	padding: 0 0.7rem;
	font-weight: 700;
	min-width: 0;
}

.chatsessiontoolbar-actions {
	display: flex;
	gap: 0.45rem;
	flex-shrink: 0;
}

.chatsessiontoolbar button {
	width: 38px;
	height: 38px;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	color: var(--color-normal-text);
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	transition: border-color 0.2s, color 0.2s, background 0.2s;
}

.chatsessiontoolbar button:hover,
.chatsessiontoolbar button:focus-visible,
.chatsessiontoolbar select:focus-visible {
	border-color: var(--color-highlight);
	color: var(--color-highlight);
	outline: none;
}

.chatsessiontoolbar button:disabled,
.chatsessiontoolbar select:disabled {
	opacity: 0.45;
	cursor: not-allowed;
}

@media (max-width: 760px) {
	.chatsessiontoolbar {
		grid-template-columns: minmax(0, 1fr);
	}

	.chatsessiontoolbar-actions {
		justify-content: flex-end;
	}
}
</style>