<template>
  <div class="share-view">
    <div v-if="error" class="share-error">
      <div class="share-error-icon">?</div>
      <div class="share-error-text">{{ error }}</div>
      <button class="toolbar-btn" @click="$router.push('/')">Go Home</button>
    </div>
    <div v-else class="share-loading">
      <div class="share-loading-text">Loading share...</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useShare } from '../composables/useShare'

const route = useRoute()
const router = useRouter()
const { getShare, setPendingShare } = useShare()

const error = ref('')

// Map tool names to their router paths
const toolPaths = {
  md: '/md',
  html: '/html',
  json: '/json',
  diff: '/diff',
  regex: '/regex',
  csv: '/csv',
  sql: '/sql',
  base64: '/base64',
  jwt: '/jwt',
  url: '/url',
  hash: '/hash',
  color: '/color',
  epoch: '/epoch',
  uuid: '/uuid',
  cron: '/cron',
  qr: '/qr',
}

onMounted(async () => {
  const id = route.params.id
  try {
    const data = await getShare(id)
    const path = toolPaths[data.tool]
    if (!path) {
      error.value = 'Unknown tool type'
      return
    }
    setPendingShare(data.tool, data.state)
    router.replace(path)
  } catch (e) {
    error.value = e.message || 'Failed to load share'
  }
})
</script>

<style scoped>
.share-view {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-deep);
}

.share-loading {
  text-align: center;
}

.share-loading-text {
  font-family: var(--font-sans);
  font-size: 0.9rem;
  color: var(--text-muted);
}

.share-error {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.share-error-icon {
  font-size: 2rem;
  color: var(--text-muted);
}

.share-error-text {
  font-family: var(--font-sans);
  font-size: 0.9rem;
  color: var(--text-secondary);
}
</style>
