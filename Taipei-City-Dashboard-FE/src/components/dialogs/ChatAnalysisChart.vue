<script setup>
import { computed } from "vue";
import VueApexCharts from "vue3-apexcharts";
import { buildAnalysisChart } from "./chatAnalysisChart";

const props = defineProps({
	card: { type: Object, required: true },
});

const display = computed(() => buildAnalysisChart(props.card));
</script>

<template>
  <div
    v-if="display"
    class="chatanalysischart analysis-chart"
  >
    <div
      v-if="display.mode === 'summary'"
      class="summary-grid"
    >
      <div
        v-for="item in display.metrics"
        :key="item.label"
      >
        <span>{{ item.label }}</span>
        <strong>{{ item.value ?? '-' }}</strong>
      </div>
    </div>
    <table
      v-else-if="display.mode === 'table'"
      class="analysis-table"
    >
      <thead>
        <tr>
          <th
            v-for="column in display.columns"
            :key="column"
          >
            {{ column }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, index) in display.rows"
          :key="`${row.x || 'row'}-${index}`"
        >
          <td
            v-for="column in display.columns"
            :key="column"
          >
            {{ row[column] ?? '-' }}
          </td>
        </tr>
      </tbody>
    </table>
    <VueApexCharts
      v-else
      width="100%"
      :height="display.height"
      :type="display.type"
      :options="display.options"
      :series="display.series"
    />
    <p v-if="display.caption">
      {{ display.caption }}
    </p>
  </div>
</template>

<style lang="scss" scoped>
.chatanalysischart {
	width: 100%;
	min-height: 150px;
	margin: 0.55rem 0 0.35rem;
	border: 1px solid #2e343b;
	border-radius: 8px;
	background: #11161c;
	overflow: hidden;
}

.chatanalysischart :deep(.apexcharts-canvas) {
	max-width: 100%;
}

.summary-grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(86px, 1fr));
	gap: 0.5rem;
	padding: 0.65rem;
}

.summary-grid div {
	border: 1px solid #2e343b;
	border-radius: 6px;
	background: #151a20;
	padding: 0.55rem;
}

.summary-grid span,
.chatanalysischart p {
	color: #aeb9c6;
	font-size: 12px;
}

.summary-grid strong {
	display: block;
	margin-top: 0.15rem;
	color: #f4f7fb;
	font-size: 18px;
}

.analysis-table {
	width: 100%;
	border-collapse: collapse;
	font-size: 12px;
}

.analysis-table th,
.analysis-table td {
	border-bottom: 1px solid #2e343b;
	padding: 0.45rem;
	text-align: left;
}

.analysis-table th {
	color: #dbe7ff;
	font-weight: 700;
}

.analysis-table td {
	color: #f4f7fb;
}

.chatanalysischart p {
	margin: 0;
	padding: 0 0.65rem 0.65rem;
	line-height: 1.5;
}
</style>
