<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <button class="toolbar-btn active" @click="generate">Generate</button>
      <select v-model.number="count">
        <option :value="1">1</option>
        <option :value="5">5</option>
        <option :value="10">10</option>
        <option :value="25">25</option>
        <option :value="50">50</option>
      </select>
      <select v-model="format">
        <option value="lower">lowercase</option>
        <option value="upper">UPPERCASE</option>
        <option value="nodash">No dashes</option>
      </select>
      <div class="spacer"></div>
      <ShareButton tool="uuid" :getState="() => ({ uuids: [...uuids], count, format, validateInput })" />
      <span class="tag">{{ uuids.length }} generated</span>
      <button class="toolbar-btn" @click="copyAll">Copy all</button>
      <button class="toolbar-btn" @click="uuids = []">Clear</button>
    </div>
    <div class="tool-scroll">
      <div class="uuid-hero" v-if="uuids.length" @click="copy(uuids[0], 'UUID copied')">
        <div class="uuid-hero-label">Latest UUID</div>
        <div class="uuid-hero-value">{{ uuids[0] }}</div>
        <div class="uuid-hero-hint">click to copy</div>
      </div>
      <div class="uuid-hero uuid-hero-empty" v-else>
        <div class="uuid-hero-label">UUID Generator</div>
        <div class="uuid-hero-value" style="color: var(--text-muted)">Click generate to create UUIDs</div>
      </div>

      <div class="uuid-validate">
        <div class="uuid-section-label">Validate UUID</div>
        <div class="uuid-validate-row">
          <input type="text" v-model="validateInput" placeholder="Paste a UUID to validate..." class="uuid-validate-field">
          <span class="tag green" v-if="validation.valid">Valid</span>
          <span class="tag red" v-if="validateInput && !validation.valid">Invalid</span>
        </div>
        <div class="uuid-validate-info" v-if="validation.valid">
          <div class="kv-row">
            <span class="kv-label">Version</span>
            <span class="kv-value">{{ validation.version }}</span>
          </div>
          <div class="kv-row">
            <span class="kv-label">Variant</span>
            <span class="kv-value">{{ validation.variant }}</span>
          </div>
        </div>
      </div>

      <div class="uuid-list" v-if="uuids.length > 1">
        <div class="uuid-section-label">History</div>
        <div
          v-for="(uuid, i) in uuids"
          :key="i"
          class="uuid-item"
          @click="copy(uuid, 'UUID copied')"
        >
          <span class="uuid-item-index">{{ i + 1 }}</span>
          <span class="uuid-item-value">{{ uuid }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { copy } = useClipboard()

const { useShareable } = useShare()

const uuids = ref([])
const count = ref(1)
const format = ref('lower')
const validateInput = ref('')

useShareable('uuid', (s) => {
  if (s.uuids) uuids.value = s.uuids
  if (s.count) count.value = s.count
  if (s.format) format.value = s.format
  if (s.validateInput) validateInput.value = s.validateInput
})

function formatUuid(uuid) {
  if (format.value === 'upper') return uuid.toUpperCase()
  if (format.value === 'nodash') return uuid.replace(/-/g, '')
  return uuid
}

function generate() {
  const newUuids = []
  for (let i = 0; i < count.value; i++) {
    newUuids.push(formatUuid(crypto.randomUUID()))
  }
  uuids.value = [...newUuids, ...uuids.value]
}

function copyAll() {
  copy(uuids.value.join('\n'), `${uuids.value.length} UUIDs copied`)
}

const validation = computed(() => {
  const input = validateInput.value.trim()
  if (!input) return { valid: false }
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-([0-9a-f])[0-9a-f]{3}-([0-9a-f])[0-9a-f]{3}-[0-9a-f]{12}$/i
  const match = input.match(uuidRegex)
  if (!match) return { valid: false }

  const version = parseInt(match[1], 16)
  const variantChar = parseInt(match[2], 16)
  let variant = 'Unknown'
  if ((variantChar & 0x8) === 0) variant = 'NCS (reserved)'
  else if ((variantChar & 0xc) === 0x8) variant = 'RFC 4122'
  else if ((variantChar & 0xe) === 0xc) variant = 'Microsoft (reserved)'
  else variant = 'Future (reserved)'

  return { valid: true, version: `v${version}`, variant }
})
</script>

<style scoped>
.uuid-hero {
  text-align: center;
  padding: 36px 24px;
  margin-bottom: 24px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  transition: border-color var(--transition);
}

.uuid-hero:hover {
  border-color: var(--accent-dim);
}

.uuid-hero-empty {
  cursor: default;
}

.uuid-hero-empty:hover {
  border-color: var(--border);
}

.uuid-hero-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 10px;
}

.uuid-hero-value {
  font-family: var(--font-mono);
  font-size: 1.4rem;
  font-weight: 400;
  color: var(--accent);
  letter-spacing: 0.02em;
  word-break: break-all;
}

.uuid-hero-hint {
  font-size: 0.6rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-top: 10px;
  opacity: 0;
  transition: opacity var(--transition);
}

.uuid-hero:hover .uuid-hero-hint {
  opacity: 1;
}

.uuid-section-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 10px;
}

.uuid-validate {
  margin-bottom: 24px;
}

.uuid-validate-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.uuid-validate-field {
  flex: 1;
  max-width: 440px;
}

.uuid-validate-info {
  padding: 4px 0;
}

.uuid-list {
  margin-bottom: 24px;
}

.uuid-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 12px;
  cursor: pointer;
  border-radius: var(--radius);
  transition: background var(--transition);
}

.uuid-item:hover {
  background: var(--accent-glow);
}

.uuid-item-index {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  color: var(--text-muted);
  width: 28px;
  text-align: right;
  flex-shrink: 0;
}

.uuid-item-value {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text-primary);
  word-break: break-all;
}
</style>
