<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <button class="toolbar-btn active" @click="setNow">Now</button>
      <button class="toolbar-btn" @click="setPreset('day')">Start of day</button>
      <button class="toolbar-btn" @click="setPreset('week')">Start of week</button>
      <button class="toolbar-btn" @click="setPreset('month')">Start of month</button>
      <button class="toolbar-btn" @click="setPreset('year')">Start of year</button>
      <div class="spacer"></div>
      <ShareButton tool="epoch" :getState="() => ({ epochInput, epochUnit, dateInput })" />
    </div>
    <div class="tool-scroll">
      <div class="epoch-current">
        <div class="epoch-current-label">Current Unix Timestamp</div>
        <div class="epoch-current-value" @click="copy(String(nowEpoch), 'Epoch copied')">{{ nowEpoch }}</div>
      </div>

      <div class="epoch-section">
        <div class="epoch-section-label">Epoch to Date</div>
        <div class="epoch-input-row">
          <input type="text" v-model="epochInput" placeholder="Enter epoch timestamp..." class="epoch-field">
          <select v-model="epochUnit">
            <option value="s">Seconds</option>
            <option value="ms">Milliseconds</option>
          </select>
        </div>
        <div class="epoch-results" v-if="epochDate">
          <div class="kv-row">
            <span class="kv-label">UTC</span>
            <span class="kv-value">{{ epochDate.toUTCString() }}</span>
          </div>
          <div class="kv-row">
            <span class="kv-label">Local</span>
            <span class="kv-value">{{ epochDate.toLocaleString() }}</span>
          </div>
          <div class="kv-row">
            <span class="kv-label">ISO 8601</span>
            <span class="kv-value" style="cursor: pointer" @click="copy(epochDate.toISOString(), 'ISO copied')">{{ epochDate.toISOString() }}</span>
          </div>
          <div class="kv-row">
            <span class="kv-label">Relative</span>
            <span class="kv-value">{{ relativeTime }}</span>
          </div>
        </div>
        <div v-else-if="epochInput" class="epoch-error">Invalid timestamp</div>
      </div>

      <div class="epoch-section">
        <div class="epoch-section-label">Date to Epoch</div>
        <div class="epoch-input-row">
          <input type="datetime-local" v-model="dateInput" class="epoch-field">
        </div>
        <div class="epoch-results" v-if="dateEpoch !== null">
          <div class="kv-row">
            <span class="kv-label">Seconds</span>
            <span class="kv-value" style="cursor: pointer" @click="copy(String(dateEpoch), 'Epoch copied')">{{ dateEpoch }}</span>
          </div>
          <div class="kv-row">
            <span class="kv-label">Milliseconds</span>
            <span class="kv-value" style="cursor: pointer" @click="copy(String(dateEpoch * 1000), 'Epoch (ms) copied')">{{ dateEpoch * 1000 }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { copy } = useClipboard()

const nowEpoch = ref(Math.floor(Date.now() / 1000))
const { useShareable } = useShare()

const epochInput = ref('')
const epochUnit = ref('s')
const dateInput = ref('')

useShareable('epoch', (s) => {
  if (s.epochInput) epochInput.value = s.epochInput
  if (s.epochUnit) epochUnit.value = s.epochUnit
  if (s.dateInput) dateInput.value = s.dateInput
})

let ticker = null
onMounted(() => {
  ticker = setInterval(() => {
    nowEpoch.value = Math.floor(Date.now() / 1000)
  }, 1000)
})
onUnmounted(() => clearInterval(ticker))

const epochDate = computed(() => {
  if (!epochInput.value) return null
  const n = Number(epochInput.value)
  if (isNaN(n)) return null
  const ms = epochUnit.value === 's' ? n * 1000 : n
  const d = new Date(ms)
  return isNaN(d.getTime()) ? null : d
})

const relativeTime = computed(() => {
  if (!epochDate.value) return ''
  const now = Date.now()
  const diff = now - epochDate.value.getTime()
  const absDiff = Math.abs(diff)
  const future = diff < 0

  if (absDiff < 60000) return future ? 'in a few seconds' : 'just now'
  if (absDiff < 3600000) {
    const m = Math.floor(absDiff / 60000)
    return future ? `in ${m} minute${m > 1 ? 's' : ''}` : `${m} minute${m > 1 ? 's' : ''} ago`
  }
  if (absDiff < 86400000) {
    const h = Math.floor(absDiff / 3600000)
    return future ? `in ${h} hour${h > 1 ? 's' : ''}` : `${h} hour${h > 1 ? 's' : ''} ago`
  }
  const d = Math.floor(absDiff / 86400000)
  return future ? `in ${d} day${d > 1 ? 's' : ''}` : `${d} day${d > 1 ? 's' : ''} ago`
})

const dateEpoch = computed(() => {
  if (!dateInput.value) return null
  const d = new Date(dateInput.value)
  return isNaN(d.getTime()) ? null : Math.floor(d.getTime() / 1000)
})

function setNow() {
  epochInput.value = String(Math.floor(Date.now() / 1000))
  epochUnit.value = 's'
}

function setPreset(type) {
  const now = new Date()
  let d
  switch (type) {
    case 'day': d = new Date(now.getFullYear(), now.getMonth(), now.getDate()); break
    case 'week':
      d = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      d.setDate(d.getDate() - d.getDay())
      break
    case 'month': d = new Date(now.getFullYear(), now.getMonth(), 1); break
    case 'year': d = new Date(now.getFullYear(), 0, 1); break
  }
  epochInput.value = String(Math.floor(d.getTime() / 1000))
  epochUnit.value = 's'
}
</script>

<style scoped>
.epoch-current {
  text-align: center;
  padding: 32px 0 24px;
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 24px;
}

.epoch-current-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.epoch-current-value {
  font-family: var(--font-mono);
  font-size: 2.2rem;
  font-weight: 300;
  color: var(--accent);
  cursor: pointer;
  transition: opacity var(--transition);
  letter-spacing: -0.02em;
}

.epoch-current-value:hover {
  opacity: 0.8;
}

.epoch-section {
  margin-bottom: 32px;
}

.epoch-section-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 10px;
}

.epoch-input-row {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.epoch-field {
  flex: 1;
  max-width: 400px;
}

.epoch-results {
  padding: 4px 0;
}

.epoch-error {
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.82rem;
}

input[type="datetime-local"] {
  background: var(--bg-deep);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.85rem;
  padding: 8px 12px;
  outline: none;
  color-scheme: dark;
}

input[type="datetime-local"]:focus {
  border-color: var(--accent-dim);
}
</style>
