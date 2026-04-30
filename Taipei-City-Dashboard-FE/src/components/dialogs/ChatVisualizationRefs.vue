<script setup>
defineProps({
	items: { type: Array, default: () => [] },
});

const cityLabel = (city) => (city === "taipei" ? "臺北" : "雙北");
const chartTypes = (item) => {
	const types = item?.chart_config?.types;
	return Array.isArray(types) && types.length ? types.join(" / ") : "未提供";
};
const mapCount = (item) => {
	const maps = item?.map_config;
	if (!Array.isArray(maps)) return 0;
	return maps.filter(Boolean).length;
};
const confidenceLabel = (item) =>
	item?.summary?.confidence || item?.uncertainty?.confidence_label || "未標示";
const detailPath = (item) =>
	item?.component_index ? `/component/${item.component_index}` : "";
</script>

<template>
  <section
    v-if="items.length"
    class="chatvisualizationrefs"
  >
    <strong>視覺化引用</strong>
    <article
      v-for="item in items"
      :key="item.trace_id || `${item.component_index}-${item.city}`"
      class="visualization-ref"
    >
      <div class="ref-header">
        <span>{{ cityLabel(item.city) }}</span>
        <span>{{ item.query_type || "未提供格式" }}</span>
        <span>{{ confidenceLabel(item) }}</span>
      </div>
      <h4>{{ item.summary?.headline || item.component_index }}</h4>
      <p v-if="item.summary?.takeaway">
        {{ item.summary.takeaway }}
      </p>
      <dl>
        <div>
          <dt>元件</dt>
          <dd>{{ item.component_index || "未提供" }}</dd>
        </div>
        <div>
          <dt>圖表</dt>
          <dd>{{ chartTypes(item) }}</dd>
        </div>
        <div>
          <dt>地圖圖層</dt>
          <dd>{{ mapCount(item) }}</dd>
        </div>
      </dl>
      <small v-if="item.data_source?.name">
        {{ item.data_source.name }}
      </small>
      <div class="ref-footer">
        <code>{{ item.audit_ref }}</code>
        <a
          v-if="detailPath(item)"
          :href="detailPath(item)"
        >
          查看元件
        </a>
      </div>
    </article>
  </section>
</template>

<style lang="scss" scoped>
.chatvisualizationrefs {
	border: 1px solid #3d4651;
	border-radius: 10px;
	background: #151a20;
	color: var(--color-normal-text);
	padding: 0.75rem 0.85rem;
	display: flex;
	flex-direction: column;
	gap: 0.55rem;
}

.chatvisualizationrefs > strong {
	color: #f4f7fb;
	font-size: 13px;
}

.visualization-ref {
	border-top: 1px solid #2e343b;
	padding-top: 0.65rem;
	display: flex;
	flex-direction: column;
	gap: 0.45rem;
}

.visualization-ref:first-of-type {
	border-top: none;
	padding-top: 0;
}

.ref-header,
.ref-footer {
	display: flex;
	align-items: center;
	gap: 0.45rem;
	flex-wrap: wrap;
}

.ref-header span {
	border: 1px solid #3d4651;
	border-radius: 999px;
	background: #20252b;
	color: #d7e6ff;
	font-size: 11px;
	padding: 0.1rem 0.45rem;
}

h4,
p,
dl {
	margin: 0;
}

h4 {
	color: #ffffff;
	font-size: 14px;
	line-height: 1.35;
}

p,
small {
	color: #b8c1cc;
	font-size: 12px;
	line-height: 1.5;
}

dl {
	display: grid;
	gap: 0.3rem;
}

dl div {
	display: flex;
	justify-content: space-between;
	gap: 0.75rem;
	font-size: 12px;
}

dt {
	color: #8d99a6;
}

dd {
	margin: 0;
	color: #d8e1eb;
	text-align: right;
	word-break: break-word;
}

code {
	color: #aeb9c6;
	font-size: 11px;
	word-break: break-word;
}

a {
	color: #9fc5ff;
	font-size: 12px;
	font-weight: 700;
	text-decoration: none;
}
</style>
