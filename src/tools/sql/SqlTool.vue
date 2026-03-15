<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <button class="toolbar-btn active" @click="runQuery">Run (Ctrl+Enter)</button>
      <button class="toolbar-btn" @click="loadSample">Load sample data</button>
      <div class="spacer"></div>
      <ShareButton tool="sql" :getState="() => ({ query })" />
      <span class="tag" v-if="queryTime !== null">{{ queryTime }}ms</span>
      <span class="tag" v-if="resultRows.length">{{ resultRows.length }} rows</span>
      <button class="toolbar-btn" @click="clearAll">Clear</button>
    </div>
    <div class="sql-layout">
      <div class="pane editor-pane sql-editor-pane">
        <PaneHeader label="SQL Query">
          <span v-if="dbReady" class="valid-indicator">SQLite ready</span>
          <span v-else class="loading-indicator">Loading...</span>
        </PaneHeader>
        <textarea
          v-model="query"
          placeholder="SELECT * FROM users;"
          spellcheck="false"
          @keydown.ctrl.enter.prevent="runQuery"
          @keydown.meta.enter.prevent="runQuery"
        ></textarea>
      </div>
      <div class="pane preview-pane sql-results-pane">
        <PaneHeader label="Results">
          <span v-if="error" class="error-indicator">Error</span>
        </PaneHeader>
        <div class="sql-results">
          <div v-if="error" class="sql-error">{{ error }}</div>
          <div v-else-if="!resultColumns.length" class="sql-empty">
            {{ dbReady ? 'Run a query to see results' : 'Loading SQLite...' }}
          </div>
          <div v-else class="data-table-wrapper">
            <table class="data-table sql-results-table">
              <thead>
                <tr>
                  <th v-for="col in resultColumns" :key="col">{{ col }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, i) in resultRows" :key="i">
                  <td v-for="col in resultColumns" :key="col">{{ row[col] }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-if="message" class="sql-message">{{ message }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import initSqlJs from 'sql.js'
import PaneHeader from '../../components/PaneHeader.vue'
import { useToast } from '../../composables/useToast'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { toast } = useToast()

const { useShareable } = useShare()

const query = ref('SELECT * FROM users;')

useShareable('sql', (s) => {
  if (s.query) query.value = s.query
})
const resultColumns = ref([])
const resultRows = ref([])
const error = ref('')
const message = ref('')
const queryTime = ref(null)
const dbReady = ref(false)

let db = null

onMounted(async () => {
  try {
    const SQL = await initSqlJs({
      locateFile: file => `https://sql.js.org/dist/${file}`,
    })
    db = new SQL.Database()
    dbReady.value = true
    loadSample()
  } catch (e) {
    error.value = 'Failed to load SQLite: ' + e.message
  }
})

onUnmounted(() => {
  if (db) db.close()
})

function runQuery() {
  if (!db || !query.value.trim()) return
  error.value = ''
  message.value = ''
  resultColumns.value = []
  resultRows.value = []

  const start = performance.now()
  try {
    const results = db.exec(query.value)
    queryTime.value = Math.round(performance.now() - start)

    if (results.length > 0) {
      const res = results[0]
      resultColumns.value = res.columns
      resultRows.value = res.values.map(row => {
        const obj = {}
        res.columns.forEach((col, i) => { obj[col] = row[i] })
        return obj
      })
    } else {
      message.value = `Query executed successfully (${queryTime.value}ms)`
    }
  } catch (e) {
    queryTime.value = Math.round(performance.now() - start)
    error.value = e.message
  }
}

function loadSample() {
  if (!db) return
  try {
    db.run('DROP TABLE IF EXISTS orders;')
    db.run('DROP TABLE IF EXISTS users;')
    db.run(`CREATE TABLE users (
      id INTEGER PRIMARY KEY,
      name TEXT NOT NULL,
      email TEXT,
      role TEXT DEFAULT 'user',
      created_at TEXT
    );`)
    db.run(`INSERT INTO users VALUES
      (1, 'Alice Chen', 'alice@example.com', 'admin', '2024-01-15'),
      (2, 'Bob Smith', 'bob@example.com', 'user', '2024-02-20'),
      (3, 'Carol Davis', 'carol@example.com', 'editor', '2024-03-10'),
      (4, 'Dan Wilson', 'dan@example.com', 'user', '2024-04-05'),
      (5, 'Eve Martinez', 'eve@example.com', 'admin', '2024-05-12');`)
    db.run(`CREATE TABLE orders (
      id INTEGER PRIMARY KEY,
      user_id INTEGER,
      product TEXT,
      amount REAL,
      created_at TEXT,
      FOREIGN KEY (user_id) REFERENCES users(id)
    );`)
    db.run(`INSERT INTO orders VALUES
      (1, 1, 'Widget A', 29.99, '2024-06-01'),
      (2, 2, 'Widget B', 49.99, '2024-06-02'),
      (3, 1, 'Widget C', 19.99, '2024-06-03'),
      (4, 3, 'Widget A', 29.99, '2024-06-04'),
      (5, 4, 'Widget D', 99.99, '2024-06-05'),
      (6, 5, 'Widget B', 49.99, '2024-06-06'),
      (7, 2, 'Widget A', 29.99, '2024-06-07');`)
    toast('Sample data loaded (users, orders)')
    runQuery()
  } catch (e) {
    error.value = e.message
  }
}

function clearAll() {
  query.value = ''
  resultColumns.value = []
  resultRows.value = []
  error.value = ''
  message.value = ''
  queryTime.value = null
}
</script>

<style scoped>
.sql-layout {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sql-editor-pane {
  flex: 0 0 35%;
  min-height: 120px;
}

.sql-results-pane {
  flex: 1;
  border-top: 1px solid var(--border-subtle);
}

.sql-results {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
}

.sql-results-table thead th {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--bg-secondary);
}

.sql-results-table tbody tr:hover {
  background: rgba(255, 255, 255, 0.03);
}

.sql-error {
  padding: 16px 20px;
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.82rem;
  line-height: 1.5;
}

.sql-empty {
  padding: 24px;
  color: var(--text-muted);
  font-style: italic;
  font-size: 0.85rem;
}

.sql-message {
  padding: 12px 20px;
  color: var(--green);
  font-family: var(--font-mono);
  font-size: 0.82rem;
}

.error-indicator {
  color: var(--red);
  font-family: var(--font-mono);
}

.valid-indicator {
  color: var(--green);
  font-family: var(--font-mono);
}

.loading-indicator {
  color: var(--text-muted);
  font-family: var(--font-mono);
}
</style>
