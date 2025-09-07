# Vue.js Frontend Quick Reference for Phase 4

## 🎯 Vue.js Technology Stack

### Core Dependencies
```json
{
  "dependencies": {
    "vue": "^3.3.0",                    // Vue 3 with Composition API
    "vue-router": "^4.2.0",             // Client-side routing
    "pinia": "^2.1.0",                  // State management (Vue 3 official)
    "@tanstack/vue-query": "^4.24.0",   // Server state management
    "date-fns": "^2.29.0",              // Date utilities
    "chart.js": "^4.2.0",               // Charts and visualizations
    "vue-chartjs": "^5.2.0",            // Vue wrapper for Chart.js
    "@headlessui/vue": "^1.7.0",        // Unstyled UI components
    "lucide-vue-next": "^0.315.0",      // Icon library
    "clsx": "^1.2.0"                    // Conditional CSS classes
  }
}
```

### Build Configuration (vite.config.js)
```javascript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    minify: 'esbuild',
    rollupOptions: {
      output: {
        manualChunks: undefined, // Single bundle for embedding
      }
    }
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080'
    }
  }
})
```

## 📱 Vue.js Application Structure

### Main Entry Point (main.js)
```javascript
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { createPinia } from 'pinia'
import { VueQueryPlugin } from '@tanstack/vue-query'
import App from './App.vue'
import router from './router'
import './assets/css/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(VueQueryPlugin)
app.mount('#app')
```

### Component Structure
```
frontend/src/
├── components/
│   ├── ui/                 # Base UI components
│   │   ├── Button.vue
│   │   ├── Modal.vue
│   │   ├── Card.vue
│   │   └── Input.vue
│   ├── task/               # Task-related components
│   │   ├── TaskList.vue
│   │   ├── TaskCard.vue
│   │   ├── TaskForm.vue
│   │   └── TaskFilters.vue
│   ├── analytics/          # Analytics components
│   │   ├── Dashboard.vue
│   │   ├── Charts.vue
│   │   └── MetricsCard.vue
│   └── github/             # GitHub integration
│       ├── RepoList.vue
│       ├── IssueSync.vue
│       └── AuthSetup.vue
├── pages/
│   ├── Dashboard.vue
│   ├── Tasks.vue
│   ├── Analytics.vue
│   ├── Calendar.vue
│   ├── GitHub.vue
│   └── Settings.vue
├── composables/            # Composition API utilities
│   ├── useApi.js
│   ├── useTasks.js
│   ├── useAuth.js
│   └── useWebSocket.js
├── stores/                 # Pinia stores
│   ├── tasks.js
│   ├── auth.js
│   └── settings.js
└── utils/
    ├── api.js
    ├── date.js
    └── formatters.js
```

## 🔧 Key Vue.js Patterns for Journal MCP

### Composition API Example (Task Management)
```vue
<template>
  <div class="task-manager">
    <TaskFilters v-model:filters="filters" />
    <TaskList :tasks="tasks" :loading="isLoading" />
    <TaskForm @submit="createTask" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { useTasks } from '@/composables/useTasks'

const filters = ref({
  status: 'all',
  type: 'all',
  dateRange: null
})

const { 
  tasks, 
  isLoading, 
  createTask, 
  updateTask 
} = useTasks(filters)
</script>
```

### Pinia Store Example (Task Store)
```javascript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/utils/api'

export const useTaskStore = defineStore('tasks', () => {
  const tasks = ref([])
  const currentTask = ref(null)
  
  const activeTasks = computed(() => 
    tasks.value.filter(task => task.status === 'active')
  )
  
  const completedTasks = computed(() =>
    tasks.value.filter(task => task.status === 'completed')
  )
  
  async function fetchTasks(filters = {}) {
    const response = await api.get('/tasks', { params: filters })
    tasks.value = response.data
  }
  
  async function createTask(taskData) {
    const response = await api.post('/tasks', taskData)
    tasks.value.push(response.data)
    return response.data
  }
  
  return {
    tasks,
    currentTask,
    activeTasks,
    completedTasks,
    fetchTasks,
    createTask
  }
})
```

### WebSocket Integration with Vue
```javascript
// composables/useWebSocket.js
import { ref, onMounted, onUnmounted } from 'vue'

export function useWebSocket(url) {
  const isConnected = ref(false)
  const socket = ref(null)
  const lastMessage = ref(null)
  
  const connect = () => {
    socket.value = new WebSocket(url)
    
    socket.value.onopen = () => {
      isConnected.value = true
    }
    
    socket.value.onmessage = (event) => {
      lastMessage.value = JSON.parse(event.data)
    }
    
    socket.value.onclose = () => {
      isConnected.value = false
    }
  }
  
  const disconnect = () => {
    if (socket.value) {
      socket.value.close()
    }
  }
  
  const send = (data) => {
    if (socket.value && isConnected.value) {
      socket.value.send(JSON.stringify(data))
    }
  }
  
  onMounted(connect)
  onUnmounted(disconnect)
  
  return {
    isConnected,
    lastMessage,
    send,
    connect,
    disconnect
  }
}
```

## 🎨 UI Component Examples

### Task Card Component
```vue
<template>
  <div class="task-card" :class="statusClass">
    <div class="task-header">
      <h3>{{ task.title }}</h3>
      <span class="task-status">{{ task.status }}</span>
    </div>
    
    <div class="task-meta">
      <span class="task-type">{{ task.type }}</span>
      <span class="task-date">{{ formatDate(task.created_at) }}</span>
    </div>
    
    <div class="task-actions">
      <button @click="editTask" class="btn-edit">Edit</button>
      <button @click="toggleStatus" class="btn-toggle">
        {{ task.status === 'active' ? 'Complete' : 'Reactivate' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { formatDate } from '@/utils/date'

const props = defineProps({
  task: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['edit', 'toggle-status'])

const statusClass = computed(() => ({
  'task-active': props.task.status === 'active',
  'task-completed': props.task.status === 'completed',
  'task-paused': props.task.status === 'paused'
}))

const editTask = () => emit('edit', props.task)
const toggleStatus = () => emit('toggle-status', props.task)
</script>
```

## 🚀 Vue.js Advantages for Journal MCP

### Why Vue.js is Perfect for This Project:
1. **Gentle Learning Curve**: Easy to pick up and highly productive
2. **Composition API**: Perfect for complex state management and reusability
3. **Excellent Tooling**: Vite provides lightning-fast development experience
4. **Bundle Size**: Smaller production bundles than React
5. **Template Syntax**: More familiar and readable than JSX
6. **Official Libraries**: Pinia, Vue Router provide complete ecosystem

### Offline-First Architecture:
- **Service Worker**: Cache API responses and static assets
- **IndexedDB**: Store tasks and entries locally via Pinia persistence
- **Progressive Enhancement**: App works offline, syncs when online
- **Optimistic Updates**: Immediate UI feedback, background sync

This Vue.js setup will provide a modern, efficient, and maintainable frontend for the Journal MCP system!
