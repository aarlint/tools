<template>
  <AppSidebar />
  <main class="main-content">
    <router-view v-slot="{ Component }">
      <keep-alive :max="5">
        <component :is="Component" />
      </keep-alive>
    </router-view>
  </main>
  <AppToast />

  <!-- Mobile nav toggle -->
  <button class="mobile-nav-toggle" @click="mobileNavOpen = !mobileNavOpen" :class="{ active: mobileNavOpen }">
    <span></span><span></span><span></span>
  </button>
  <div class="mobile-nav-backdrop" v-if="mobileNavOpen" @click="mobileNavOpen = false"></div>
  <teleport to="body">
    <div class="mobile-sidebar" :class="{ open: mobileNavOpen }">
      <AppSidebar @navigate="mobileNavOpen = false" />
    </div>
  </teleport>
</template>

<script setup>
import { ref } from 'vue'
import AppSidebar from './components/AppSidebar.vue'
import AppToast from './components/AppToast.vue'

const mobileNavOpen = ref(false)
</script>

<style>
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

/* Desktop: sidebar always visible, mobile nav hidden */
@media (min-width: 769px) {
  .mobile-nav-toggle,
  .mobile-nav-backdrop,
  .mobile-sidebar { display: none !important; }
}

/* Mobile: hide desktop sidebar, show hamburger */
@media (max-width: 768px) {
  #app > .sidebar { display: none; }

  .mobile-nav-toggle {
    position: fixed;
    top: 10px;
    left: 10px;
    z-index: 1001;
    width: 36px;
    height: 36px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg-raised);
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: 0;
  }

  .mobile-nav-toggle span {
    display: block;
    width: 16px;
    height: 1.5px;
    background: var(--text-secondary);
    transition: all var(--transition);
  }

  .mobile-nav-toggle.active span:nth-child(1) {
    transform: translateY(5.5px) rotate(45deg);
  }
  .mobile-nav-toggle.active span:nth-child(2) { opacity: 0; }
  .mobile-nav-toggle.active span:nth-child(3) {
    transform: translateY(-5.5px) rotate(-45deg);
  }

  .mobile-nav-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: 999;
  }

  .mobile-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: var(--sidebar-width);
    z-index: 1000;
    transform: translateX(-100%);
    transition: transform 250ms ease;
  }

  .mobile-sidebar.open {
    transform: translateX(0);
  }

  .mobile-sidebar .sidebar {
    display: flex !important;
    height: 100%;
  }
}
</style>
