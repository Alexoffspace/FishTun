<script setup lang="ts">
import { reactive, watch } from 'vue'
import type { Config } from '../types'

const props = defineProps<{
  config: Config
  connecting: boolean
}>()

const emit = defineEmits<{
  (e: 'connect', config: Config, password: string): void
}>()

const form = reactive({
  localPort: props.config.LocalPort.toString(),
  remotePort: props.config.RemotePort.toString(),
  sshHost: props.config.SSHHost,
  sshUser: props.config.SSHUser,
  sshPort: props.config.SSHPort.toString(),
  localHost: props.config.LocalHost || '127.0.0.1',
  password: '',
  remoteCmd: props.config.RemoteCmd,
  disconnectCmd: props.config.DisconnectCmd,
})

watch(() => ({ ...props.config }), (cfg) => {
  form.localPort = cfg.LocalPort.toString()
  form.remotePort = cfg.RemotePort.toString()
  form.sshHost = cfg.SSHHost
  form.sshUser = cfg.SSHUser
  form.sshPort = cfg.SSHPort.toString()
  form.localHost = cfg.LocalHost || '127.0.0.1'
  form.remoteCmd = cfg.RemoteCmd
  form.disconnectCmd = cfg.DisconnectCmd
})

function parsePort(s: string, def: number): number {
  const v = parseInt(s, 10)
  return isNaN(v) || v < 1 || v > 65535 ? def : v
}

function handleSubmit() {
  const cfg: Config = {
    LocalPort: parsePort(form.localPort, 4096),
    RemotePort: parsePort(form.remotePort, 8080),
    SSHHost: form.sshHost,
    SSHUser: form.sshUser,
    SSHPort: parsePort(form.sshPort, 22),
    LocalHost: form.localHost,
    RemoteCmd: form.remoteCmd,
    DisconnectCmd: form.disconnectCmd,
  }
  emit('connect', cfg, form.password)
}
</script>

<template>
  <div class="form-card">
    <div class="section-title">Tunnel Ports</div>
    <div class="field">
      <label>Local</label>
      <input v-model="form.localPort" type="text" placeholder="2222" :disabled="connecting" />
    </div>
    <div class="field">
      <label>Bind</label>
      <select v-model="form.localHost" :disabled="connecting">
        <option value="127.0.0.1">127.0.0.1</option>
        <option value="0.0.0.0">0.0.0.0</option>
      </select>
    </div>
    <div class="field">
      <label>Remote</label>
      <input v-model="form.remotePort" type="text" placeholder="8080" :disabled="connecting" />
    </div>

    <div class="section-title">SSH Connection</div>
    <div class="field">
      <label>Host</label>
      <input v-model="form.sshHost" type="text" placeholder="my.server.com" :disabled="connecting" />
    </div>
    <div class="field">
      <label>User</label>
      <input v-model="form.sshUser" type="text" placeholder="root" :disabled="connecting" />
    </div>
    <div class="field">
      <label>Port</label>
      <input v-model="form.sshPort" type="text" placeholder="22" :disabled="connecting" />
    </div>
    <div class="field">
      <label>Password</label>
      <input v-model="form.password" type="password" placeholder="leave empty for key auth" :disabled="connecting" />
    </div>

    <div class="section-title">Remote Command</div>
    <div class="field">
      <label>Command</label>
      <input v-model="form.remoteCmd" type="text" placeholder="e.g. myapp --port 8080" :disabled="connecting" />
    </div>

    <div class="section-title">Disconnect Command (optional)</div>
    <div class="field">
      <label>Command</label>
      <input v-model="form.disconnectCmd" type="text" placeholder="kill ... / leave empty" :disabled="connecting" />
    </div>

    <div class="btn-row">
      <button class="btn btn-primary" :disabled="connecting" @click="handleSubmit">
        {{ connecting ? 'Connecting...' : 'Connect' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
@keyframes cardIn {
  from { opacity: 0; transform: scale(0.95) translateY(8px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

@keyframes shimmer {
  0% { background-position: -200% center; }
  100% { background-position: 200% center; }
}

.form-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 16px;
  animation: cardIn 0.5s ease-out;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 12px 0 8px;
}

.section-title:first-child {
  margin-top: 0;
}

.field {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.field label {
  width: 72px;
  font-size: 13px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.field input {
  flex: 1;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 8px 10px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  transition: border-color 0.3s, box-shadow 0.3s;
}

.field input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(67,97,238,0.25);
}

.field input:disabled {
  opacity: 0.5;
}

.field select {
  flex: 1;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 8px 10px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  transition: border-color 0.3s, box-shadow 0.3s;
  cursor: pointer;
  appearance: auto;
}

.field select:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(67,97,238,0.25);
}

.field select:disabled {
  opacity: 0.5;
}

.btn-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  gap: 8px;
}

.btn {
  padding: 8px 20px;
  border-radius: 6px;
  border: none;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s, transform 0.2s, box-shadow 0.2s;
}

.btn:active:not(:disabled) {
  transform: scale(0.96);
}

.btn-primary {
  background: var(--accent);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-hover);
  box-shadow: 0 4px 14px rgba(67,97,238,0.35);
  transform: translateY(-1px);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: linear-gradient(90deg, var(--accent), var(--accent-hover), var(--accent));
  background-size: 200% 100%;
  animation: shimmer 1.5s linear infinite;
}
</style>
