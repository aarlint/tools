<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <button class="toolbar-btn" :class="{ active: autoFormat }" @click="autoFormat = !autoFormat">Auto-format</button>
      <select v-model.number="indent">
        <option :value="2">2 spaces</option>
        <option :value="4">4 spaces</option>
        <option :value="1">Tab</option>
      </select>
      <div class="spacer"></div>
      <button class="toolbar-btn" @click="format">Format</button>
      <button class="toolbar-btn" @click="minify">Minify</button>
      <button class="toolbar-btn" @click="sortKeys">Sort keys</button>
      <ShareButton tool="json" :getState="() => ({ input: input, indent, autoFormat })" />
      <button class="toolbar-btn" @click="copy(output, 'JSON copied')">Copy</button>
    </div>
    <div class="tool-content">
      <SplitPane :initial-split="50">
        <template #left>
          <div class="pane editor-pane">
            <PaneHeader label="Input">
              <span v-if="error" class="error-indicator">Error</span>
              <span v-else-if="input.trim()" class="valid-indicator">Valid</span>
            </PaneHeader>
            <textarea
              v-model="input"
              placeholder="Paste or type JSON here..."
              spellcheck="false"
              @paste="onPaste"
            ></textarea>
          </div>
        </template>
        <template #right>
          <div class="pane preview-pane">
            <PaneHeader label="Output">
              <span v-if="jsonPath" class="json-path">{{ jsonPath }}</span>
            </PaneHeader>
            <div class="json-output" ref="outputEl">
              <div v-if="error" class="json-error">{{ error }}</div>
              <pre v-else><code v-html="highlighted"></code></pre>
            </div>
          </div>
        </template>
      </SplitPane>
    </div>
    <DropOverlay :active="dragging" text="Drop your JSON file" hint=".json" />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import SplitPane from '../../components/SplitPane.vue'
import PaneHeader from '../../components/PaneHeader.vue'
import DropOverlay from '../../components/DropOverlay.vue'
import { useToast } from '../../composables/useToast'
import { useClipboard } from '../../composables/useClipboard'
import { useFileDrop } from '../../composables/useFileDrop'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { toast } = useToast()
const { copy } = useClipboard()

const input = ref('')
const indent = ref(2)
const autoFormat = ref(true)
const error = ref('')
const output = ref('')
const jsonPath = ref('')
const outputEl = ref(null)

const { useShareable } = useShare()
useShareable('json', (s) => {
  input.value = s.input || ''
  if (s.indent !== undefined) indent.value = s.indent
  if (s.autoFormat !== undefined) autoFormat.value = s.autoFormat
})

const { dragging } = useFileDrop((content, name) => {
  input.value = content
  toast(`Loaded ${name}`)
}, { extensions: ['.json'] })

function tryParse(str) {
  try {
    return { data: JSON.parse(str), error: null }
  } catch (e) {
    return { data: null, error: e.message }
  }
}

function sortObj(obj) {
  if (Array.isArray(obj)) return obj.map(sortObj)
  if (obj && typeof obj === 'object') {
    return Object.keys(obj).sort().reduce((acc, key) => {
      acc[key] = sortObj(obj[key])
      return acc
    }, {})
  }
  return obj
}

function format() {
  const { data, error: err } = tryParse(input.value)
  if (err) { error.value = err; output.value = ''; return }
  const indentStr = indent.value === 1 ? '\t' : indent.value
  output.value = JSON.stringify(data, null, indentStr)
  error.value = ''
}

function minify() {
  const { data, error: err } = tryParse(input.value)
  if (err) { error.value = err; output.value = ''; return }
  output.value = JSON.stringify(data)
  error.value = ''
}

function sortKeys() {
  const { data, error: err } = tryParse(input.value)
  if (err) { error.value = err; output.value = ''; return }
  const indentStr = indent.value === 1 ? '\t' : indent.value
  output.value = JSON.stringify(sortObj(data), null, indentStr)
  error.value = ''
}

function onPaste() {
  if (autoFormat.value) {
    setTimeout(format, 0)
  }
}

// Syntax highlighting for JSON
const highlighted = computed(() => {
  if (!output.value) return ''
  return output.value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"([^"\\]*(\\.[^"\\]*)*)"(\s*:)/g, '<span class="json-key">"$1"</span>$3')
    .replace(/"([^"\\]*(\\.[^"\\]*)*)"/g, '<span class="json-string">"$1"</span>')
    .replace(/\b(-?\d+\.?\d*([eE][+-]?\d+)?)\b/g, '<span class="json-number">$1</span>')
    .replace(/\b(true|false)\b/g, '<span class="json-bool">$1</span>')
    .replace(/\bnull\b/g, '<span class="json-null">null</span>')
})

watch(input, (val) => {
  if (!val.trim()) {
    output.value = ''
    error.value = ''
    return
  }
  const { data, error: err } = tryParse(val)
  if (err) {
    error.value = err
    output.value = ''
  } else {
    error.value = ''
    if (autoFormat.value) {
      const indentStr = indent.value === 1 ? '\t' : indent.value
      output.value = JSON.stringify(data, null, indentStr)
    } else {
      output.value = val
    }
  }
})

watch(indent, () => {
  if (output.value && !error.value) format()
})
</script>

<style scoped>
.json-output {
  flex: 1;
  overflow: auto;
  padding: 20px 24px;
}

.json-output pre {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 0.82rem;
  line-height: 1.65;
}

.json-output code {
  background: transparent;
  border: none;
  padding: 0;
  color: var(--text-primary);
  font-size: inherit;
}

.json-error {
  font-family: var(--font-mono);
  font-size: 0.82rem;
  color: var(--red);
  padding: 4px 0;
  line-height: 1.6;
}

.error-indicator {
  color: var(--red);
  font-family: var(--font-mono);
}

.valid-indicator {
  color: var(--green);
  font-family: var(--font-mono);
}

.json-path {
  font-family: var(--font-mono);
  color: var(--accent);
}

:deep(.json-key) { color: var(--blue); }
:deep(.json-string) { color: var(--green); }
:deep(.json-number) { color: var(--accent); }
:deep(.json-bool) { color: var(--purple); }
:deep(.json-null) { color: var(--text-muted); }
</style>
