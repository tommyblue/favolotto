import { onMounted, onUnmounted, ref } from "vue";

export function usePolling(fetchFunction, interval = 5000) {
  const data = ref(null);
  let timer = null;

  const fetchData = async () => {
    try {
      data.value = await fetchFunction();
    } catch (error) {
      console.error("Errore nel polling:", error);
    }
  };

  onMounted(() => {
    fetchData();
    timer = setInterval(fetchData, interval);
  });

  onUnmounted(() => {
    clearInterval(timer);
  });

  return { data };
}
