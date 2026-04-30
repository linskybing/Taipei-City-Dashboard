<script setup>
defineProps({
	modelValue: { type: String, required: true },
});

const emit = defineEmits(["update:modelValue", "send"]);
</script>

<template>
  <div class="input-area">
    <input
      :value="modelValue"
      type="text"
      autocomplete="off"
      placeholder="輸入城市議題或 component_id..."
      @input="emit('update:modelValue', $event.target.value)"
      @keyup.enter="emit('send', modelValue)"
    >
    <button
      type="button"
      aria-label="送出訊息"
      @click="emit('send', modelValue)"
    >
      <span class="material-icons-round">send</span>
    </button>
  </div>
</template>

<style lang="scss" scoped>
.input-area {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	padding: 0.85rem 1rem 1rem;
	background: #151719;
	border-top: 1px solid #34383d;
}

.input-area input[type="text"] {
	height: 42px;
	width: 100%;
	border: 1px solid #3d4651;
	border-radius: 8px;
	background: #f3f5f7;
	padding: 0 0.9rem;
	outline: none;
	color: #151719;
	font-size: 15px;
	transition: border-color 0.2s, box-shadow 0.2s;
}

.input-area input[type="text"]:focus {
	border-color: var(--color-highlight);
	box-shadow: 0 0 0 3px rgb(90 156 248 / 25%);
}

.input-area button {
	width: 42px;
	height: 42px;
	border: 1px solid #5a9cf8;
	border-radius: 8px;
	background: #1f65c8;
	color: #ffffff;
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	transition: background 0.2s, border-color 0.2s;
}

.input-area button:hover,
.input-area button:focus-visible {
	background: #2b77e0;
	border-color: #8cbcff;
	outline: none;
}
</style>
