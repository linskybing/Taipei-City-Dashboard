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
      <button class="toggle-btn">
        {{ open ? "-" : "+" }}
      </button>
    </div>
    <div
      v-show="open"
      class="sticky-body"
    >
      <span>決策助理會依主題、城市與受眾檢索本站組件資料，整理來源、關聯組件與可追溯建議。<br><br>
        無資料或信心不足時，助理會保守回覆並標示原因。</span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.chat-message {
	padding: 4px 10px;
	margin: 0 8px;
	border-radius: 8px;
	background-color: #090909;
}

.sticky-message {
	border: 1px solid #ffffff;
	position: sticky;
	top: 0;
	z-index: 10;
}

.sticky-header {
	display: flex;
	font-weight: bold;
	justify-content: space-between;
	align-items: center;
	cursor: pointer;
	padding: 8px 12px;
}

.sticky-body {
	padding: 8px 12px;
	font-weight: 400;
	font-size: 14px;
}

.toggle-btn {
	background: none;
	border: none;
	font-size: 14px;
	cursor: pointer;
	color: #ffffff;
}
</style>
