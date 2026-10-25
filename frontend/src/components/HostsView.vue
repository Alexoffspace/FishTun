<script setup lang="ts">
import type { SavedHost } from '../types'

const props = defineProps<{
  hosts: SavedHost[]
}>()

const emit = defineEmits<{
  (e: 'connect', host: SavedHost): void
  (e: 'edit', host: SavedHost): void
  (e: 'delete', name: string): void
  (e: 'new'): void
}>()
</script>

<template>
  <div class="hosts-view">
    <div v-if="hosts.length === 0" class="empty-hint">
      No saved hosts yet. Fill the form and press Save.
    </div>

    <TransitionGroup name="card">
      <div
        v-for="(host, index) in hosts"
        :key="host.name"
        class="host-card"
        :style="{ animationDelay: `${index * 0.06}s` }"
      >
        <div class="host-title">{{ host.name }}</div>

        <div class="info-text">
          {{ host.SSHUser }}@{{ host.SSHHost }}:{{ host.SSHPort }}
          <br />
          {{ host.LocalHost }}:{{ host.LocalPort }} ↔ remote 127.0.0.1:{{ host.RemotePort }}
        </div>

        <div class="btn-row">
          <button class="btn btn-primary" @click="emit('connect', host)">Connect</button>
          <button class="btn btn-secondary" @click="emit('edit', host)">Edit</button>
          <button class="btn btn-danger-outline" @click="emit('delete', host.name)">Delete</button>
        </div>
      </div>
    </TransitionGroup>

    <div class="global-actions">
      <button class="btn btn-primary" @click="emit('new')">+ New Host</button>
    </div>
  </div>
</template>

<style scoped>
@keyframes slideInRight {
  from { opacity: 0; transform: translateX(30px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

.hosts-view {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.empty-hint {
  font-size: 13px;
  color: var(--text-secondary);
  text-align: center;
  padding: 24px 12px;
  animation: fadeSlideUp 0.4s ease-out;
}

.host-card {
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

.host-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
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
  position: relative;
  overflow: hidden;
  padding: 8px 20px;
  border-radius: 6px;
  border: none;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transform-origin: center;
  transition: transform 0.15s, background 0.25s, box-shadow 0.25s, border-color 0.25s;
}

.btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 50%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.35), transparent);
  transform: translateX(-150%) skewX(-20deg);
  transition: transform 0.6s ease;
  pointer-events: none;
}

.btn:hover::before {
  transform: translateX(150%) skewX(-20deg);
}

.btn:active {
  transform: scaleY(0.9);
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
