<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <span class="tag green" v-if="description && !error">Valid</span>
      <span class="tag red" v-if="error">Invalid</span>
      <div class="spacer"></div>
      <ShareButton tool="cron" :getState="() => ({ expression })" />
      <button class="toolbar-btn" @click="copy(expression, 'Cron expression copied')">Copy</button>
    </div>
    <div class="tool-scroll">
      <div class="cron-input-section">
        <input
          type="text"
          v-model="expression"
          placeholder="* * * * *"
          class="cron-input"
          spellcheck="false"
        >
        <div class="cron-description" v-if="description">{{ description }}</div>
        <div class="cron-error" v-if="error">{{ error }}</div>
      </div>

      <div class="cron-presets">
        <div class="cron-section-label">Presets</div>
        <div class="cron-preset-grid">
          <button
            v-for="p in presets"
            :key="p.expr"
            class="cron-preset"
            :class="{ 'cron-preset-active': expression.trim() === p.expr }"
            @click="expression = p.expr"
          >
            <span class="cron-preset-expr">{{ p.expr }}</span>
            <span class="cron-preset-label">{{ p.label }}</span>
          </button>
        </div>
      </div>

      <div class="cron-fields" v-if="fields.length && !error">
        <div class="cron-section-label">Field Breakdown</div>
        <div class="cron-field-visual">
          <span
            v-for="(f, i) in fields"
            :key="f.name"
            class="cron-field-segment"
            :class="{ 'cron-field-active': activeField === i }"
            @click="activeField = activeField === i ? -1 : i"
          >{{ f.value }}</span>
        </div>
        <div class="cron-field-detail" v-if="activeField >= 0 && fields[activeField]">
          <span class="cron-field-detail-name">{{ fields[activeField].name }}</span>
          <span class="cron-field-detail-desc">{{ fields[activeField].desc }}</span>
        </div>
        <div class="kv-row" v-for="f in fields" :key="f.name">
          <span class="kv-label">{{ f.name }}</span>
          <span class="kv-value">
            <code>{{ f.value }}</code>
            <span class="field-desc">{{ f.desc }}</span>
          </span>
        </div>
      </div>

      <div class="cron-next" v-if="nextRuns.length">
        <div class="cron-section-label">Next {{ nextRuns.length }} Runs</div>
        <div class="cron-run-list">
          <div v-for="(run, i) in nextRuns" :key="i" class="cron-run-item">
            <span class="cron-run-index">{{ i + 1 }}</span>
            <span class="cron-run-time">{{ formatDate(run) }}</span>
            <span class="cron-run-rel">{{ relativeTime(run) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import cronstrue from 'cronstrue'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { copy } = useClipboard()

const { useShareable } = useShare()

const expression = ref('0 9 * * 1-5')

useShareable('cron', (s) => {
  if (s.expression) expression.value = s.expression
})
const activeField = ref(-1)

const presets = [
  { expr: '* * * * *', label: 'Every minute' },
  { expr: '0 * * * *', label: 'Every hour' },
  { expr: '0 0 * * *', label: 'Daily at midnight' },
  { expr: '0 9 * * 1-5', label: 'Weekdays at 9am' },
  { expr: '0 0 * * 1', label: 'Weekly on Monday' },
  { expr: '0 0 1 * *', label: 'Monthly on 1st' },
  { expr: '0 0 1 1 *', label: 'Yearly on Jan 1' },
  { expr: '*/5 * * * *', label: 'Every 5 minutes' },
  { expr: '*/15 * * * *', label: 'Every 15 minutes' },
  { expr: '0 */6 * * *', label: 'Every 6 hours' },
]

const error = ref('')

const description = computed(() => {
  const expr = expression.value.trim()
  if (!expr) { error.value = ''; return '' }
  try {
    const desc = cronstrue.toString(expr)
    error.value = ''
    return desc
  } catch (e) {
    error.value = e.toString()
    return ''
  }
})

const fieldNames5 = ['Minute', 'Hour', 'Day of Month', 'Month', 'Day of Week']
const fieldNames6 = ['Second', 'Minute', 'Hour', 'Day of Month', 'Month', 'Day of Week']

const fieldRanges5 = [
  { min: 0, max: 59 },
  { min: 0, max: 23 },
  { min: 1, max: 31 },
  { min: 1, max: 12 },
  { min: 0, max: 7 },
]

const fieldRanges6 = [
  { min: 0, max: 59 },
  { min: 0, max: 59 },
  { min: 0, max: 23 },
  { min: 1, max: 31 },
  { min: 1, max: 12 },
  { min: 0, max: 7 },
]

const fields = computed(() => {
  const parts = expression.value.trim().split(/\s+/)
  if (parts.length < 5 || parts.length > 6) return []
  const names = parts.length === 6 ? fieldNames6 : fieldNames5
  const ranges = parts.length === 6 ? fieldRanges6 : fieldRanges5
  return parts.map((val, i) => ({
    name: names[i],
    value: val,
    desc: describeField(val, names[i], ranges[i]),
  }))
})

function describeField(val, name, range) {
  const lower = name.toLowerCase()
  if (val === '*') return `Every ${lower}`
  if (val.includes(',')) {
    const parts = val.split(',')
    if (val.includes('/') || val.includes('-')) {
      return parts.map(p => describeFieldPart(p, lower, range)).join('; ')
    }
    return `At ${lower} ${parts.join(', ')}`
  }
  return describeFieldPart(val, lower, range)
}

function describeFieldPart(val, lower, range) {
  if (val.includes('/')) {
    const [base, step] = val.split('/')
    const from = base === '*' ? '' : ` starting at ${base}`
    return `Every ${step} ${lower}(s)${from}`
  }
  if (val.includes('-')) {
    const [a, b] = val.split('-')
    return `${lower} ${a} through ${b}`
  }
  return `At ${lower} ${val}`
}

// --- Cron next-run calculator ---

function expandField(field, min, max) {
  if (field === '*') return null
  const values = new Set()
  for (const part of field.split(',')) {
    if (part.includes('/')) {
      const [range, step] = part.split('/')
      const s = parseInt(step)
      let start = min, end = max
      if (range !== '*') {
        if (range.includes('-')) {
          const [a, b] = range.split('-').map(Number)
          start = a
          end = b
        } else {
          start = parseInt(range)
        }
      }
      for (let i = start; i <= end; i += s) values.add(i)
    } else if (part.includes('-')) {
      const [a, b] = part.split('-').map(Number)
      for (let i = a; i <= b; i++) values.add(i)
    } else {
      values.add(parseInt(part))
    }
  }
  return values
}

function matchesField(val, field, min, max) {
  const expanded = expandField(field, min, max)
  if (!expanded) return true
  return expanded.has(val)
}

const nextRuns = computed(() => {
  const parts = expression.value.trim().split(/\s+/)
  if (parts.length !== 5 || error.value) return []

  const [minF, hourF, domF, monF, dowF] = parts
  const runs = []
  const now = new Date()
  const candidate = new Date(
    now.getFullYear(), now.getMonth(), now.getDate(),
    now.getHours(), now.getMinutes() + 1, 0, 0,
  )

  let safety = 0
  const maxIter = 525960 // roughly one year of minutes
  while (runs.length < 5 && safety < maxIter) {
    const min = candidate.getMinutes()
    const hour = candidate.getHours()
    const dom = candidate.getDate()
    const mon = candidate.getMonth() + 1
    const dow = candidate.getDay()

    if (
      matchesField(min, minF, 0, 59) &&
      matchesField(hour, hourF, 0, 23) &&
      matchesField(dom, domF, 1, 31) &&
      matchesField(mon, monF, 1, 12) &&
      (matchesField(dow, dowF, 0, 7) || (dow === 0 && matchesField(7, dowF, 0, 7)))
    ) {
      runs.push(new Date(candidate))
    }
    candidate.setMinutes(candidate.getMinutes() + 1)
    safety++
  }
  return runs
})

function formatDate(date) {
  const pad = (n) => String(n).padStart(2, '0')
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  return `${days[date.getDay()]} ${months[date.getMonth()]} ${date.getDate()}, ${date.getFullYear()} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function relativeTime(date) {
  const diff = date.getTime() - Date.now()
  if (diff < 0) return 'past'
  if (diff < 60000) return 'in < 1 min'
  if (diff < 3600000) {
    const m = Math.floor(diff / 60000)
    return `in ${m} min`
  }
  if (diff < 86400000) {
    const h = Math.floor(diff / 3600000)
    const m = Math.floor((diff % 3600000) / 60000)
    return m > 0 ? `in ${h}h ${m}m` : `in ${h}h`
  }
  const d = Math.floor(diff / 86400000)
  const h = Math.floor((diff % 86400000) / 3600000)
  return h > 0 ? `in ${d}d ${h}h` : `in ${d} day${d > 1 ? 's' : ''}`
}
</script>

<style scoped>
.cron-input-section {
  margin-bottom: 28px;
  text-align: center;
}

.cron-input {
  width: 100%;
  max-width: 500px;
  font-family: var(--font-mono) !important;
  font-size: 1.6rem !important;
  letter-spacing: 0.08em;
  padding: 14px 20px !important;
  text-align: center;
  background: var(--bg-surface) !important;
}

.cron-description {
  margin-top: 12px;
  font-size: 1rem;
  color: var(--text-secondary);
  font-style: italic;
  font-family: var(--font-serif);
}

.cron-error {
  margin-top: 12px;
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.8rem;
}

.cron-section-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 10px;
}

/* Presets */
.cron-presets {
  margin-bottom: 28px;
}

.cron-preset-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.cron-preset {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: transparent;
  cursor: pointer;
  transition: all var(--transition);
  text-align: left;
  font-family: var(--font-sans);
  color: var(--text-secondary);
}

.cron-preset:hover {
  background: var(--bg-hover);
  border-color: var(--accent-dim);
}

.cron-preset-active {
  border-color: var(--accent-dim);
  background: var(--accent-glow);
}

.cron-preset-expr {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--accent);
}

.cron-preset-label {
  font-size: 0.68rem;
  color: var(--text-muted);
}

/* Visual field editor */
.cron-field-visual {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin-bottom: 12px;
  padding: 14px 0;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
}

.cron-field-segment {
  font-family: var(--font-mono);
  font-size: 1.2rem;
  padding: 4px 12px;
  border-radius: var(--radius);
  cursor: pointer;
  transition: all var(--transition);
  color: var(--text-primary);
}

.cron-field-segment:hover {
  background: var(--bg-hover);
  color: var(--accent);
}

.cron-field-active {
  background: var(--accent-glow);
  color: var(--accent);
}

.cron-field-detail {
  text-align: center;
  margin-bottom: 14px;
  padding: 8px 0;
}

.cron-field-detail-name {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--accent);
  margin-right: 10px;
}

.cron-field-detail-desc {
  font-size: 0.82rem;
  color: var(--text-secondary);
}

/* Field breakdown */
.cron-fields {
  margin-bottom: 28px;
}

.cron-fields code {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--accent);
  background: var(--bg-raised);
  padding: 1px 6px;
  border-radius: 3px;
  border: 1px solid var(--border-subtle);
}

.field-desc {
  margin-left: 10px;
  font-family: var(--font-sans);
  font-size: 0.78rem;
  color: var(--text-muted);
}

/* Next runs */
.cron-next {
  margin-bottom: 28px;
}

.cron-run-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.cron-run-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-subtle);
}

.cron-run-item:last-child {
  border-bottom: none;
}

.cron-run-index {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  color: var(--text-muted);
  width: 20px;
  text-align: right;
  flex-shrink: 0;
}

.cron-run-time {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text-primary);
}

.cron-run-rel {
  font-size: 0.72rem;
  color: var(--text-muted);
  font-style: italic;
  margin-left: auto;
}
</style>
