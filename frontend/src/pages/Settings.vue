<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800 mb-6">Settings</h1>
    
    <!-- GitHub Integration -->
    <div class="card mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">🔗 GitHub Integration</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            GitHub Personal Access Token
          </label>
          <input 
            v-model="config.github.token"
            type="password"
            placeholder="ghp_xxxxxxxxxxxxxxxxxxxx"
            class="form-input w-full"
          />
          <p class="text-gray-600 text-sm mt-1">
            Used to sync GitHub issues with tasks. Generate at: 
            <a href="https://github.com/settings/tokens" target="_blank" class="text-blue-600 hover:underline">
              GitHub Settings → Tokens
            </a>
          </p>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            GitHub Username
          </label>
          <input 
            v-model="config.github.username"
            type="text"
            placeholder="your-username"
            class="form-input w-full"
          />
        </div>
        
        <div>
          <label class="flex items-center">
            <input 
              v-model="config.github.autoSync"
              type="checkbox"
              class="form-checkbox"
            />
            <span class="ml-2">Enable automatic sync</span>
          </label>
        </div>
      </div>
    </div>

    <!-- General Settings -->
    <div class="card mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">⚙️ General Settings</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Default Task Type
          </label>
          <select v-model="config.general.defaultTaskType" class="form-select w-full">
            <option value="work">Work</option>
            <option value="learning">Learning</option>
            <option value="personal">Personal</option>
            <option value="investigation">Investigation</option>
          </select>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Timezone
          </label>
          <select v-model="config.general.timeZone" class="form-select w-full">
            <option value="UTC">UTC</option>
            <option value="America/New_York">Eastern Time</option>
            <option value="America/Chicago">Central Time</option>
            <option value="America/Denver">Mountain Time</option>
            <option value="America/Los_Angeles">Pacific Time</option>
            <option value="Europe/London">London</option>
            <option value="Europe/Paris">Paris</option>
            <option value="Asia/Tokyo">Tokyo</option>
          </select>
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Date Format
          </label>
          <select v-model="config.general.dateFormat" class="form-select w-full">
            <option value="YYYY-MM-DD">2024-01-15 (ISO)</option>
            <option value="MM/DD/YYYY">01/15/2024 (US)</option>
            <option value="DD/MM/YYYY">15/01/2024 (EU)</option>
            <option value="MMM DD, YYYY">Jan 15, 2024</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Backup Settings -->
    <div class="card mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">💾 Backup & Data</h3>
      <div class="space-y-4">
        <div>
          <label class="flex items-center">
            <input 
              v-model="config.backup.autoBackup"
              type="checkbox"
              class="form-checkbox"
            />
            <span class="ml-2">Enable automatic backups</span>
          </label>
        </div>
        
        <div v-if="config.backup.autoBackup">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Backup Interval (hours)
          </label>
          <input 
            v-model.number="config.backup.backupInterval"
            type="number"
            min="1"
            max="168"
            class="form-input w-32"
          />
        </div>
        
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Maximum Backups to Keep
          </label>
          <input 
            v-model.number="config.backup.maxBackups"
            type="number"
            min="1"
            max="50"
            class="form-input w-32"
          />
        </div>
        
        <div class="flex gap-3">
          <button @click="createBackup" class="btn btn-secondary" :disabled="backupLoading">
            {{ backupLoading ? '⏳ Creating...' : '📁 Create Backup Now' }}
          </button>
          <button @click="showRestoreModal = true" class="btn btn-secondary">
            📂 Restore Backup
          </button>
        </div>
      </div>
    </div>

    <!-- Export/Import -->
    <div class="card mb-6">
      <h3 class="text-lg font-semibold text-gray-800 mb-4">📤 Export & Import</h3>
      <div class="space-y-4">
        <div>
          <h4 class="font-medium text-gray-800 mb-2">Export Data</h4>
          <div class="flex gap-3">
            <button @click="exportData('json')" class="btn btn-secondary">
              Export as JSON
            </button>
            <button @click="exportData('markdown')" class="btn btn-secondary">
              Export as Markdown
            </button>
            <button @click="exportData('csv')" class="btn btn-secondary">
              Export as CSV
            </button>
          </div>
        </div>
        
        <div>
          <h4 class="font-medium text-gray-800 mb-2">Import Data</h4>
          <input 
            ref="fileInput"
            type="file"
            accept=".json,.md,.csv,.txt"
            @change="handleFileSelect"
            class="form-input w-full"
          />
          <p class="text-gray-600 text-sm mt-1">
            Supports JSON, Markdown, CSV, and text files
          </p>
        </div>
      </div>
    </div>

    <!-- Save Button -->
    <div class="flex gap-3">
      <button @click="saveSettings" class="btn btn-primary" :disabled="saving">
        {{ saving ? '⏳ Saving...' : '💾 Save Settings' }}
      </button>
      <button @click="resetSettings" class="btn btn-secondary">
        🔄 Reset to Defaults
      </button>
    </div>

    <!-- Status Messages -->
    <div v-if="statusMessage" class="mt-4 p-4 rounded" :class="statusType === 'success' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'">
      {{ statusMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'

// Reactive data
const config = reactive({
  github: {
    token: '',
    username: '',
    repositories: [],
    autoSync: false,
    syncInterval: 60
  },
  web: {
    enabled: true,
    port: 8080
  },
  backup: {
    autoBackup: false,
    backupInterval: 24,
    backupLocation: '',
    maxBackups: 10
  },
  general: {
    defaultTaskType: 'work',
    timeZone: 'UTC',
    dateFormat: 'YYYY-MM-DD'
  }
})

const saving = ref(false)
const backupLoading = ref(false)
const showRestoreModal = ref(false)
const statusMessage = ref('')
const statusType = ref('success')
const fileInput = ref(null)

// Methods
const loadSettings = async () => {
  try {
    // TODO: Implement API call to load settings
    // For now, loading defaults
    console.log('Loading settings...')
  } catch (error) {
    console.error('Failed to load settings:', error)
    showStatus('Failed to load settings', 'error')
  }
}

const saveSettings = async () => {
  saving.value = true
  try {
    // TODO: Implement API call to save settings
    await new Promise(resolve => setTimeout(resolve, 1000)) // Simulate API call
    
    console.log('Saving settings:', config)
    showStatus('Settings saved successfully!', 'success')
  } catch (error) {
    console.error('Failed to save settings:', error)
    showStatus('Failed to save settings', 'error')
  } finally {
    saving.value = false
  }
}

const resetSettings = () => {
  if (confirm('Are you sure you want to reset all settings to defaults?')) {
    // Reset to defaults
    Object.assign(config, {
      github: {
        token: '',
        username: '',
        repositories: [],
        autoSync: false,
        syncInterval: 60
      },
      web: {
        enabled: true,
        port: 8080
      },
      backup: {
        autoBackup: false,
        backupInterval: 24,
        backupLocation: '',
        maxBackups: 10
      },
      general: {
        defaultTaskType: 'work',
        timeZone: 'UTC',
        dateFormat: 'YYYY-MM-DD'
      }
    })
    showStatus('Settings reset to defaults', 'success')
  }
}

const createBackup = async () => {
  backupLoading.value = true
  try {
    // TODO: Implement backup API call
    await new Promise(resolve => setTimeout(resolve, 2000))
    showStatus('Backup created successfully!', 'success')
  } catch (error) {
    console.error('Backup failed:', error)
    showStatus('Backup failed', 'error')
  } finally {
    backupLoading.value = false
  }
}

const exportData = async (format) => {
  try {
    // TODO: Implement export API call
    console.log(`Exporting data as ${format}`)
    showStatus(`Data exported as ${format.toUpperCase()}`, 'success')
  } catch (error) {
    console.error('Export failed:', error)
    showStatus('Export failed', 'error')
  }
}

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    // TODO: Implement file import
    console.log('Selected file for import:', file.name)
    showStatus(`Ready to import: ${file.name}`, 'success')
  }
}

const showStatus = (message, type = 'success') => {
  statusMessage.value = message
  statusType.value = type
  setTimeout(() => {
    statusMessage.value = ''
  }, 3000)
}

// Lifecycle
onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.form-input {
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background-color: white;
  transition: border-color 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-select {
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  background-color: white;
}

.form-checkbox {
  width: 1rem;
  height: 1rem;
  color: #3b82f6;
  border-radius: 0.25rem;
}

.space-y-4 > * + * {
  margin-top: 1rem;
}
</style>