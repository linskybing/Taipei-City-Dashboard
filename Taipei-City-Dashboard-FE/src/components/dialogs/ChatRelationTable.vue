<script setup>
defineProps({
	relations: { type: Array, default: () => [] },
});

const cityLabel = (city) => (city === "taipei" ? "臺北" : "雙北");
</script>

<template>
  <div
    v-if="relations.length"
    v-horizontal-wheel
    class="relation-area"
  >
    <table class="relation-table">
      <thead>
        <tr>
          <th>排名</th>
          <th>城市名</th>
          <th>組件名</th>
          <th>關聯性</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(item, index) in relations"
          :key="item.id || index"
        >
          <td>{{ index + 1 }}</td>
          <td>{{ cityLabel(item.city) }}</td>
          <td>{{ item.name }}</td>
          <td>{{ item.score }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style lang="scss" scoped>
.relation-area {
	width: 100%;
	display: flex;
	align-items: center;
	margin: 0.35rem 0;
	border: 1px solid #3d4651;
	border-radius: 8px;
	overflow-x: auto;
	background: #151a20;
}

.relation-table {
	min-width: 100%;
	border-collapse: collapse;
	font-size: 13px;
	color: var(--color-normal-text);
}

.relation-table th,
.relation-table td {
	border-bottom: 1px solid #2e343b;
	text-align: left;
	padding: 0.6rem 0.75rem;
	line-height: 1.35;
	vertical-align: middle;
}

.relation-table tr:last-child td {
	border-bottom: none;
}

.relation-table th {
	font-weight: bold;
	color: #aeb7c3;
	background: #20252b;
	white-space: nowrap;
}

.relation-table td:first-child,
.relation-table td:last-child {
	color: #9fc5ff;
	font-weight: 700;
}
</style>
