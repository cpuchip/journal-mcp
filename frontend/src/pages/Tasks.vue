<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-800">Tasks</h1>
      <button class="btn btn-primary" @click="showCreateForm = true">
        ➕ New Task
      </button>
    </div>

    <!-- Filter Bar -->
    <div class="card mb-6">
      <div class="flex flex-wrap gap-4 items-center">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
          <select v-model="filters.status" class="form-select">
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
            <option value="paused">Paused</option>
            <option value="blocked">Blocked</option>
          </select>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
          <select v-model="filters.type" class="form-select">
            <option value="">All Types</option>
            <option value="work">Work</option>
            <option value="learning">Learning</option>
            <option value="personal">Personal</option>
            <option value="investigation">Investigation</option>
          </select>
        </div>
        
        <div>
          <button @click="loadTasks" class="btn btn-secondary">
            🔍 Filter
          </button>
        </div>
      </div>
    </div>

    <!-- Task List -->
    <div class="space-y-4">
      <div v-if="loading" class="text-center py-8">
        Loading tasks...
      </div>
      
      <div v-else-if="tasks.length === 0" class="card text-center py-8">
        <p class="text-gray-600">No tasks found. Create your first task to get started!</p>
      </div>
      
      <div v-else>
        <div 
          v-for="task in tasks" 
          :key="task.id"
          class="card task-card"
          @click="selectTask(task)"
        >
          <div class="flex justify-between items-start">
            <div class="flex-1">
              <h3 class="text-lg font-semibold text-gray-800">{{ task.title }}</h3>
              <p class="text-gray-600 text-sm mb-2">{{ task.id }}</p>
              
              <div class="flex flex-wrap gap-2 mb-3">
                <span class="badge badge-type">{{ task.type }}</span>
                <span class="badge badge-status" :class="getStatusColor(task.status)">
                  {{ task.status }}
                </span>
                <span v-if="task.priority" class="badge badge-priority">
                  {{ task.priority }}
                </span>
              </div>
              
              <div v-if="task.tags && task.tags.length" class="mb-2">
                <span 
                  v-for="tag in task.tags" 
                  :key="tag"
                  class="inline-block bg-gray-200 text-gray-700 text-xs px-2 py-1 rounded mr-2"
                >
                  {{ tag }}
                </span>
              </div>
              
              <p class="text-gray-500 text-sm">
                Updated: {{ formatDate(task.updated) }} | 
                Entries: {{ task.entries ? task.entries.length : 0 }}
              </p>
            </div>
            
            <div class="flex gap-2">
              <button 
                @click.stop="editTask(task)"
                class="btn-icon"
                title="Edit task"
              >
                ✏️
              </button>
              <button 
                @click.stop="viewTask(task)"
                class="btn-icon"
                title="View details"
              >
                👁️
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalTasks > tasks.length" class="card text-center mt-6">
      <button @click="loadMore" class="btn btn-secondary">
        Load More Tasks
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'

// Reactive data
const tasks = ref([])
const loading = ref(false)
const totalTasks = ref(0)
const showCreateForm = ref(false)

const filters = reactive({
  status: '',
  type: '',
  limit: 20,
  offset: 0
})

// Methods
const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString()
}

const getStatusColor = (status) => {
  const colors = {
    active: 'bg-blue-100 text-blue-800',
    completed: 'bg-green-100 text-green-800',
    paused: 'bg-yellow-100 text-yellow-800',
    blocked: 'bg-red-100 text-red-800'
  }
  return colors[status] || 'bg-gray-100 text-gray-800'
}

const loadTasks = async () => {
  loading.value = true
  try {
    // TODO: Implement API call to load tasks
    // For now, using mock data
    await new Promise(resolve => setTimeout(resolve, 500)) // Simulate API delay
    
    tasks.value = [
      {
        id: 'TEST-123',
        title: 'Test task from restructured code',
        type: 'work',
        status: 'active',
        priority: 'medium',
        tags: ['test', 'restructure'],
        updated: '2025-09-07T23:31:55.635856131Z',
        entries: [{ id: '1', content: 'Task created' }]
      }
    ]
    totalTasks.value = 1
  } catch (error) {
    console.error('Failed to load tasks:', error)
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  filters.offset += filters.limit
  loadTasks()
}

const selectTask = (task) => {
  console.log('Selected task:', task.id)
  // TODO: Navigate to task detail view or open modal
}

const editTask = (task) => {
  console.log('Edit task:', task.id)
  // TODO: Open edit modal
}

const viewTask = (task) => {
  console.log('View task:', task.id)
  // TODO: Open view modal
}

// Lifecycle
onMounted(() => {
  loadTasks()
})
</script>

<style scoped>
.task-card {
  cursor: pointer;
  transition: all 0.2s;
}

.task-card:hover {
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
}

.badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.badge-type {
  background-color: #e5e7eb;
  color: #374151;
}

.badge-status {
  /* Colors defined in getStatusColor method */
}

.badge-priority {
  background-color: #fbbf24;
  color: #92400e;
}

.form-select {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background-color: white;
}

.btn-icon {
  padding: 0.5rem;
  background: none;
  border: none;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-icon:hover {
  background-color: #f3f4f6;
}

.space-y-4 > * + * {
  margin-top: 1rem;
}
</style>