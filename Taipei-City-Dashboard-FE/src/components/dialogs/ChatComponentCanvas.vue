<script setup>
import ChatComponentCard from "./ChatComponentCard.vue";

defineProps({
	open: { type: Boolean, default: false },
	items: { type: Array, default: () => [] },
	creating: { type: Boolean, default: false },
});

const emit = defineEmits(["close", "create-dashboard"]);
</script>

<template>
  <aside
    v-if="open && items.length"
    class="component-canvas"
    aria-label="組件參考畫布"
  >
    <header>
      <div>
        <p>組件參考</p>
        <strong>已整理 {{ items.length }} 個相關組件</strong>
      </div>
      <button
        type="button"
        aria-label="收合組件參考畫布"
        @click="emit('close')"
      >
        ×
      </button>
    </header>
    <div class="canvas-list scrollbar-custom">
      <ChatComponentCard
        v-for="item in items"
        :key="item.key"
        :item="item"
      />
    </div>
    <footer>
      <button
        type="button"
        :disabled="creating"
        @click="emit('create-dashboard')"
      >
        {{ creating ? "建立中..." : "建立儀表板" }}
      </button>
    </footer>
  </aside>
</template>

<style lang="scss" scoped>
.component-canvas {
	box-sizing: border-box;
	width: min(360px, 34vw);
	height: 100%;
	border: 1px solid #34383d;
	border-radius: 12px;
	background: #101418;
	box-shadow: 0 18px 48px rgb(0 0 0 / 38%);
	display: flex;
	flex-direction: column;
	overflow: hidden;
}

header,
footer {
	background: #151a1f;
	border-bottom: 1px solid #2d343c;
	padding: 0.85rem;
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.75rem;
}

footer {
	border-top: 1px solid #2d343c;
	border-bottom: 0;
}

p,
strong {
	margin: 0;
}

p {
	color: #8d99a6;
	font-size: 12px;
}

strong {
	color: #ffffff;
	font-size: 15px;
}

button {
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	color: #ffffff;
	cursor: pointer;
	font-weight: 700;
	transition: border-color 0.2s, color 0.2s, background 0.2s;
}

header button {
	width: 32px;
	height: 32px;
	font-size: 22px;
	line-height: 1;
}

footer button {
	width: 100%;
	min-height: 40px;
	font-size: 14px;
}

button:hover,
button:focus-visible {
	border-color: var(--color-highlight);
	color: var(--color-highlight);
	outline: none;
}

button:disabled {
	cursor: wait;
	opacity: 0.68;
}

.canvas-list {
	flex: 1;
	overflow-y: auto;
	padding: 0.85rem;
	display: flex;
	flex-direction: column;
	gap: 0.7rem;
}

.scrollbar-custom::-webkit-scrollbar {
	width: 6px;
	background: transparent;
}

.scrollbar-custom::-webkit-scrollbar-thumb {
	background: #4b5563;
	border-radius: 8px;
}

@media (max-width: 900px) {
	.component-canvas {
		position: absolute;
		inset: auto 0 0 0;
		width: 100%;
		height: min(72%, 560px);
		z-index: 4;
		border-radius: 12px 12px 0 0;
		box-shadow: 0 -16px 42px rgb(0 0 0 / 48%);
	}
}
</style>
