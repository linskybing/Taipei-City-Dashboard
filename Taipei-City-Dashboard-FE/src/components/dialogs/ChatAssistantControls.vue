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
      aria-label="選擇主題"
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
      aria-label="選擇城市範圍"
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
      aria-label="選擇受眾"
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
	padding: 0.8rem 1rem 0;
	background: #151719;

	select {
		min-width: 0;
		height: 36px;
		border-radius: 8px;
		border: 1px solid #3d4651;
		background: #20252b;
		color: var(--color-normal-text);
		padding: 0 0.7rem;
		font-weight: 700;
		cursor: pointer;
		transition: border-color 0.2s, background 0.2s;
	}

	select:hover,
	select:focus-visible {
		border-color: var(--color-highlight);
		background: #252b33;
		outline: none;
	}
}
</style>
