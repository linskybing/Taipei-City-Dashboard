<script setup>
import SendIcon from "../icons/SendIcon.vue";

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
      placeholder="輸入訊息..."
      @input="emit('update:modelValue', $event.target.value)"
      @keyup.enter="emit('send', modelValue)"
    >
    <button @click="emit('send', modelValue)">
      <SendIcon />
    </button>
  </div>
</template>

<style lang="scss" scoped>
.input-area {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	padding: 1.5rem 1.125rem;
	background: #494b4e;
}

.input-area input[type="text"] {
	background: #ffffff;
	height: 35px;
	width: 100%;
	border-radius: 20px;
	padding: 0 1rem;
	border: none;
	outline: none;
	color: black;
}

.input-area button {
	height: 35px;
	display: flex;
	align-items: center;
	justify-content: center;
	background: transparent;
	border: none;
	cursor: pointer;
}

.input-area button:hover {
	filter: brightness(0.5);
}
</style>
