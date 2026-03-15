import { reactive, onActivated } from 'vue'

// Shared state store for passing share data between ShareView and tool components
const pendingShare = reactive({
  tool: null,
  state: null,
})

export function useShare() {
  async function createShare(tool, state, ttl) {
    const res = await fetch('/api/share', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tool, state, ttl }),
    })
    if (!res.ok) {
      throw new Error(`Failed to create share: ${res.statusText}`)
    }
    const data = await res.json()
    return {
      id: data.id,
      url: `${window.location.origin}/s/${data.id}`,
      expiresAt: data.expires_at,
    }
  }

  async function getShare(id) {
    const res = await fetch(`/api/share/${id}`)
    if (!res.ok) {
      throw new Error(res.status === 404 ? 'Share not found or expired' : 'Failed to load share')
    }
    return await res.json()
  }

  function setPendingShare(tool, state) {
    pendingShare.tool = tool
    pendingShare.state = state
  }

  function consumePendingShare(tool) {
    if (pendingShare.tool === tool && pendingShare.state) {
      const state = pendingShare.state
      pendingShare.tool = null
      pendingShare.state = null
      return state
    }
    return null
  }

  // Use in tool setup() - consumes pending share on mount and keep-alive reactivation
  function useShareable(tool, loadState) {
    const shared = consumePendingShare(tool)
    if (shared) loadState(shared)

    onActivated(() => {
      const shared = consumePendingShare(tool)
      if (shared) loadState(shared)
    })
  }

  return { createShare, getShare, setPendingShare, consumePendingShare, pendingShare, useShareable }
}
