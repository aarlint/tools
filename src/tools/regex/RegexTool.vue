<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <span class="tag" v-if="matches.length">{{ matches.length }} match{{ matches.length !== 1 ? 'es' : '' }}</span>
      <span class="tag red" v-if="regexError">Invalid</span>
      <div class="spacer"></div>
      <ShareButton tool="regex" :getState="() => ({ pattern, flags: [...flags], testString })" />
      <label class="flag-toggle" v-for="f in flagOptions" :key="f.value">
        <input type="checkbox" :value="f.value" v-model="flags">
        <span>{{ f.label }}</span>
      </label>
      <button class="toolbar-btn" @click="copy(matchSummary, 'Matches copied')">Copy matches</button>
    </div>
    <div class="regex-layout">
      <div class="regex-input-row">
        <span class="regex-slash">/</span>
        <input
          type="text"
          v-model="pattern"
          placeholder="enter regex pattern..."
          class="regex-field"
          spellcheck="false"
        >
        <span class="regex-slash">/{{ flags.join('') }}</span>
      </div>
      <div class="regex-body">
        <div class="pane editor-pane regex-test-pane">
          <PaneHeader label="Test String">{{ testString.length }} chars</PaneHeader>
          <textarea v-model="testString" placeholder="Enter test string..." spellcheck="false"></textarea>
        </div>
        <div class="pane preview-pane regex-results-pane">
          <PaneHeader label="Results">
            <span v-if="matches.length">{{ matches.length }} match{{ matches.length !== 1 ? 'es' : '' }}</span>
          </PaneHeader>
          <div class="regex-results">
            <div v-if="regexError" class="regex-error">{{ regexError }}</div>
            <div v-else-if="!pattern" class="regex-empty">Enter a regex pattern above</div>
            <div v-else-if="!matches.length && testString" class="regex-empty">No matches found</div>
            <div v-else-if="!testString" class="regex-empty">Enter a test string</div>
            <template v-else>
              <div class="highlighted-text" v-html="highlightedText"></div>
              <div class="match-list">
                <div v-for="(m, i) in matches" :key="i" class="match-item">
                  <div class="match-header">
                    <span class="tag green">Match {{ i + 1 }}</span>
                    <span class="match-index">index {{ m.index }}</span>
                  </div>
                  <div class="match-value">{{ m.value }}</div>
                  <div v-if="m.groups.length" class="match-groups">
                    <div v-for="(g, gi) in m.groups" :key="gi" class="match-group">
                      <span class="group-label">Group {{ gi + 1 }}:</span>
                      <span class="group-value">{{ g ?? 'undefined' }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import PaneHeader from '../../components/PaneHeader.vue'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { copy } = useClipboard()

const pattern = ref('')
const flags = ref(['g'])
const testString = ref('')
const regexError = ref('')

const { useShareable } = useShare()
useShareable('regex', (s) => {
  pattern.value = s.pattern || ''
  if (s.flags) flags.value = s.flags
  testString.value = s.testString || ''
})

const flagOptions = [
  { label: 'g', value: 'g' },
  { label: 'i', value: 'i' },
  { label: 'm', value: 'm' },
  { label: 's', value: 's' },
  { label: 'u', value: 'u' },
]

const regex = computed(() => {
  if (!pattern.value) { regexError.value = ''; return null }
  try {
    const r = new RegExp(pattern.value, flags.value.join(''))
    regexError.value = ''
    return r
  } catch (e) {
    regexError.value = e.message
    return null
  }
})

const matches = computed(() => {
  if (!regex.value || !testString.value) return []
  const r = regex.value
  const results = []
  if (r.global) {
    let m
    const re = new RegExp(r.source, r.flags)
    while ((m = re.exec(testString.value)) !== null) {
      results.push({
        value: m[0],
        index: m.index,
        groups: m.slice(1),
      })
      if (!m[0].length) re.lastIndex++
      if (results.length > 500) break
    }
  } else {
    const m = r.exec(testString.value)
    if (m) {
      results.push({
        value: m[0],
        index: m.index,
        groups: m.slice(1),
      })
    }
  }
  return results
})

const highlightedText = computed(() => {
  if (!matches.value.length || !testString.value) return ''
  const text = testString.value
  const parts = []
  let lastIndex = 0
  for (const m of matches.value) {
    if (m.index > lastIndex) {
      parts.push(esc(text.slice(lastIndex, m.index)))
    }
    parts.push(`<mark class="regex-match">${esc(m.value)}</mark>`)
    lastIndex = m.index + m.value.length
  }
  if (lastIndex < text.length) {
    parts.push(esc(text.slice(lastIndex)))
  }
  return parts.join('')
})

const matchSummary = computed(() => {
  return matches.value.map((m, i) => `Match ${i + 1}: "${m.value}" at index ${m.index}`).join('\n')
})

function esc(str) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}
</script>

<style scoped>
.regex-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.regex-input-row {
  display: flex;
  align-items: center;
  gap: 0;
  padding: 12px 20px;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-subtle);
}

.regex-slash {
  font-family: var(--font-mono);
  font-size: 1.1rem;
  color: var(--accent);
  flex-shrink: 0;
}

.regex-field {
  flex: 1;
  background: transparent !important;
  border: none !important;
  font-size: 1rem !important;
  padding: 4px 8px !important;
  color: var(--text-primary);
}

.regex-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.regex-test-pane {
  flex: 1;
}

.regex-results-pane {
  flex: 1;
  border-left: 1px solid var(--border-subtle);
}

.regex-results {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
}

.regex-error {
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.82rem;
}

.regex-empty {
  color: var(--text-muted);
  font-style: italic;
  font-size: 0.85rem;
}

.highlighted-text {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  line-height: 1.7;
  padding: 12px 16px;
  background: var(--bg-deep);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  white-space: pre-wrap;
  word-break: break-all;
  margin-bottom: 16px;
  color: var(--text-primary);
}

:deep(.regex-match) {
  background: rgba(196, 149, 106, 0.25);
  color: var(--accent);
  border-radius: 2px;
  padding: 1px 2px;
}

.match-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.match-item {
  padding: 10px 14px;
  background: var(--bg-deep);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius);
}

.match-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.match-index {
  font-family: var(--font-mono);
  font-size: 0.7rem;
  color: var(--text-muted);
}

.match-value {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text-primary);
  padding: 4px 0;
}

.match-groups {
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px solid var(--border-subtle);
}

.match-group {
  display: flex;
  gap: 8px;
  font-size: 0.8rem;
  padding: 2px 0;
}

.group-label {
  color: var(--text-muted);
  font-family: var(--font-mono);
  font-size: 0.72rem;
}

.group-value {
  color: var(--purple);
  font-family: var(--font-mono);
}

.flag-toggle {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  font-family: var(--font-mono);
  font-size: 0.78rem;
  color: var(--text-muted);
  user-select: none;
}

.flag-toggle input { display: none; }
.flag-toggle input:checked + span { color: var(--accent); }
.flag-toggle span {
  padding: 2px 6px;
  border: 1px solid var(--border);
  border-radius: 3px;
  transition: all var(--transition);
}
.flag-toggle input:checked + span {
  border-color: var(--accent-dim);
  background: var(--accent-glow);
}
</style>
