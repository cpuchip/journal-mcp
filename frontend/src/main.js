import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { createPinia } from 'pinia'
import App from './App.vue'

// Import pages
import Dashboard from './pages/Dashboard.vue'
import Tasks from './pages/Tasks.vue'
import Analytics from './pages/Analytics.vue'
import Settings from './pages/Settings.vue'

// Import styles
import './style.css'

// Create router
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'Dashboard', component: Dashboard },
    { path: '/tasks', name: 'Tasks', component: Tasks },
    { path: '/analytics', name: 'Analytics', component: Analytics },
    { path: '/settings', name: 'Settings', component: Settings },
  ]
})

// Create Pinia store
const pinia = createPinia()

// Create and mount app
const app = createApp(App)
app.use(router)
app.use(pinia)
app.mount('#app')