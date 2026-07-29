<script setup lang="ts">
import type { TunnelInfo } from '../types'

const props = defineProps<{
  tunnels: TunnelInfo[]
}>()

const emit = defineEmits<{
  (e: 'open-browser', id: string): void
  (e: 'disconnect', id: string): void
  (e: 'disconnect-all'): void
  (e: 'make-more'): void
  (e: 'minimize'): void
}>()
</script>

<template>
  <div class="tunnels-view">
    <TransitionGroup name="card">
      <div
        v-for="(tunnel, index) in tunnels"
        :key="tunnel.id"
        class="tunnel-card"
        :style="{ animationDelay: `${index * 0.08}s` }"
      >
        <div class="status-row">
          <span class="status-dot"></span>
          <span class="tunnel-title">{{ tunnel.id }}</span>
        </div>

        <div class="info-text">
          {{ tunnel.sshUser }}@{{ tunnel.sshHost }}:{{ tunnel.sshPort }}
          <br />
          {{ tunnel.localHost }}:{{ tunnel.localPort }} ↔ remote 127.0.0.1:{{ tunnel.remotePort }}
        </div>

        <div class="btn-row">
          <button class="btn btn-primary" @click="emit('open-browser', tunnel.id)">
            Open Browser
          </button>
          <button class="btn btn-danger-outline" @click="emit('disconnect', tunnel.id)">
            Disconnect
          </button>
        </div>
      </div>
    </TransitionGroup>

    <div class="global-actions">
      <button class="btn btn-primary" @click="emit('make-more')">
        + Make More Tunnels
      </button>
      <button class="btn btn-secondary" @click="emit('minimize')">
        Minimize to Tray
      </button>
      <button class="btn btn-danger" @click="emit('disconnect-all')">
        Disconnect All
      </button>
    </div>
  </div>
</template>

<style scoped>
@keyframes slideInRight {
  from { opacity: 0; transform: translateX(30px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes pulseDot {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.4); opacity: 0.7; }
}

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.tunnels-view {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tunnel-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 16px;
  animation: slideInRight 0.4s ease-out;
  animation-fill-mode: both;
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.card-enter-active {
  animation: slideInRight 0.4s ease-out;
  animation-fill-mode: both;
}

.card-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.card-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

.card-move {
  transition: transform 0.35s ease;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--success);
  display: inline-block;
  animation: pulseDot 2s ease-in-out infinite;
}

.tunnel-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
}

.info-text {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
  margin-bottom: 12px;
}

.btn-row {
  display: flex;
  gap: 8px;
  animation: fadeSlideUp 0.3s ease-out 0.15s both;
}

.global-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 4px;
  animation: fadeSlideUp 0.4s ease-out 0.25s both;
}

.btn {
  padding: 8px 20px;
  border-radius: 6px;
  border: none;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s, background 0.25s, box-shadow 0.25s, border-color 0.25s;
}

.btn:hover {
  transform: scale(1.03);
}

.btn:active {
  transform: scale(0.96);
}

.btn-primary {
  background: var(--accent);
  color: white;
}

.btn-primary:hover {
  background: var(--accent-hover);
  box-shadow: 0 4px 14px rgba(67,97,238,0.35);
}

.btn-secondary {
  background: var(--text-secondary);
  color: white;
}

.btn-secondary:hover {
  background: var(--text-primary);
  box-shadow: 0 4px 14px rgba(160,160,176,0.35);
}

.btn-danger {
  background: var(--danger);
  color: white;
}

.btn-danger:hover {
  background: var(--danger-hover);
  box-shadow: 0 4px 14px rgba(231,76,60,0.35);
}

.btn-danger-outline {
  background: transparent;
  color: var(--danger);
  border: 1px solid var(--danger);
}

.btn-danger-outline:hover {
  background: var(--danger);
  color: white;
  box-shadow: 0 4px 14px rgba(231,76,60,0.35);
}
</style>
