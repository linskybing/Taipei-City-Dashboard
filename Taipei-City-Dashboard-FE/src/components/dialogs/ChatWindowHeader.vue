<script setup>
defineProps({
	standalone: { type: Boolean, default: false },
});
</script>

<template>
  <div class="chatwindowheader">
    <div class="chatwindowheader-title">
      <span class="material-icons-round">analytics</span>
      <div>
        <h3>臺北城市儀表板決策助理</h3>
        <p>Taipei City Dashboard</p>
      </div>
    </div>
    <div class="chatwindowheader-tools">
      <a
        v-if="!standalone"
        class="chatwindowheader-action"
        href="/chat"
        target="_blank"
        rel="noopener noreferrer"
        title="在新分頁開啟聊天室"
        aria-label="在新分頁開啟聊天室"
      >
        <span class="material-icons-round">open_in_new</span>
      </a>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.chatwindowheader {
	padding: 0.9rem 1rem;
	background: #151719;
	border-bottom: 1px solid #34383d;
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
}

.chatwindowheader-title {
	min-width: 0;
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.chatwindowheader-title > .material-icons-round {
	width: 36px;
	height: 36px;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	color: var(--color-highlight);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 20px;
}

.chatwindowheader h3,
.chatwindowheader p {
	min-width: 0;
	margin: 0;
}

.chatwindowheader h3 {
	color: var(--color-normal-text);
	font-size: 17px;
	font-weight: 700;
	line-height: 1.2;
}

.chatwindowheader p {
	color: #a8b0ba;
	font-size: 12px;
	line-height: 1.4;
}

.chatwindowheader-tools {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-shrink: 0;
}

.chatwindowheader-action {
	width: 34px;
	height: 34px;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #20252b;
	color: var(--color-normal-text);
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	flex-shrink: 0;
	text-decoration: none;
}

.chatwindowheader-action:hover,
.chatwindowheader-action:focus-visible {
	border-color: var(--color-highlight);
	color: var(--color-highlight);
	outline: none;
}

.chatwindowheader-action .material-icons-round {
	font-size: 20px;
}
</style>
