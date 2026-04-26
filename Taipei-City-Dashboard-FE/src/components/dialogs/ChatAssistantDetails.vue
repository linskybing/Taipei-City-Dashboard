<script setup>
defineProps({
	actions: { type: Array, default: () => [] },
	sources: { type: Array, default: () => [] },
	confidenceNotes: { type: Array, default: () => [] },
});

const sourceHref = (source) => source?.url || source?.urls?.[0] || "";
const sourceKey = (source, index) =>
	`${source?.name || "source"}-${sourceHref(source)}-${index}`;
</script>

<template>
  <div
    v-if="actions.length"
    class="assistant-list"
  >
    <strong>建議行動</strong>
    <ul>
      <li
        v-for="action in actions"
        :key="action"
      >
        {{ action }}
      </li>
    </ul>
  </div>
  <div
    v-if="sources.length"
    class="assistant-list"
  >
    <strong>資料來源</strong>
    <ul>
      <li
        v-for="(source, index) in sources"
        :key="sourceKey(source, index)"
      >
        <span>{{ source.name || "未命名來源" }}</span>
        <a
          v-if="sourceHref(source)"
          :href="sourceHref(source)"
          target="_blank"
          rel="noopener noreferrer"
        >連結</a>
      </li>
    </ul>
  </div>
  <div
    v-if="confidenceNotes.length"
    class="assistant-list"
  >
    <strong>資料信心</strong>
    <ul>
      <li
        v-for="note in confidenceNotes"
        :key="note"
      >
        {{ note }}
      </li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
.assistant-list {
	border: 1px solid #888787;
	border-radius: 10px;
	background: #282a2c;
	color: #ffffff;
	padding: 8px 12px;
	font-size: 13px;
}

.assistant-list strong {
	display: block;
	margin-bottom: 4px;
}

.assistant-list ul {
	margin: 0;
	padding-left: 18px;
}

.assistant-list a {
	color: #ffffff;
	margin-left: 8px;
	text-decoration: underline;
}
</style>
