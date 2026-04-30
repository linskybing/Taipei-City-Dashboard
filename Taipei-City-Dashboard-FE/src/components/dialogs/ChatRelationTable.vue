<script setup>
defineProps({
	relations: { type: Array, default: () => [] },
});

const emit = defineEmits(["show"]);
</script>

<template>
  <section
    v-if="relations.length"
    class="relation-summary"
  >
    <div>
      <strong>已整理 {{ Math.min(relations.length, 8) }} 個相關組件</strong>
      <span>詳細名稱、來源與 component_id 已放在組件參考畫布。</span>
    </div>
    <button
      type="button"
      @click="emit('show', relations)"
    >
      查看畫布
    </button>
  </section>
</template>

<style lang="scss" scoped>
.relation-summary {
	width: 100%;
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.75rem;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #151a20;
	margin: 0.35rem 0;
	padding: 0.75rem 0.85rem;
}

.relation-summary div {
	min-width: 0;
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
}

strong {
	color: #ffffff;
	font-size: 14px;
	line-height: 1.35;
}

span {
	color: #aeb7c3;
	font-size: 13px;
	line-height: 1.45;
}

button {
	flex-shrink: 0;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	color: #ffffff;
	font-weight: 700;
	padding: 0.45rem 0.75rem;
	cursor: pointer;
	transition: border-color 0.2s, color 0.2s;
}

button:hover,
button:focus-visible {
	border-color: var(--color-highlight);
	color: var(--color-highlight);
	outline: none;
}

@media (max-width: 520px) {
	.relation-summary {
		align-items: stretch;
		flex-direction: column;
	}
}
</style>
