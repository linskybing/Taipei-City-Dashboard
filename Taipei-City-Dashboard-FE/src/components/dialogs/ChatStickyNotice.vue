<script setup>
defineProps({
	open: { type: Boolean, required: true },
});

const emit = defineEmits(["toggle"]);
</script>

<template>
  <div class="chat-message sticky-message">
    <div
      class="sticky-header"
      @click="emit('toggle')"
    >
      <span>置頂公告：決策助理使用須知</span>
      <button
        class="toggle-btn"
        type="button"
        :aria-label="open ? '收合公告' : '展開公告'"
      >
        <span class="material-icons-round">{{ open ? "remove" : "add" }}</span>
      </button>
    </div>
    <div
      v-show="open"
      class="sticky-body"
    >
      <span>決策助理會依主題、城市與受眾檢索本站組件資料，整理來源、關聯組件與可追溯建議。<br><br>
        若輸入 component_id 或要求統計分析，助理會使用資料品質、描述統計、趨勢、季節、異常、檢定與預測工具。無資料或信心不足時，助理會保守回覆並標示原因。</span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.chat-message {
	padding: 0;
	margin: 0 0 0.75rem;
	border-radius: 8px;
	background: #101418;
}

.sticky-message {
	border: 1px solid #3d4651;
	position: sticky;
	top: 0;
	z-index: 10;
	box-shadow: 0 10px 24px rgb(0 0 0 / 22%);
}

.sticky-header {
	display: flex;
	font-weight: bold;
	justify-content: space-between;
	align-items: center;
	cursor: pointer;
	padding: 0.7rem 0.85rem;
	color: #f4f7fb;
}

.sticky-body {
	border-top: 1px solid #2e343b;
	padding: 0.75rem 0.85rem;
	font-weight: 400;
	font-size: 14px;
	line-height: 1.6;
	color: #c5ccd6;
}

.toggle-btn {
	width: 28px;
	height: 28px;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	cursor: pointer;
	color: #ffffff;
	display: flex;
	align-items: center;
	justify-content: center;
}

.toggle-btn:hover,
.toggle-btn:focus-visible {
	border-color: var(--color-highlight);
	color: var(--color-highlight);
	outline: none;
}

.toggle-btn .material-icons-round {
	font-size: 18px;
}
</style>
