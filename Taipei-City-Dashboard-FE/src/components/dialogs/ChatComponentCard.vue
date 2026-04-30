<script setup>
import { computed } from "vue";

const props = defineProps({
	item: { type: Object, required: true },
});

const component = computed(() => props.item.component || {});
const cityLabel = computed(() =>
	component.value.city === "taipei" ? "臺北" : "雙北",
);
const scoreLabel = computed(() => {
	const score = Number(component.value.score);
	if (!Number.isFinite(score) || score <= 0) return "未標示";
	return score <= 1 ? `${Math.round(score * 100)}%` : score.toFixed(2);
});
const summary = computed(
	() => component.value.short_desc || component.value.long_desc || "此組件尚無摘要，可先依名稱與資料欄位判讀。",
);
const source = computed(() => component.value.source || "資料來源未補齊");
const detailPath = computed(() =>
	component.value.index ? `/component/${component.value.index}` : "",
);
</script>

<template>
  <article class="component-card">
    <div class="card-topline">
      <span>#{{ item.rank }}</span>
      <span>{{ cityLabel }}</span>
      <span>關聯度 {{ scoreLabel }}</span>
    </div>
    <h3>{{ component.name || "未命名組件" }}</h3>
    <dl>
      <div>
        <dt>component_id</dt>
        <dd>{{ component.id || "未提供" }}</dd>
      </div>
      <div>
        <dt>index</dt>
        <dd>{{ component.index || "未提供" }}</dd>
      </div>
      <div>
        <dt>query_type</dt>
        <dd>{{ component.query_type || "未提供" }}</dd>
      </div>
    </dl>
    <p>{{ summary }}</p>
    <small>{{ source }}</small>
    <div class="card-actions">
      <span v-if="item.failed">使用 AI 關聯資料顯示</span>
      <a
        v-if="detailPath"
        :href="detailPath"
      >
        查看詳情
      </a>
    </div>
  </article>
</template>

<style lang="scss" scoped>
.component-card {
	border: 1px solid #303841;
	border-radius: 8px;
	background: #171c21;
	padding: 0.85rem;
	display: flex;
	flex-direction: column;
	gap: 0.55rem;
	flex: 0 0 auto;
	min-width: 0;
	overflow: visible;
}

.component-card * {
	min-width: 0;
	overflow: visible;
}

.card-topline,
.card-actions {
	display: flex;
	align-items: center;
	gap: 0.45rem;
	flex-wrap: wrap;
}

.card-topline span {
	border: 1px solid #37424d;
	border-radius: 999px;
	color: #b9c5d2;
	font-size: 12px;
	padding: 0.15rem 0.45rem;
}

h3,
p,
dl {
	margin: 0;
}

h3 {
	color: #ffffff;
	font-size: 15px;
	line-height: 1.35;
	overflow-wrap: anywhere;
}

dl {
	display: grid;
	gap: 0.35rem;
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
	overflow-wrap: anywhere;
}

p {
	color: #c8d2dc;
	font-size: 13px;
	line-height: 1.55;
	overflow-wrap: anywhere;
}

small,
.card-actions span {
	color: #8d99a6;
	font-size: 12px;
	line-height: 1.4;
	overflow-wrap: anywhere;
}

.card-actions {
	justify-content: space-between;
}

a {
	color: #8fc2ff;
	font-size: 13px;
	font-weight: 700;
	text-decoration: none;
}
</style>
