<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'

const props = defineProps<{
  logs: string[]
}>()

const logContainer = ref<HTMLDivElement | null>(null)

watch(
  () => props.logs.length,
  async () => {
    await nextTick()
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  },
)
</script>

<template>
  <div class="log-section">
    <div class="log-header">Console</div>
    <div ref="logContainer" class="log-list">
      <div v-for="(line, i) in logs" :key="i" class="log-line">{{ line }}</div>
      <div v-if="logs.length === 0" class="log-empty">No output yet</div>
    </div>
  </div>
</template>

<style scoped>
@keyframes expandIn {
  from { max-height: 0; opacity: 0; margin-top: 0; }
  to { max-height: 160px; opacity: 1; margin-top: 8px; }
}

@keyframes fadeInSlide {
  from { opacity: 0; transform: translateX(-6px); }
  to { opacity: 1; transform: translateX(0); }
}

.log-section {
  margin-top: 8px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.03);
  max-height: 160px;
  display: flex;
  flex-direction: column;
  animation: expandIn 0.4s ease-out;
}

.log-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  padding: 6px 10px;
  border-bottom: 1px solid var(--border);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.log-list {
  flex: 1;
  overflow-y: auto;
  padding: 6px 10px;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.5;
}

.log-line {
  color: var(--text-secondary);
  word-break: break-all;
  animation: fadeInSlide 0.3s ease-out;
}

.log-empty {
  color: var(--text-secondary);
  opacity: 0.4;
  font-style: italic;
}
</style>
