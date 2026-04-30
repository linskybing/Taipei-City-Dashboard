<script setup>
defineProps({
	actions: { type: Array, default: () => [] },
	analysisCards: { type: Array, default: () => [] },
	sources: { type: Array, default: () => [] },
	confidenceNotes: { type: Array, default: () => [] },
});

const sourceHref = (source) => source?.url || source?.urls?.[0] || "";
const sourceKey = (source, index) =>
	`${source?.name || "source"}-${sourceHref(source)}-${index}`;
const toolLabel = (tool) => {
	const labels = {
		clean_impute: "資料品質",
		descriptive_report: "描述統計",
		trend_detect: "趨勢偵測",
		seasonal_decompose: "季節分解",
		anomaly_detect: "異常偵測",
		hypothesis_test: "假設檢定",
		forecast_short_mid: "短中期預測",
	};
	return labels[tool] || tool || "統計分析";
};
</script>

<template>
  <div
    v-if="analysisCards.length"
    class="assistant-list analysis-list"
  >
    <strong>統計分析</strong>
    <article
      v-for="card in analysisCards"
      :key="`${card.tool}-${card.headline}`"
      class="analysis-card"
    >
      <div class="analysis-card-header">
        <span>{{ toolLabel(card.tool) }}</span>
        <small>{{ card.confidence_label || card.uncertainty?.confidence_label }}</small>
      </div>
      <p>{{ card.headline }}</p>
      <ul v-if="card.key_findings?.length">
        <li
          v-for="finding in card.key_findings.slice(0, 3)"
          :key="finding"
        >
          {{ finding }}
        </li>
      </ul>
      <small v-if="card.assumptions?.length">
        假設：{{ card.assumptions.slice(0, 2).join("；") }}
      </small>
    </article>
  </div>
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
	border: 1px solid #3d4651;
	border-radius: 10px;
	background: #151a20;
	color: var(--color-normal-text);
	padding: 0.75rem 0.85rem;
	font-size: 13px;
	line-height: 1.5;
}

.assistant-list strong {
	display: block;
	margin-bottom: 0.45rem;
	color: #f4f7fb;
	font-size: 13px;
	letter-spacing: 0;
}

.assistant-list ul {
	margin: 0;
	padding-left: 1rem;
}

.assistant-list a {
	color: #9fc5ff;
	margin-left: 8px;
	text-decoration: underline;
}

.analysis-list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.analysis-card {
	border-top: 1px solid #2e343b;
	padding-top: 0.65rem;
}

.analysis-card:first-of-type {
	border-top: none;
	padding-top: 0;
}

.analysis-card-header {
	display: flex;
	justify-content: space-between;
	gap: 0.75rem;
	font-weight: 700;
}

.analysis-card-header span {
	color: #dbe7ff;
}

.analysis-card-header small {
	border: 1px solid #3d4651;
	border-radius: 999px;
	background: #20252b;
	color: #d7e6ff;
	padding: 0.1rem 0.45rem;
	font-size: 11px;
}

.analysis-card p {
	margin: 0.35rem 0;
	color: #f4f7fb;
	font-weight: 700;
}

.analysis-card small {
	color: #b8c1cc;
}
</style>
