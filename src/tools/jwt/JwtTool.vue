<template>
  <div class="tool-page">
    <div class="tool-toolbar">
      <span class="tag" :class="expiryClass" v-if="decoded">{{ expiryLabel }}</span>
      <span class="tag" v-if="decoded">{{ decoded.header.alg || 'unknown' }}</span>
      <div class="spacer"></div>
      <ShareButton tool="jwt" :getState="() => ({ token })" />
      <button class="toolbar-btn" @click="copy(headerJson, 'Header copied')">Copy header</button>
      <button class="toolbar-btn" @click="copy(payloadJson, 'Payload copied')">Copy payload</button>
    </div>
    <div class="tool-content">
      <SplitPane :initial-split="45">
        <template #left>
          <div class="pane editor-pane">
            <PaneHeader label="JWT Token" />
            <textarea v-model="token" placeholder="Paste JWT token here..." spellcheck="false"></textarea>
          </div>
        </template>
        <template #right>
          <div class="pane preview-pane">
            <PaneHeader label="Decoded">
              <span v-if="error" style="color: var(--red)">Invalid token</span>
            </PaneHeader>
            <div class="jwt-decoded" v-if="decoded">
              <div class="jwt-section">
                <div class="jwt-section-label">Header</div>
                <pre class="jwt-json"><code v-html="highlightJson(headerJson)"></code></pre>
              </div>
              <div class="jwt-section">
                <div class="jwt-section-label">Payload</div>
                <pre class="jwt-json"><code v-html="highlightJson(payloadJson)"></code></pre>
              </div>
              <div class="jwt-section" v-if="claims.length">
                <div class="jwt-section-label">Claims</div>
                <div class="jwt-claims">
                  <div class="kv-row" v-for="c in claims" :key="c.key">
                    <span class="kv-label">{{ c.key }}</span>
                    <span class="kv-value">{{ c.value }}</span>
                  </div>
                </div>
              </div>
            </div>
            <div v-else-if="error" class="jwt-error">{{ error }}</div>
            <div v-else class="jwt-empty">Paste a JWT token to decode it</div>
          </div>
        </template>
      </SplitPane>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import SplitPane from '../../components/SplitPane.vue'
import PaneHeader from '../../components/PaneHeader.vue'
import { useClipboard } from '../../composables/useClipboard'
import { useShare } from '../../composables/useShare'
import ShareButton from '../../components/ShareButton.vue'

const { copy } = useClipboard()

const { useShareable } = useShare()

const token = ref('')

useShareable('jwt', (s) => {
  if (s.token) token.value = s.token
})
const decoded = ref(null)
const error = ref('')

function decodeBase64Url(str) {
  let base64 = str.replace(/-/g, '+').replace(/_/g, '/')
  while (base64.length % 4) base64 += '='
  return JSON.parse(atob(base64))
}

watch(token, (val) => {
  const trimmed = val.trim()
  if (!trimmed) { decoded.value = null; error.value = ''; return }
  const parts = trimmed.split('.')
  if (parts.length !== 3) { decoded.value = null; error.value = 'JWT must have 3 parts (header.payload.signature)'; return }
  try {
    const header = decodeBase64Url(parts[0])
    const payload = decodeBase64Url(parts[1])
    decoded.value = { header, payload, signature: parts[2] }
    error.value = ''
  } catch (e) {
    decoded.value = null
    error.value = 'Failed to decode: ' + e.message
  }
})

const headerJson = computed(() => decoded.value ? JSON.stringify(decoded.value.header, null, 2) : '')
const payloadJson = computed(() => decoded.value ? JSON.stringify(decoded.value.payload, null, 2) : '')

const claims = computed(() => {
  if (!decoded.value) return []
  const p = decoded.value.payload
  const items = []
  if (p.iss) items.push({ key: 'Issuer', value: p.iss })
  if (p.sub) items.push({ key: 'Subject', value: p.sub })
  if (p.aud) items.push({ key: 'Audience', value: Array.isArray(p.aud) ? p.aud.join(', ') : p.aud })
  if (p.iat) items.push({ key: 'Issued At', value: new Date(p.iat * 1000).toLocaleString() })
  if (p.exp) items.push({ key: 'Expires', value: new Date(p.exp * 1000).toLocaleString() })
  if (p.nbf) items.push({ key: 'Not Before', value: new Date(p.nbf * 1000).toLocaleString() })
  if (p.jti) items.push({ key: 'JWT ID', value: p.jti })
  return items
})

const expiryLabel = computed(() => {
  if (!decoded.value?.payload?.exp) return 'No expiry'
  const now = Date.now() / 1000
  return decoded.value.payload.exp < now ? 'Expired' : 'Valid'
})

const expiryClass = computed(() => {
  if (!decoded.value?.payload?.exp) return ''
  const now = Date.now() / 1000
  return decoded.value.payload.exp < now ? 'red' : 'green'
})

function highlightJson(json) {
  if (!json) return ''
  return json
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/"([^"\\]*(\\.[^"\\]*)*)"(\s*:)/g, '<span class="json-key">"$1"</span>$3')
    .replace(/"([^"\\]*(\\.[^"\\]*)*)"/g, '<span class="json-string">"$1"</span>')
    .replace(/\b(-?\d+\.?\d*)\b/g, '<span class="json-number">$1</span>')
    .replace(/\b(true|false)\b/g, '<span class="json-bool">$1</span>')
    .replace(/\bnull\b/g, '<span class="json-null">null</span>')
}
</script>

<style scoped>
.jwt-decoded {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
}

.jwt-section {
  margin-bottom: 20px;
}

.jwt-section-label {
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.jwt-json {
  margin: 0;
  padding: 14px 16px;
  background: var(--bg-deep);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-family: var(--font-mono);
  font-size: 0.82rem;
  line-height: 1.65;
  overflow-x: auto;
}

.jwt-json code {
  background: transparent;
  border: none;
  padding: 0;
  color: var(--text-primary);
  font-size: inherit;
}

.jwt-claims {
  padding: 4px 0;
}

.jwt-error {
  padding: 20px;
  color: var(--red);
  font-family: var(--font-mono);
  font-size: 0.82rem;
}

.jwt-empty {
  padding: 24px;
  color: var(--text-muted);
  font-style: italic;
  font-size: 0.85rem;
}

:deep(.json-key) { color: var(--blue); }
:deep(.json-string) { color: var(--green); }
:deep(.json-number) { color: var(--accent); }
:deep(.json-bool) { color: var(--purple); }
:deep(.json-null) { color: var(--text-muted); }
</style>
