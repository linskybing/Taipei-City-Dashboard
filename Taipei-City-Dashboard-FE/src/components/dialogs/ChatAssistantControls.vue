<script setup>
defineProps({
	theme: { type: String, required: true },
	city: { type: String, required: true },
	audience: { type: String, required: true },
});

const emit = defineEmits(["update:theme", "update:city", "update:audience"]);

const themeOptions = [
	{ value: "auto", label: "自動" },
	{ value: "commuting", label: "智慧通勤" },
	{ value: "disaster", label: "韌性防災" },
	{ value: "environment", label: "永續環境" },
	{ value: "health", label: "食安健康" },
	{ value: "labor", label: "勞動福祉" },
	{ value: "culture", label: "文化共融" },
];
const cityOptions = [
	{ value: "metrotaipei", label: "雙北" },
	{ value: "taipei", label: "臺北" },
];
const audienceOptions = [
	{ value: "government", label: "政府" },
	{ value: "public", label: "民眾" },
];
</script>

<template>
  <div class="assistant-controls">
    <select
      :value="theme"
      @change="emit('update:theme', $event.target.value)"
    >
      <option
        v-for="option in themeOptions"
        :key="option.value"
        :value="option.value"
      >
        {{ option.label }}
      </option>
    </select>
    <select
      :value="city"
      @change="emit('update:city', $event.target.value)"
    >
      <option
        v-for="option in cityOptions"
        :key="option.value"
        :value="option.value"
      >
        {{ option.label }}
      </option>
    </select>
    <select
      :value="audience"
      @change="emit('update:audience', $event.target.value)"
    >
      <option
        v-for="option in audienceOptions"
        :key="option.value"
        :value="option.value"
      >
        {{ option.label }}
      </option>
    </select>
  </div>
</template>

<style lang="scss" scoped>
.assistant-controls {
	display: grid;
	grid-template-columns: 1fr 1fr 1fr;
	gap: 0.5rem;
	padding: 0.75rem 1.125rem 0;
	background: #494b4e;

	select {
		min-width: 0;
		height: 32px;
		border-radius: 8px;
		border: 1px solid #888787;
		background: #282a2c;
		color: #ffffff;
		padding: 0 8px;
	}
}
</style>
