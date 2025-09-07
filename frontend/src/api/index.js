// API service for communicating with the Journal MCP backend
class ApiService {
  constructor(baseURL = '/api') {
    this.baseURL = baseURL
  }

  // Generic API request method
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`
    const config = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    }

    try {
      const response = await fetch(url, config)
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      
      const data = await response.json()
      return data
    } catch (error) {
      console.error('API request failed:', error)
      throw error
    }
  }

  // Task methods
  async createTask(taskData) {
    return this.request('/tasks', {
      method: 'POST',
      body: JSON.stringify(taskData),
    })
  }

  async getTasks(filters = {}) {
    const params = new URLSearchParams(filters)
    return this.request(`/tasks?${params}`)
  }

  async getTask(taskId) {
    return this.request(`/tasks/${taskId}`)
  }

  async updateTask(taskId, updates) {
    return this.request(`/tasks/${taskId}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    })
  }

  async deleteTask(taskId) {
    return this.request(`/tasks/${taskId}`, {
      method: 'DELETE',
    })
  }

  async addTaskEntry(taskId, entry) {
    return this.request(`/tasks/${taskId}/entries`, {
      method: 'POST',
      body: JSON.stringify(entry),
    })
  }

  // Analytics methods
  async getAnalytics(options = {}) {
    const params = new URLSearchParams(options)
    return this.request(`/analytics?${params}`)
  }

  async getDashboardStats() {
    return this.request('/dashboard')
  }

  // Configuration methods
  async getConfiguration() {
    return this.request('/config')
  }

  async updateConfiguration(config) {
    return this.request('/config', {
      method: 'PUT',
      body: JSON.stringify(config),
    })
  }

  // Backup methods
  async createBackup(options = {}) {
    return this.request('/backup', {
      method: 'POST',
      body: JSON.stringify(options),
    })
  }

  async restoreBackup(backupData) {
    return this.request('/backup/restore', {
      method: 'POST',
      body: JSON.stringify(backupData),
    })
  }

  // Export methods
  async exportData(format, filters = {}) {
    const params = new URLSearchParams({ format, ...filters })
    return this.request(`/export?${params}`)
  }

  // Import methods
  async importData(data, format) {
    return this.request('/import', {
      method: 'POST',
      body: JSON.stringify({ data, format }),
    })
  }

  // GitHub integration
  async syncWithGitHub(options) {
    return this.request('/github/sync', {
      method: 'POST',
      body: JSON.stringify(options),
    })
  }

  async pullIssueUpdates(options) {
    return this.request('/github/pull-updates', {
      method: 'POST',
      body: JSON.stringify(options),
    })
  }

  // Search
  async searchEntries(query, filters = {}) {
    const params = new URLSearchParams({ query, ...filters })
    return this.request(`/search?${params}`)
  }

  // One-on-ones
  async createOneOnOne(data) {
    return this.request('/one-on-ones', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async getOneOnOnes(limit = 10) {
    return this.request(`/one-on-ones?limit=${limit}`)
  }

  // Daily/Weekly logs
  async getDailyLog(date) {
    return this.request(`/logs/daily/${date}`)
  }

  async getWeeklyLog(weekStart) {
    return this.request(`/logs/weekly/${weekStart}`)
  }
}

// Create and export singleton instance
export const api = new ApiService()
export default api