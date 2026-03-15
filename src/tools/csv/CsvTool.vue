<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <select v-model="delimiter">
        <option value=",">Comma (,)</option>
        <option value="&#9;">Tab</option>
        <option value="|">Pipe (|)</option>
        <option value=";">Semicolon (;)</option>
      </select>
      <label class="flag-toggle">
        <input type="checkbox" v-model="hasHeader">
        <span>Header row</span>
      </label>
      <div class="spacer"></div>
      <ShareButton tool="csv" :getState="() => ({ input: input, delimiter, hasHeader })" />
      <span class="tag" v-if="rows.length">{{ rows.length }} rows</span>
      <span class="tag" v-if="columns.length">{{ columns.length }} cols</span>
      <button class="toolbar-btn" @click="downloadCsv">Download</button>
      <button class="toolbar-btn" @click="copy(input, 'CSV copied')">Copy</button>
    </div>
    <div class="tool-content">
      <SplitPane :initial-split="40">
        <template #left>
          <div class="pane editor-pane">
            <PaneHeader label="CSV Input" />
            <textarea
              v-model="input"
              placeholder="Paste CSV data here..."
              spellcheck="false"
              @paste="onPaste"
            ></textarea>
          </div>
        </template>
        <template #right>
          <div class="pane preview-pane">
            <PaneHeader label="Table">
              <span v-if="parseError" class="error-indicator">Parse error</span>
            </PaneHeader>
            <div class="data-table-wrapper" v-if="columns.length">
              <table class="data-table csv-table">
                <thead>
                  <tr>
                    <th class="row-num">#</th>
                    <th v-for="(col, i) in columns" :key="i">{{ col }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, ri) in rows" :key="ri">
                    <td class="row-num">{{ ri + 1 }}</td>
                    <td v-for="(cell, ci) in row" :key="ci">{{ cell }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="csv-empty">Paste or drop CSV data to see the table</div>
          </div>
        </template>
      </SplitPane>
    </div>
    <DropOverlay :active="dragging" text="Drop your CSV file" hint=".csv · .tsv" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import Papa from 'papaparse'
import SplitPane from '../../components/SplitPane.vue'
import PaneHeader from '../../components/PaneHeader.vue'
import DropOverlay from '../../components/DropOverlay.vue'
import { useToast } from '../../composables/useToast'
import { useClipboard } from '../../composables/useClipboard'
import { useDownload } from '../../composables/useDownload'
import { useFileDrop } from '../../composables/useFileDrop'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { toast } = useToast()
const { copy } = useClipboard()
const { download } = useDownload()

const input = ref('')
const delimiter = ref(',')
const hasHeader = ref(true)
const parseError = ref('')

const { useShareable } = useShare()
useShareable('csv', (s) => {
  input.value = s.input || ''
  if (s.delimiter) delimiter.value = s.delimiter
  if (s.hasHeader !== undefined) hasHeader.value = s.hasHeader
})

const { dragging } = useFileDrop((content, name) => {
  input.value = content
  if (name.endsWith('.tsv')) delimiter.value = '\t'
  else autoDetectDelimiter(content)
  toast(`Loaded ${name}`)
}, { extensions: ['.csv', '.tsv', '.txt'] })

function autoDetectDelimiter(text) {
  const firstLine = text.split('\n')[0] || ''
  const counts = {
    '\t': (firstLine.match(/\t/g) || []).length,
    '|': (firstLine.match(/\|/g) || []).length,
    ';': (firstLine.match(/;/g) || []).length,
    ',': (firstLine.match(/,/g) || []).length,
  }
  const best = Object.entries(counts).sort((a, b) => b[1] - a[1])[0]
  if (best && best[1] > 0) {
    delimiter.value = best[0]
  }
}

function onPaste() {
  setTimeout(() => {
    autoDetectDelimiter(input.value)
  }, 0)
}

const parsed = computed(() => {
  if (!input.value.trim()) return { columns: [], rows: [] }
  try {
    const result = Papa.parse(input.value.trim(), {
      delimiter: delimiter.value,
      header: false,
      skipEmptyLines: true,
    })
    parseError.value = ''
    const data = result.data
    if (!data.length) return { columns: [], rows: [] }

    if (hasHeader.value) {
      return {
        columns: data[0],
        rows: data.slice(1),
      }
    } else {
      const cols = data[0].map((_, i) => `Col ${i + 1}`)
      return { columns: cols, rows: data }
    }
  } catch (e) {
    parseError.value = e.message
    return { columns: [], rows: [] }
  }
})

const columns = computed(() => parsed.value.columns)
const rows = computed(() => parsed.value.rows)

function downloadCsv() {
  if (!input.value.trim()) return
  download(input.value, 'data.csv', 'text/csv')
  toast('CSV downloaded')
}
</script>

<style scoped>
.csv-empty {
  padding: 24px;
  color: var(--text-muted);
  font-style: italic;
  font-size: 0.85rem;
}

.csv-table thead th {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--bg-secondary);
}

.csv-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.03);
}

.row-num {
  color: var(--text-muted);
  font-size: 0.7rem;
  width: 40px;
  text-align: right;
}

.error-indicator {
  color: var(--red);
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
.flag-toggle span {
  padding: 2px 6px;
  border: 1px solid var(--border);
  border-radius: 3px;
  transition: all var(--transition);
}
.flag-toggle input:checked + span {
  color: var(--accent);
  border-color: var(--accent-dim);
  background: var(--accent-glow);
}
</style>
