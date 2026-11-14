<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import Form from './components/Form.vue'
import TunnelsView from './components/TunnelsView.vue'
import HostsView from './components/HostsView.vue'
import LogView from './components/LogView.vue'
import { GetConfig, IsConnected, SaveConfig, Connect, Disconnect, DisconnectAll, OpenBrowser, GetTunnels, GetHosts, SaveHost, DeleteHost, MinimizeToTray, ExitApp, HideToTray } from '../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'
import type { Config, TunnelInfo, SavedHost } from './types'

const config = ref<Config>({
  LocalPort: 4096,
  RemotePort: 4096,
  SSHHost: '',
  SSHUser: '',
  SSHPort: 22,
  LocalHost: '127.0.0.1',
  RemoteCmd: '',
  DisconnectCmd: '',
})

const tunnels = ref<TunnelInfo[]>([])
const hosts = ref<SavedHost[]>([])
const connecting = ref(false)
const logs = ref<string[]>([])
const showForm = ref(false)
const viewMode = ref<'tunnels' | 'hosts'>('tunnels')
const editingName = ref('')
const errorMsg = ref('')
const isLight = ref(false)
const showCloseDialog = ref(false)
const logExpanded = ref(false)

let mq: MediaQueryList | null = null

function onThemeChange(e: MediaQueryListEvent | MediaQueryList) {
  isLight.value = e.matches
}

function toggleLogExpand() {
  logExpanded.value = !logExpanded.value
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && logExpanded.value) {
    logExpanded.value = false
  }
}

onMounted(async () => {
  mq = window.matchMedia('(prefers-color-scheme: light)')
  onThemeChange(mq)
  mq.addEventListener('change', onThemeChange)
  window.addEventListener('keydown', onKeydown)
  const cfg = await GetConfig()
  if (cfg) config.value = cfg

  const tuns = await GetTunnels()
  if (tuns && tuns.length > 0) {
    tunnels.value = tuns
  }

  const saved = await GetHosts()
  if (saved) {
    hosts.value = saved
  }

  const ready = () => {
    EventsOn('log', (msg: string) => {
      logs.value.push(msg)
      if (logs.value.length > 200) logs.value.splice(0, logs.value.length - 200)
    })

    EventsOn('status', (status: string) => {
      if (status === 'connected') {
        connecting.value = false
        showForm.value = false
        viewMode.value = 'tunnels'
      } else if (status === 'disconnected') {
        connecting.value = false
      } else if (status === 'connecting' || status === 'waiting') {
        connecting.value = true
      }
    })

    EventsOn('error', (msg: string) => {
      errorMsg.value = msg
      connecting.value = false
    })

    EventsOn('tunnels', (tuns: TunnelInfo[]) => {
      tunnels.value = tuns
    })

    EventsOn('hosts', (hs: SavedHost[]) => {
      hosts.value = hs
    })

    EventsOn('confirm-close', () => {
      showCloseDialog.value = true
    })
  }

  if (window.runtime) {
    ready()
  } else {
    window.addEventListener('runtime-ready', ready, { once: true })
  }
})

onUnmounted(() => {
  EventsOff('log')
  EventsOff('status')
  EventsOff('error')
  EventsOff('tunnels')
  EventsOff('hosts')
  EventsOff('confirm-close')
  mq?.removeEventListener('change', onThemeChange)
  window.removeEventListener('keydown', onKeydown)
})

async function handleConnect(cfg: Config, password: string) {
  errorMsg.value = ''
  connecting.value = true
  viewMode.value = 'tunnels'
  await SaveConfig(cfg)
  config.value = cfg

  try {
    await Connect(
      cfg.SSHUser,
      cfg.SSHHost,
      cfg.SSHPort,
      cfg.LocalPort,
      cfg.RemotePort,
      password,
      cfg.LocalHost,
      cfg.RemoteCmd,
      cfg.DisconnectCmd,
    )
  } catch (e: any) {
    errorMsg.value = e.message || e.toString()
    connecting.value = false
  }
}

async function handleDisconnect(id: string) {
  await Disconnect(id)
}

async function handleDisconnectAll() {
  tunnels.value = []
  await DisconnectAll()
}

async function handleOpenBrowser(id: string) {
  await OpenBrowser(id)
}

async function handleMinimize() {
  await MinimizeToTray()
}

async function handleCloseApp() {
  showCloseDialog.value = false
  await ExitApp()
}

async function handleKeepInTray() {
  showCloseDialog.value = false
  await HideToTray()
}

function handleMakeMore() {
  viewMode.value = 'tunnels'
  showForm.value = true
  editingName.value = ''
}

async function handleSaveHost(name: string, cfg: Config) {
  errorMsg.value = ''
  if (!name || !name.trim()) {
    errorMsg.value = 'Host name is required'
    return
  }
  try {
    await SaveConfig(cfg)
    config.value = cfg
    editingName.value = name.trim()
    await SaveHost(name.trim(), cfg)
  } catch (e: any) {
    errorMsg.value = e.message || e.toString()
  }
}

function handleConnectHost(host: SavedHost) {
  errorMsg.value = ''
  const cfg: Config = {
    LocalPort: host.LocalPort,
    RemotePort: host.RemotePort,
    SSHHost: host.SSHHost,
    SSHUser: host.SSHUser,
    SSHPort: host.SSHPort,
    LocalHost: host.LocalHost,
    RemoteCmd: host.RemoteCmd,
    DisconnectCmd: host.DisconnectCmd,
  }
  config.value = cfg
  connecting.value = true
  viewMode.value = 'tunnels'

  Connect(
    cfg.SSHUser,
    cfg.SSHHost,
    cfg.SSHPort,
    cfg.LocalPort,
    cfg.RemotePort,
    '',
    cfg.LocalHost,
    cfg.RemoteCmd,
    cfg.DisconnectCmd,
  ).catch((e: any) => {
    const msg = e?.message || e?.toString() || ''
    errorMsg.value = msg
    connecting.value = false
    if (msg.toLowerCase().includes('password')) {
      openFormForHost(host)
    }
  })
}

function handleEditHost(host: SavedHost) {
  openFormForHost(host)
}

function openFormForHost(host: SavedHost) {
  config.value = {
    LocalPort: host.LocalPort,
    RemotePort: host.RemotePort,
    SSHHost: host.SSHHost,
    SSHUser: host.SSHUser,
    SSHPort: host.SSHPort,
    LocalHost: host.LocalHost,
    RemoteCmd: host.RemoteCmd,
    DisconnectCmd: host.DisconnectCmd,
  }
  viewMode.value = 'hosts'
  editingName.value = host.name
  showForm.value = true
}

async function handleDeleteHost(name: string) {
  if (!window.confirm(`Delete host "${name}"?`)) return
  try {
    await DeleteHost(name)
    if (editingName.value === name) editingName.value = ''
  } catch (e: any) {
    errorMsg.value = e.message || e.toString()
  }
}
</script>

<template>
  <div class="app">
    <div class="header">
      <h1>FishTun</h1>
      <div class="header-nav">
        <button
          class="nav-btn"
          :class="{ active: viewMode === 'tunnels' }"
          @click="viewMode = 'tunnels'"
        >
          Tunnels
        </button>
        <button
          class="nav-btn"
          :class="{ active: viewMode === 'hosts' }"
          @click="viewMode = 'hosts'"
        >
          Hosts
        </button>
      </div>
    </div>

    <div v-if="errorMsg" class="error-bar">
      {{ errorMsg }}
      <button class="error-close" @click="errorMsg = ''">&times;</button>
    </div>

    <Transition name="view" mode="out-in">
      <div v-if="showForm" key="add-form" class="main-content">
        <Form :config="config" :connecting="connecting" :initial-name="editingName" @connect="handleConnect" @save="handleSaveHost" />
        <div v-if="connecting" class="connecting-bar">
          <span class="spinner"></span>
          Establishing SSH tunnel...
        </div>
        <div v-if="!connecting" class="btn-row-centered">
          <button class="btn btn-secondary cancel-btn" @click="showForm = false">Cancel</button>
        </div>
      </div>

      <div v-else-if="viewMode === 'hosts' || tunnels.length === 0" key="hosts" class="main-content">
        <HostsView
          :hosts="hosts"
          @connect="handleConnectHost"
          @edit="handleEditHost"
          @delete="handleDeleteHost"
          @new="handleMakeMore"
        />
      </div>

      <div v-else key="tunnels" class="main-content">
        <TunnelsView
          :tunnels="tunnels"
          @open-browser="handleOpenBrowser"
          @disconnect="handleDisconnect"
          @disconnect-all="handleDisconnectAll"
          @make-more="handleMakeMore"
          @minimize="handleMinimize"
        />
      </div>
    </Transition>

    <LogView :logs="logs" :expanded="logExpanded" @toggle-expand="toggleLogExpand" />

    <Transition name="dialog">
      <div v-if="showCloseDialog" class="dialog-overlay" @click.self="handleKeepInTray">
        <div class="dialog-box">
          <h2 class="dialog-title">Close FishTun?</h2>
          <p class="dialog-text">
            Keep FishTun running in the system tray, or quit the application and disconnect all tunnels?
          </p>
          <div class="dialog-actions">
            <button class="btn btn-primary" @click="handleKeepInTray">System tray</button>
            <button class="btn btn-danger" @click="handleCloseApp">Close app</button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style>
:root {
  --bg-primary: #12121e;
  --bg-secondary: #1a1a2e;
  --bg-card: #16213e;
  --text-primary: #e0e0e0;
  --text-secondary: #a0a0b0;
  --btn-secondary-bg: #2a2a4a;
  --btn-secondary-bg-hover: #35355c;
  --btn-secondary-text: #d8d8e6;
  --btn-secondary-text-hover: #ffffff;
  --btn-secondary-border: #3a3a5f;
  --btn-secondary-border-hover: #4361ee;
  --btn-secondary-shadow: rgba(67, 97, 238, 0.30);
  --accent: #4361ee;
  --accent-hover: #3a56d4;
  --success: #50dc78;
  --danger: #e74c3c;
  --danger-hover: #c0392b;
  --border: #2a2a4a;
  --error-bg: #4a1520;
  --error-text: #ff6b6b;
}

@media (prefers-color-scheme: light) {
  :root {
    --bg-primary: #f0f0f2;
    --bg-secondary: #ffffff;
    --bg-card: #ffffff;
    --text-primary: #1d1d1f;
    --text-secondary: #6e6e73;
    --btn-secondary-bg: #e6e6ea;
    --btn-secondary-bg-hover: #dcdce2;
    --btn-secondary-text: #3a3a44;
    --btn-secondary-text-hover: #1d1d1f;
    --btn-secondary-border: #cfcfd6;
    --btn-secondary-border-hover: #4361ee;
    --btn-secondary-shadow: rgba(67, 97, 238, 0.20);
    --accent: #4361ee;
    --accent-hover: #3a56d4;
    --success: #34c759;
    --danger: #ff3b30;
    --danger-hover: #d62d20;
    --border: #d2d2d7;
    --error-bg: #fce8e8;
    --error-text: #c0392b;
  }
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

@keyframes gradientShift {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

@keyframes glowPulse {
  0%, 100% { text-shadow: 0 0 8px rgba(67,97,238,0.3); }
  50% { text-shadow: 0 0 24px rgba(67,97,238,0.7); }
}

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideDown {
  from { max-height: 0; opacity: 0; }
  to { max-height: 60px; opacity: 1; }
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: linear-gradient(-45deg, #12121e, #1a1a2e, #16213e, #0f0f1a);
  background-size: 400% 400%;
  animation: gradientShift 15s ease infinite;
  color: var(--text-primary);
  overflow-y: auto;
}

@media (prefers-color-scheme: light) {
  body {
    background: linear-gradient(-45deg, #f0f0f2, #ffffff, #e8e8ec, #f5f5f7);
    background-size: 400% 400%;
    animation: gradientShift 15s ease infinite;
  }
}

.app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  padding: 16px;
  animation: fadeSlideUp 0.6s ease-out;
}

.header h1 {
  font-size: 20px;
  text-align: center;
  padding: 8px 0 16px;
  color: var(--accent);
  letter-spacing: 1px;
  animation: glowPulse 3s ease-in-out infinite;
}

.header-nav {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-bottom: 12px;
}

.nav-btn {
  position: relative;
  overflow: hidden;
  padding: 6px 18px;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s, color 0.2s, border-color 0.2s, transform 0.15s;
}

.nav-btn::before {
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

.nav-btn:hover {
  color: var(--text-primary);
}

.nav-btn:hover::before {
  transform: translateX(150%) skewX(-20deg);
}

.nav-btn.active {
  background: var(--accent);
  border-color: var(--accent);
  color: white;
}

.main-content {
  flex: 1;
  overflow-y: auto;
}

.error-bar {
  background: var(--error-bg);
  color: var(--error-text);
  padding: 8px 12px;
  border-radius: 6px;
  margin-bottom: 8px;
  font-size: 13px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  overflow: hidden;
  animation: slideDown 0.3s ease-out;
}

.error-close {
  background: none;
  border: none;
  color: var(--error-text);
  font-size: 18px;
  cursor: pointer;
  padding: 0 4px;
}

.spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  vertical-align: middle;
  margin-right: 6px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.connecting-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 12px;
  padding: 10px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  animation: fadeSlideUp 0.4s ease-out;
}

.btn-row-centered {
  display: flex;
  justify-content: center;
  margin-top: 12px;
}

.cancel-btn {
  animation: fadeSlideUp 0.3s ease-out both;
}

.view-enter-active {
  animation: fadeSlideUp 0.35s ease-out;
}

.view-leave-active {
  animation: fadeSlideUp 0.2s ease-in reverse;
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
  transition: background 0.2s, transform 0.15s, box-shadow 0.2s;
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

.btn:active:not(:disabled) {
  transform: scaleY(0.9);
}

.btn-secondary {
  background: var(--btn-secondary-bg);
  color: var(--btn-secondary-text);
  border: 1px solid var(--btn-secondary-border);
}

.btn-secondary:hover {
  background: var(--btn-secondary-bg-hover);
  border-color: var(--btn-secondary-border-hover);
  color: var(--btn-secondary-text-hover);
  box-shadow: 0 4px 14px var(--btn-secondary-shadow);
}

.btn-danger {
  background: var(--danger);
  color: white;
}

.btn-primary {
  background: var(--accent);
  color: white;
}

.btn-primary:hover {
  background: var(--accent-hover);
  box-shadow: 0 4px 14px rgba(67, 97, 238, 0.4);
}

.btn-danger:hover {
  background: var(--danger-hover);
  box-shadow: 0 4px 14px rgba(231,76,60,0.4);
}

.dialog-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(2px);
}

.dialog-box {
  width: min(360px, 88vw);
  padding: 20px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg-secondary);
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.45);
  animation: fadeSlideUp 0.2s ease-out;
}

.dialog-title {
  font-size: 16px;
  margin-bottom: 10px;
  color: var(--accent);
  text-align: center;
}

.dialog-text {
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-secondary);
  text-align: center;
  margin-bottom: 18px;
}

.dialog-actions {
  display: flex;
  justify-content: center;
  gap: 10px;
}

.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 0.18s ease;
}

.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}
</style>
