<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <button class="toolbar-btn" :class="{ active: uppercase }" @click="uppercase = !uppercase">Uppercase</button>
      <div class="spacer"></div>
      <ShareButton tool="hash" :getState="() => ({ input: input, uppercase })" />
      <span class="tag" v-if="input">{{ byteSize }}</span>
    </div>
    <div class="hash-layout">
      <div class="pane editor-pane hash-input-pane">
        <PaneHeader label="Input">{{ input.length }} chars</PaneHeader>
        <textarea v-model="input" placeholder="Enter text to hash..." spellcheck="false"></textarea>
      </div>
      <div class="pane preview-pane hash-results-pane">
        <PaneHeader label="Hashes" />
        <div class="hash-results">
          <div v-if="!input" class="hash-empty">Enter text above to generate hashes</div>
          <template v-else>
            <div class="kv-row hash-row" v-for="h in hashes" :key="h.algo">
              <span class="kv-label">{{ h.algo }}</span>
              <span class="kv-value hash-value" @click="copy(h.value, h.algo + ' copied')">
                {{ h.value || 'Computing...' }}
                <span class="hash-copy-hint">click to copy</span>
              </span>
            </div>
          </template>
        </div>
      </div>
    </div>
    <DropOverlay :active="dragging" text="Drop a file to hash" hint="any file" />
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
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
const uppercase = ref(false)
const hashes = ref([
  { algo: 'SHA-1', value: '' },
  { algo: 'SHA-256', value: '' },
  { algo: 'SHA-384', value: '' },
  { algo: 'SHA-512', value: '' },
])

const { useShareable } = useShare()
useShareable('hash', (s) => {
  input.value = s.input || ''
  if (s.uppercase !== undefined) uppercase.value = s.uppercase
})

const { dragging } = useFileDrop((content, name) => {
  input.value = content
  toast(`Loaded ${name}`)
})

const byteSize = computed(() => {
  const bytes = new TextEncoder().encode(input.value).length
  if (bytes < 1024) return `${bytes} B`
  return `${(bytes / 1024).toFixed(1)} KB`
})

async function computeHash(algo, text) {
  const data = new TextEncoder().encode(text)
  const buffer = await crypto.subtle.digest(algo, data)
  const array = Array.from(new Uint8Array(buffer))
  const hex = array.map(b => b.toString(16).padStart(2, '0')).join('')
  return uppercase.value ? hex.toUpperCase() : hex
}

let hashGeneration = 0
watch([input, uppercase], async () => {
  const gen = ++hashGeneration
  if (!input.value) {
    hashes.value = hashes.value.map(h => ({ ...h, value: '' }))
    return
  }
  for (const h of hashes.value) {
    const val = await computeHash(h.algo, input.value)
    if (gen !== hashGeneration) return
    h.value = val
  }
}, { immediate: true })
</script>

<style scoped>
.hash-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.hash-input-pane {
  flex: 0 0 40%;
  min-height: 100px;
}

.hash-results-pane {
  flex: 1;
  border-top: 1px solid var(--border-subtle);
}

.hash-results {
  flex: 1;
  overflow-y: auto;
  padding: 12px 20px;
}

.hash-empty {
  padding: 12px 0;
  color: var(--text-muted);
  font-style: italic;
  font-size: 0.85rem;
}

.hash-row {
  cursor: pointer;
  padding: 10px 0;
  transition: background var(--transition);
  border-radius: var(--radius);
}

.hash-row:hover {
  background: var(--accent-glow);
}

.hash-value {
  position: relative;
  word-break: break-all;
}

.hash-copy-hint {
  font-size: 0.6rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-left: 8px;
  opacity: 0;
  transition: opacity var(--transition);
}

.hash-row:hover .hash-copy-hint {
  opacity: 1;
}
</style>
