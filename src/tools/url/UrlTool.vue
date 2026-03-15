<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <button class="toolbar-btn" :class="{ active: direction === 'encode' }" @click="direction = 'encode'">Encode</button>
      <button class="toolbar-btn" :class="{ active: direction === 'decode' }" @click="direction = 'decode'">Decode</button>
      <select v-model="encodeMode" v-if="direction === 'encode'">
        <option value="component">encodeURIComponent</option>
        <option value="uri">encodeURI</option>
      </select>
      <div class="spacer"></div>
      <ShareButton tool="url" :getState="() => ({ input: input, direction, encodeMode })" />
      <button class="toolbar-btn" @click="copy(output, 'Copied')">Copy output</button>
      <button class="toolbar-btn" @click="input = ''">Clear</button>
    </div>
    <div class="url-layout">
      <div class="tool-content">
        <SplitPane :initial-split="50">
          <template #left>
            <div class="pane editor-pane">
              <PaneHeader :label="direction === 'encode' ? 'Plain Text' : 'Encoded URL'">{{ input.length.toLocaleString() }} chars</PaneHeader>
              <textarea
                v-model="input"
                :placeholder="direction === 'encode' ? 'Enter text to URL-encode...' : 'Paste URL-encoded text...'"
                spellcheck="false"
              ></textarea>
            </div>
          </template>
          <template #right>
            <div class="pane preview-pane">
              <PaneHeader :label="direction === 'encode' ? 'Encoded Output' : 'Decoded Text'">{{ output.length.toLocaleString() }} chars</PaneHeader>
              <div class="url-output">
                <div v-if="error" class="url-error">{{ error }}</div>
                <pre v-else class="url-pre">{{ output }}</pre>
              </div>
            </div>
          </template>
        </SplitPane>
      </div>
      <div class="url-parser" v-if="urlParts">
        <PaneHeader label="URL Breakdown" />
        <div class="url-parts">
          <div class="kv-row" v-if="urlParts.protocol">
            <span class="kv-label">Protocol</span>
            <span class="kv-value">{{ urlParts.protocol }}</span>
          </div>
          <div class="kv-row" v-if="urlParts.host">
            <span class="kv-label">Host</span>
            <span class="kv-value">{{ urlParts.host }}</span>
          </div>
          <div class="kv-row" v-if="urlParts.port">
            <span class="kv-label">Port</span>
            <span class="kv-value">{{ urlParts.port }}</span>
          </div>
          <div class="kv-row" v-if="urlParts.pathname && urlParts.pathname !== '/'">
            <span class="kv-label">Path</span>
            <span class="kv-value">{{ urlParts.pathname }}</span>
          </div>
          <div class="kv-row" v-if="urlParts.hash">
            <span class="kv-label">Fragment</span>
            <span class="kv-value">{{ urlParts.hash }}</span>
          </div>
          <template v-if="params.length">
            <div class="params-header">Query Parameters</div>
            <div class="kv-row param-row" v-for="(p, i) in params" :key="i">
              <input
                class="param-key"
                :value="p.key"
                @input="updateParam(i, 'key', $event.target.value)"
                placeholder="key"
              />
              <span class="param-eq">=</span>
              <input
                class="param-value"
                :value="p.value"
                @input="updateParam(i, 'value', $event.target.value)"
                placeholder="value"
              />
              <button class="param-remove" @click="removeParam(i)" title="Remove parameter">&times;</button>
            </div>
            <button class="toolbar-btn add-param-btn" @click="addParam">+ Add parameter</button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import SplitPane from '../../components/SplitPane.vue'
import PaneHeader from '../../components/PaneHeader.vue'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { copy } = useClipboard()

const input = ref('')
const output = ref('')
const direction = ref('encode')
const encodeMode = ref('component')
const error = ref('')

const { useShareable } = useShare()
useShareable('url', (s) => {
  input.value = s.input || ''
  if (s.direction) direction.value = s.direction
  if (s.encodeMode) encodeMode.value = s.encodeMode
})

watch([input, direction, encodeMode], () => {
  if (!input.value) {
    output.value = ''
    error.value = ''
    return
  }
  try {
    if (direction.value === 'encode') {
      output.value = encodeMode.value === 'component'
        ? encodeURIComponent(input.value)
        : encodeURI(input.value)
    } else {
      output.value = decodeURIComponent(input.value)
    }
    error.value = ''
  } catch (e) {
    error.value = e.message
    output.value = ''
  }
})

// URL parser
const urlParts = computed(() => {
  const raw = direction.value === 'encode' ? input.value : output.value
  if (!raw || !raw.trim()) return null
  try {
    const url = new URL(raw.trim())
    return {
      protocol: url.protocol,
      host: url.host,
      pathname: url.pathname,
      port: url.port,
      hash: url.hash,
    }
  } catch {
    return null
  }
})

const params = ref([])
let updatingFromParams = false

// Sync params from parsed URL
watch(() => urlParts.value, (parts) => {
  if (updatingFromParams) return
  if (!parts) { params.value = []; return }
  try {
    const raw = direction.value === 'encode' ? input.value : output.value
    const url = new URL(raw.trim())
    const entries = []
    url.searchParams.forEach((value, key) => {
      entries.push({ key, value })
    })
    params.value = entries
  } catch {
    params.value = []
  }
}, { immediate: true })

function rebuildUrlFromParams() {
  const raw = direction.value === 'encode' ? input.value : output.value
  if (!raw || !raw.trim()) return
  try {
    const url = new URL(raw.trim())
    // Clear existing params and rebuild
    const keys = [...url.searchParams.keys()]
    keys.forEach(k => url.searchParams.delete(k))
    params.value.forEach(p => {
      if (p.key) url.searchParams.append(p.key, p.value)
    })
    updatingFromParams = true
    if (direction.value === 'encode') {
      input.value = url.toString()
    } else {
      input.value = url.toString()
    }
    updatingFromParams = false
  } catch {
    // If URL parsing fails, do nothing
  }
}

function updateParam(index, field, value) {
  params.value[index][field] = value
  rebuildUrlFromParams()
}

function removeParam(index) {
  params.value.splice(index, 1)
  rebuildUrlFromParams()
}

function addParam() {
  params.value.push({ key: '', value: '' })
}
</script>

<style scoped>
.url-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.url-layout .tool-content {
  flex: 1;
}

.url-output {
  flex: 1;
  overflow: auto;
  padding: 20px 24px;
}

.url-pre {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 0.85rem;
  line-height: 1.65;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--text-primary);
}

.url-error {
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.82rem;
  line-height: 1.6;
}

.url-parser {
  border-top: 1px solid var(--border-subtle);
  max-height: 260px;
  overflow-y: auto;
  background: var(--bg-surface);
}

.url-parts {
  padding: 8px 20px 12px;
}

.params-header {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  padding: 10px 0 4px;
  border-top: 1px solid var(--border-subtle);
  margin-top: 4px;
}

.param-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
}

.param-key,
.param-value {
  flex: 1;
  background: var(--bg-deep);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.8rem;
  padding: 5px 10px;
  outline: none;
  transition: border-color var(--transition);
}

.param-key:focus,
.param-value:focus {
  border-color: var(--accent-dim);
}

.param-key {
  color: var(--accent);
  max-width: 160px;
}

.param-eq {
  color: var(--text-muted);
  font-family: var(--font-mono);
  font-size: 0.85rem;
  flex-shrink: 0;
}

.param-remove {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.1rem;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: var(--radius);
  transition: all var(--transition);
  flex-shrink: 0;
}

.param-remove:hover {
  color: var(--red);
  background: rgba(212, 121, 121, 0.1);
}

.add-param-btn {
  margin-top: 6px;
  font-size: 0.68rem;
}
</style>
