import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/variables.css'
import './styles/base.css'
import './styles/shared.css'

createApp(App).use(router).mount('#app')
