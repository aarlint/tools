<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <select v-model="mode">
        <option value="lines">Line diff</option>
        <option value="words">Word diff</option>
        <option value="chars">Char diff</option>
      </select>
      <button class="toolbar-btn" @click="swap">Swap</button>
      <div class="spacer"></div>
      <ShareButton tool="diff" :getState="() => ({ original, modified, mode })" />
      <span class="tag green" v-if="stats.added">+{{ stats.added }}</span>
      <span class="tag red" v-if="stats.removed">-{{ stats.removed }}</span>
      <span class="tag" v-if="stats.unchanged">~{{ stats.unchanged }}</span>
      <button class="toolbar-btn" @click="copyDiff">Copy</button>
      <button class="toolbar-btn" @click="clear">Clear</button>
    </div>
    <div class="diff-layout">
      <div class="diff-inputs">
        <div class="pane editor-pane">
          <PaneHeader label="Original" />
          <textarea v-model="original" placeholder="Original text..." spellcheck="false"></textarea>
        </div>
        <div class="diff-input-divider"></div>
        <div class="pane editor-pane">
          <PaneHeader label="Modified" />
          <textarea v-model="modified" placeholder="Modified text..." spellcheck="false"></textarea>
        </div>
      </div>
      <div class="pane preview-pane diff-output-pane">
        <PaneHeader label="Diff">
          <span v-if="original || modified">{{ parts.length }} chunks</span>
        </PaneHeader>
        <div class="diff-output" ref="outputEl">
          <template v-if="lines.length">
            <div
              v-for="(line, i) in lines"
              :key="i"
              class="diff-line"
              :class="{
                'diff-added': line.type === 'added',
                'diff-removed': line.type === 'removed'
              }"
            ><span class="diff-line-num">{{ line.num ?? '' }}</span><span class="diff-prefix">{{ line.prefix }}</span><span class="diff-text">{{ line.text }}</span></div>
          </template>
          <div v-else class="diff-empty">Enter text in both panes to see the diff</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { diffLines, diffWords, diffChars } from 'diff'
import PaneHeader from '../../components/PaneHeader.vue'
import { useToast } from '../../composables/useToast'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const toast = useToast()
const { copy } = useClipboard()

const original = ref('')
const modified = ref('')
const mode = ref('lines')

const { useShareable } = useShare()
useShareable('diff', (s) => {
  original.value = s.original || ''
  modified.value = s.modified || ''
  if (s.mode) mode.value = s.mode
})

const diffFns = { lines: diffLines, words: diffWords, chars: diffChars }

const parts = computed(() => {
  if (!original.value && !modified.value) return []
  const fn = diffFns[mode.value] || diffLines
  return fn(original.value, modified.value)
})

const lines = computed(() => {
  const result = []
  let num = 1
  for (const part of parts.value) {
    const prefix = part.added ? '+' : part.removed ? '-' : ' '
    const type = part.added ? 'added' : part.removed ? 'removed' : 'unchanged'

    if (mode.value === 'lines') {
      const textLines = part.value.replace(/\n$/, '').split('\n')
      for (const text of textLines) {
        result.push({ num: part.removed ? '' : num, prefix, type, text })
        if (!part.removed) num++
      }
    } else {
      result.push({ num: num++, prefix, type, text: part.value })
    }
  }
  return result
})

const stats = computed(() => {
  let added = 0
  let removed = 0
  let unchanged = 0
  for (const p of parts.value) {
    const count = p.count ?? p.value.length
    if (p.added) added += count
    else if (p.removed) removed += count
    else unchanged += count
  }
  return { added, removed, unchanged }
})

const diffText = computed(() => {
  return parts.value
    .map(p => {
      const prefix = p.added ? '+ ' : p.removed ? '- ' : '  '
      return prefix + p.value
    })
    .join('')
})

function copyDiff() {
  copy(diffText.value)
  toast.show('Diff copied')
}

function swap() {
  const tmp = original.value
  original.value = modified.value
  modified.value = tmp
}

function clear() {
  original.value = ''
  modified.value = ''
}
</script>

<style scoped>
.diff-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.diff-inputs {
  display: flex;
  flex: 0 0 40%;
  min-height: 0;
  overflow: hidden;
}

.diff-inputs .pane {
  flex: 1;
}

.diff-input-divider {
  width: 1px;
  background: var(--border-subtle);
  flex-shrink: 0;
}

.diff-output-pane {
  flex: 1;
  border-top: 1px solid var(--border-subtle);
}

.diff-output {
  flex: 1;
  overflow: auto;
  padding: 0;
  font-family: var(--font-mono);
  font-size: 0.82rem;
  line-height: 1.65;
}

.diff-line {
  display: flex;
  padding: 1px 20px;
  white-space: pre-wrap;
  word-break: break-all;
}

.diff-added {
  background: rgba(126, 200, 155, 0.08);
  color: var(--green);
}

.diff-removed {
  background: rgba(212, 121, 121, 0.08);
  color: var(--red);
}

.diff-line-num {
  display: inline-block;
  width: 40px;
  text-align: right;
  padding-right: 12px;
  color: var(--text-muted);
  user-select: none;
  flex-shrink: 0;
  opacity: 0.5;
}

.diff-prefix {
  display: inline-block;
  width: 16px;
  color: var(--text-muted);
  user-select: none;
  flex-shrink: 0;
}

.diff-text {
  white-space: pre-wrap;
  flex: 1;
  min-width: 0;
}

.diff-empty {
  padding: 24px;
  color: var(--text-muted);
  font-style: italic;
  font-family: var(--font-sans);
  font-size: 0.85rem;
}
</style>
