import { computed, ref, watch } from "vue";
import http from "../../router/axios";

const MAX_CANVAS_ITEMS = 8;

export function useChatComponentCanvas(chatData) {
	const isCanvasOpen = ref(false);
	const canvasRelations = ref([]);
	const componentDetails = ref({});
	const componentErrors = ref({});
	let requestSeq = 0;

	const canvasItems = computed(() =>
		canvasRelations.value.map((relation, index) => {
			const key = relationKey(relation);
			const detail = componentDetails.value[key] || {};
			return {
				key,
				rank: index + 1,
				relation,
				failed: Boolean(componentErrors.value[key]),
				component: { ...relation, ...detail, score: relation.score },
			};
		}),
	);

	const openCanvas = (relations = []) => {
		const normalized = normalizeRelations(relations);
		if (!normalized.length) return;
		canvasRelations.value = normalized;
		isCanvasOpen.value = true;
		enrichComponents(normalized);
	};

	const closeCanvas = () => {
		isCanvasOpen.value = false;
	};

	watch(
		() => latestRelationSignature(chatData.value),
		() => {
			const latest = [...chatData.value]
				.reverse()
				.find((item) => item.role === "bot" && item.relations?.length);
			if (latest) openCanvas(latest.relations);
		},
		{ immediate: true },
	);

	async function enrichComponents(relations) {
		const seq = ++requestSeq;
		const keys = new Set(relations.map(relationKey));
		componentDetails.value = pickKnownKeys(componentDetails.value, keys);
		componentErrors.value = pickKnownKeys(componentErrors.value, keys);

		await Promise.all(
			relations.map(async (relation) => {
				if (!relation.id) return;
				const key = relationKey(relation);
				try {
					const response = await http.get(`/component/${relation.id}`, {
						params: { city: relation.city || "taipei" },
						skipErrorNotify: true,
						skipGlobalLoading: true,
					});
					if (seq !== requestSeq) return;
					componentDetails.value = {
						...componentDetails.value,
						[key]: response.data?.data || {},
					};
				} catch {
					if (seq !== requestSeq) return;
					componentErrors.value = { ...componentErrors.value, [key]: true };
				}
			}),
		);
	}

	return { isCanvasOpen, canvasRelations, canvasItems, openCanvas, closeCanvas };
}

function normalizeRelations(relations) {
	const seen = new Set();
	return relations
		.filter(Boolean)
		.map((relation) => ({ ...relation, city: relation.city || "taipei" }))
		.filter((relation) => {
			const key = relationKey(relation);
			if ((!relation.id && !relation.index) || seen.has(key)) return false;
			seen.add(key);
			return true;
		})
		.slice(0, MAX_CANVAS_ITEMS);
}

function latestRelationSignature(items = []) {
	const latest = [...items]
		.reverse()
		.find((item) => item.role === "bot" && item.relations?.length);
	if (!latest) return "";
	return `${latest.id}:${latest.relations.map(relationKey).join("|")}`;
}

function relationKey(relation = {}) {
	return `${relation.id || relation.index || "component"}-${relation.city || "taipei"}`;
}

function pickKnownKeys(source, keys) {
	return Object.fromEntries(Object.entries(source).filter(([key]) => keys.has(key)));
}
