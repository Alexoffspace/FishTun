<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'

const props = defineProps<{
  logs: string[]
  expanded: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle-expand'): void
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

watch(
  () => props.expanded,
  async () => {
    await nextTick()
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  },
)
</script>

<template>
  <div class="log-section" :class="{ expanded }">
    <div class="log-header">
      <span class="log-title">Console</span>
      <button
        class="log-expand"
        :title="expanded ? 'Restore' : 'Expand to full window'"
        :aria-label="expanded ? 'Restore console' : 'Expand console to full window'"
        :aria-pressed="expanded"
        @click="emit('toggle-expand')"
      >
        <svg v-if="!expanded" viewBox="0 0 16 16" width="13" height="13" aria-hidden="true">
          <path
            d="M6 2H2v4M10 2h4v4M6 14H2v-4M10 14h4v-4"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        <svg v-else viewBox="0 0 16 16" width="13" height="13" aria-hidden="true">
          <path
            d="M2 6h4V2M14 6h-4V2M2 10h4v4M14 10h-4v4"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>
    </div>
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

.log-section.expanded {
  position: fixed;
  inset: 0;
  z-index: 900;
  margin: 0;
  max-height: none;
  border: none;
  border-radius: 0;
  background: var(--bg-primary);
  animation: none;
}

.log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  padding: 4px 6px 4px 10px;
  border-bottom: 1px solid var(--border);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.log-section.expanded .log-header {
  padding: 8px 12px;
}

.log-title {
  flex: 1;
}

.log-expand {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border: 1px solid transparent;
  border-radius: 5px;
  background: none;
  color: var(--text-secondary);
  cursor: pointer;
  transition: color 0.2s, background 0.2s, border-color 0.2s;
}

.log-expand:hover {
  color: var(--text-primary);
  background: var(--btn-secondary-bg);
  border-color: var(--btn-secondary-border);
}

.log-list {
  flex: 1;
  overflow-y: auto;
  padding: 6px 10px;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.5;
}

.log-section.expanded .log-list {
  padding: 12px 16px;
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
