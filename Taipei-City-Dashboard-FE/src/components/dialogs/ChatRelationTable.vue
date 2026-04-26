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
	margin-top: 8px;
	margin-bottom: 8px;
	overflow-x: auto;
}

.relation-table {
	min-width: max-content;
	font-size: 13px;
}

.relation-table th,
.relation-table td {
	border: 1px solid #cccccc;
	text-align: left;
	padding: 0 8px;
	line-height: 1.1;
	vertical-align: middle;
}

.relation-table td {
	height: 2.5rem;
}

.relation-table th {
	font-weight: bold;
	text-align: center;
}
</style>
