<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800 mb-6">Dashboard</h1>
    
    <!-- Quick Stats -->
    <div class="grid grid-cols-1 md:grid-cols-3 mb-6">
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-2">Active Tasks</h3>
        <p class="text-3xl font-bold text-blue-600">{{ stats.activeTasks }}</p>
        <p class="text-gray-600 text-sm">Currently in progress</p>
      </div>
      
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-2">Completed Today</h3>
        <p class="text-3xl font-bold text-green-600">{{ stats.completedToday }}</p>
        <p class="text-gray-600 text-sm">Tasks finished today</p>
      </div>
      
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-2">Total Entries</h3>
        <p class="text-3xl font-bold text-purple-600">{{ stats.totalEntries }}</p>
        <p class="text-gray-600 text-sm">Journal entries made</p>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="card mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">Quick Actions</h3>
      <div class="flex flex-wrap gap-3">
        <button 
          @click="showCreateTask = true"
          class="btn btn-primary"
        >
          ➕ New Task
        </button>
        <button 
          @click="showQuickEntry = true"
          class="btn btn-secondary"
        >
          📝 Quick Entry
        </button>
        <router-link to="/analytics" class="btn btn-secondary">
          📊 View Analytics
        </router-link>
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="card">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">Recent Activity</h3>
      <div v-if="recentActivity.length === 0" class="text-gray-600 text-center py-8">
        No recent activity. Create your first task to get started!
      </div>
      <div v-else class="space-y-3">
        <div 
          v-for="activity in recentActivity" 
          :key="activity.id"
          class="border-l-4 border-blue-400 pl-4 py-2"
        >
          <p class="font-medium">{{ activity.title }}</p>
          <p class="text-gray-600 text-sm">{{ activity.description }}</p>
          <p class="text-gray-500 text-xs">{{ formatDate(activity.timestamp) }}</p>
        </div>
      </div>
    </div>

    <!-- Modals would go here -->
    <!-- For now, using simple alerts -->
    <div v-if="showCreateTask" class="modal-overlay" @click="showCreateTask = false">
      <div class="modal-content" @click.stop>
        <h3 class="text-lg font-semibold mb-4">Create New Task</h3>
        <p class="text-gray-600 mb-4">
          Task creation UI would go here. For now, this is a placeholder.
        </p>
        <div class="flex gap-3">
          <button @click="showCreateTask = false" class="btn btn-primary">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'

// Reactive data
const stats = reactive({
  activeTasks: 0,
  completedToday: 0,
  totalEntries: 0
})

const recentActivity = ref([])
const showCreateTask = ref(false)
const showQuickEntry = ref(false)

// Methods
const formatDate = (date) => {
  return new Date(date).toLocaleString()
}

const loadDashboardData = async () => {
  // TODO: Implement API calls to load real data
  // For now, using mock data
  stats.activeTasks = 5
  stats.completedToday = 2
  stats.totalEntries = 48
  
  recentActivity.value = [
    {
      id: 1,
      title: 'Task TEST-123 created',
      description: 'Test task from restructured code',
      timestamp: new Date()
    }
  ]
}

// Lifecycle
onMounted(() => {
  loadDashboardData()
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 2rem;
  border-radius: 0.5rem;
  max-width: 500px;
  width: 90vw;
}

.space-y-3 > * + * {
  margin-top: 0.75rem;
}

@media (min-width: 768px) {
  .grid-cols-1.md\\:grid-cols-3 {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>