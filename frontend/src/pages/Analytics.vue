<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800 mb-6">Analytics</h1>
    
    <!-- Time Period Selector -->
    <div class="card mb-6">
      <div class="flex flex-wrap gap-4 items-center">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Time Period</label>
          <select v-model="selectedPeriod" @change="loadAnalytics" class="form-select">
            <option value="week">This Week</option>
            <option value="month">This Month</option>
            <option value="quarter">This Quarter</option>
            <option value="year">This Year</option>
            <option value="all">All Time</option>
          </select>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Task Type</label>
          <select v-model="selectedType" @change="loadAnalytics" class="form-select">
            <option value="">All Types</option>
            <option value="work">Work</option>
            <option value="learning">Learning</option>
            <option value="personal">Personal</option>
            <option value="investigation">Investigation</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Key Metrics -->
    <div class="grid grid-cols-1 md:grid-cols-4 mb-6">
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-2">Total Tasks</h3>
        <p class="text-3xl font-bold text-blue-600">{{ analytics.taskMetrics.totalTasks }}</p>
        <p class="text-gray-600 text-sm">Created in period</p>
      </div>
      
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-2">Completion Rate</h3>
        <p class="text-3xl font-bold text-green-600">
          {{ Math.round(analytics.taskMetrics.completionRate * 100) }}%
        </p>
        <p class="text-gray-600 text-sm">Tasks completed</p>
      </div>
      
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-2">Avg Entries</h3>
        <p class="text-3xl font-bold text-purple-600">
          {{ analytics.taskMetrics.averageEntries.toFixed(1) }}
        </p>
        <p class="text-gray-600 text-sm">Per task</p>
      </div>
      
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-2">Productivity Score</h3>
        <p class="text-3xl font-bold text-orange-600">
          {{ analytics.productivityMetrics.productivityScore.toFixed(1) }}
        </p>
        <p class="text-gray-600 text-sm">Out of 10</p>
      </div>
    </div>

    <!-- Charts Section -->
    <div class="grid grid-cols-1 lg:grid-cols-2 mb-6">
      <!-- Task Status Distribution -->
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">Task Status Distribution</h3>
        <div class="chart-placeholder">
          <!-- Placeholder for chart -->
          <div class="space-y-2">
            <div v-for="(count, status) in analytics.taskMetrics.byStatus" :key="status" 
                 class="flex justify-between items-center">
              <span class="capitalize">{{ status }}</span>
              <div class="flex items-center gap-2">
                <div class="w-20 h-4 bg-gray-200 rounded overflow-hidden">
                  <div 
                    class="h-full rounded transition-all duration-500"
                    :class="getStatusBarColor(status)"
                    :style="{ width: `${(count / analytics.taskMetrics.totalTasks) * 100}%` }"
                  ></div>
                </div>
                <span class="font-semibold">{{ count }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Task Type Distribution -->
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">Task Type Distribution</h3>
        <div class="chart-placeholder">
          <div class="space-y-2">
            <div v-for="(count, type) in analytics.taskMetrics.byType" :key="type" 
                 class="flex justify-between items-center">
              <span class="capitalize">{{ type }}</span>
              <div class="flex items-center gap-2">
                <div class="w-20 h-4 bg-gray-200 rounded overflow-hidden">
                  <div 
                    class="h-full bg-blue-500 rounded transition-all duration-500"
                    :style="{ width: `${(count / analytics.taskMetrics.totalTasks) * 100}%` }"
                  ></div>
                </div>
                <span class="font-semibold">{{ count }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Insights and Patterns -->
    <div class="card mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">Key Insights</h3>
      <div v-if="analytics.insights.length === 0" class="text-gray-600 text-center py-4">
        No insights available yet. Complete more tasks to see patterns!
      </div>
      <ul v-else class="space-y-2">
        <li v-for="insight in analytics.insights" :key="insight" 
            class="flex items-start gap-2">
          <span class="text-blue-500">💡</span>
          <span>{{ insight }}</span>
        </li>
      </ul>
    </div>

    <!-- Pattern Analysis -->
    <div class="grid grid-cols-1 lg:grid-cols-2">
      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">Common Tags</h3>
        <div class="flex flex-wrap gap-2">
          <span 
            v-for="tag in analytics.patternAnalysis.commonTags" 
            :key="tag"
            class="inline-block bg-blue-100 text-blue-800 text-sm px-3 py-1 rounded-full"
          >
            {{ tag }}
          </span>
        </div>
      </div>

      <div class="card">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">Most Productive Type</h3>
        <p class="text-xl font-semibold text-green-600 capitalize">
          {{ analytics.productivityMetrics.mostProductiveType }}
        </p>
        <p class="text-gray-600 text-sm mt-2">
          This task type has the highest completion rate and activity level.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'

// Reactive data
const selectedPeriod = ref('month')
const selectedType = ref('')
const loading = ref(false)

const analytics = reactive({
  taskMetrics: {
    totalTasks: 0,
    byStatus: {},
    byType: {},
    byPriority: {},
    completionRate: 0,
    averageEntries: 0,
    totalEntries: 0
  },
  productivityMetrics: {
    tasksCompletedPeriod: 0,
    entriesAddedPeriod: 0,
    averageTaskDuration: 0,
    mostProductiveType: '',
    productivityScore: 0
  },
  patternAnalysis: {
    mostFrequentType: '',
    commonTags: [],
    workPatterns: {},
    timeToCompletion: {}
  },
  insights: []
})

// Methods
const getStatusBarColor = (status) => {
  const colors = {
    active: 'bg-blue-500',
    completed: 'bg-green-500',
    paused: 'bg-yellow-500',
    blocked: 'bg-red-500'
  }
  return colors[status] || 'bg-gray-500'
}

const loadAnalytics = async () => {
  loading.value = true
  try {
    // TODO: Implement API call to get analytics
    // For now, using mock data
    await new Promise(resolve => setTimeout(resolve, 500))
    
    // Mock analytics data
    analytics.taskMetrics = {
      totalTasks: 15,
      byStatus: {
        active: 8,
        completed: 5,
        paused: 1,
        blocked: 1
      },
      byType: {
        work: 8,
        learning: 4,
        personal: 2,
        investigation: 1
      },
      byPriority: {
        high: 3,
        medium: 7,
        low: 5
      },
      completionRate: 0.33,
      averageEntries: 3.2,
      totalEntries: 48
    }
    
    analytics.productivityMetrics = {
      tasksCompletedPeriod: 5,
      entriesAddedPeriod: 12,
      averageTaskDuration: 4.5,
      mostProductiveType: 'work',
      productivityScore: 7.2
    }
    
    analytics.patternAnalysis = {
      mostFrequentType: 'work',
      commonTags: ['urgent', 'review', 'bug-fix', 'feature'],
      workPatterns: {},
      timeToCompletion: {}
    }
    
    analytics.insights = [
      'You complete 40% more work tasks than other types',
      'Your most productive day is Tuesday',
      'Tasks with priority "high" are completed 60% faster',
      'You tend to add more entries to learning tasks'
    ]
    
  } catch (error) {
    console.error('Failed to load analytics:', error)
  } finally {
    loading.value = false
  }
}

// Lifecycle
onMounted(() => {
  loadAnalytics()
})
</script>

<style scoped>
.form-select {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background-color: white;
}

.chart-placeholder {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.space-y-2 > * + * {
  margin-top: 0.5rem;
}

@media (min-width: 768px) {
  .grid-cols-1.md\\:grid-cols-4 {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (min-width: 1024px) {
  .grid-cols-1.lg\\:grid-cols-2 {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>