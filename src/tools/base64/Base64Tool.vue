<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <button class="toolbar-btn" :class="{ active: direction === 'encode' }" @click="direction = 'encode'">Encode</button>
      <button class="toolbar-btn" :class="{ active: direction === 'decode' }" @click="direction = 'decode'">Decode</button>
      <div class="spacer"></div>
      <ShareButton tool="base64" :getState="() => ({ input: input, direction })" />
      <span class="tag" v-if="inputByteSize">{{ inputByteSize }}</span>
      <span class="tag green" v-if="outputByteSize">{{ outputByteSize }} out</span>
      <span class="tag red" v-if="error">Error</span>
      <button class="toolbar-btn" @click="copy(output, 'Copied')">Copy output</button>
      <button class="toolbar-btn" @click="clear">Clear</button>
    </div>
    <div class="tool-content">
      <SplitPane :initial-split="50">
        <template #left>
          <div class="pane editor-pane">
            <PaneHeader :label="direction === 'encode' ? 'Plain Text' : 'Base64 Input'">{{ input.length.toLocaleString() }} chars</PaneHeader>
            <textarea
              v-model="input"
              :placeholder="direction === 'encode' ? 'Enter text to encode...' : 'Paste base64 to decode...'"
              spellcheck="false"
              @input="onInput"
            ></textarea>
          </div>
        </template>
        <template #right>
          <div class="pane preview-pane">
            <PaneHeader :label="direction === 'encode' ? 'Base64 Output' : 'Decoded Text'">{{ output.length.toLocaleString() }} chars</PaneHeader>
            <div class="b64-output">
              <div v-if="error" class="b64-error">{{ error }}</div>
              <pre v-else class="b64-pre">{{ output }}</pre>
            </div>
          </div>
        </template>
      </SplitPane>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import SplitPane from '../../components/SplitPane.vue'
import PaneHeader from '../../components/PaneHeader.vue'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { copy } = useClipboard()

const input = ref('')
const output = ref('')
const direction = ref('encode')
const error = ref('')
let autoDetected = false

const { useShareable } = useShare()
useShareable('base64', (s) => {
  input.value = s.input || ''
  if (s.direction) direction.value = s.direction
  autoDetected = true
})

const BASE64_RE = /^(?:data:[a-z]+\/[a-z0-9.+-]+;base64,)?[A-Za-z0-9+/\n\r]+=*$/

function utf8ToBase64(str) {
  const bytes = new TextEncoder().encode(str)
  let binary = ''
  for (const b of bytes) binary += String.fromCharCode(b)
  return btoa(binary)
}

function base64ToUtf8(str) {
  // Strip data URI prefix if present
  const cleaned = str.replace(/^data:[a-z]+\/[a-z0-9.+-]+;base64,/i, '').trim()
  const binary = atob(cleaned)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return new TextDecoder('utf-8', { fatal: true }).decode(bytes)
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

const inputByteSize = computed(() => {
  if (!input.value) return ''
  const bytes = new TextEncoder().encode(input.value).length
  return formatBytes(bytes)
})

const outputByteSize = computed(() => {
  if (!output.value || error.value) return ''
  const bytes = new TextEncoder().encode(output.value).length
  return formatBytes(bytes)
})

function looksLikeBase64(str) {
  const trimmed = str.trim()
  if (trimmed.length < 4) return false
  if (trimmed.startsWith('data:') && trimmed.includes(';base64,')) return true
  if (trimmed.length % 4 !== 0 && !trimmed.includes('\n')) return false
  return BASE64_RE.test(trimmed)
}

function onInput() {
  // Auto-detect on first substantial paste
  if (!autoDetected && input.value.length > 8) {
    if (looksLikeBase64(input.value)) {
      direction.value = 'decode'
    }
    autoDetected = true
  }
}

function clear() {
  input.value = ''
  output.value = ''
  error.value = ''
  autoDetected = false
}

watch([input, direction], () => {
  if (!input.value) {
    output.value = ''
    error.value = ''
    return
  }
  try {
    if (direction.value === 'encode') {
      output.value = utf8ToBase64(input.value)
    } else {
      output.value = base64ToUtf8(input.value)
    }
    error.value = ''
  } catch (e) {
    error.value = e.message
    output.value = ''
  }
})
</script>

<style scoped>
.b64-output {
  flex: 1;
  overflow: auto;
  padding: 20px 24px;
}

.b64-pre {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 0.85rem;
  line-height: 1.65;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--text-primary);
}

.b64-error {
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.82rem;
  line-height: 1.6;
}
</style>
