<template>
  <div class="app-layout">
    <!-- Sidebar Navigation -->
    <nav class="sidebar">
      <div class="mb-6">
        <h1 class="text-xl font-bold text-gray-800">Journal MCP</h1>
        <p class="text-gray-600 text-sm">Task Management & Journaling</p>
      </div>
      
      <ul class="space-y-2">
        <li>
          <router-link 
            to="/" 
            class="nav-link"
            :class="{ 'nav-link-active': $route.name === 'Dashboard' }"
          >
            📊 Dashboard
          </router-link>
        </li>
        <li>
          <router-link 
            to="/tasks" 
            class="nav-link"
            :class="{ 'nav-link-active': $route.name === 'Tasks' }"
          >
            ✅ Tasks
          </router-link>
        </li>
        <li>
          <router-link 
            to="/analytics" 
            class="nav-link"
            :class="{ 'nav-link-active': $route.name === 'Analytics' }"
          >
            📈 Analytics
          </router-link>
        </li>
        <li>
          <router-link 
            to="/settings" 
            class="nav-link"
            :class="{ 'nav-link-active': $route.name === 'Settings' }"
          >
            ⚙️ Settings
          </router-link>
        </li>
      </ul>
    </nav>

    <!-- Main Content -->
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup>
// Main app component - handles layout and navigation
</script>

<style scoped>
.nav-link {
  display: block;
  padding: 0.75rem 1rem;
  border-radius: 0.375rem;
  text-decoration: none;
  color: #4b5563;
  transition: all 0.2s;
}

.nav-link:hover {
  background-color: #f3f4f6;
  color: #1f2937;
}

.nav-link-active {
  background-color: #3b82f6;
  color: white;
}

.nav-link-active:hover {
  background-color: #2563eb;
}

.space-y-2 > * + * {
  margin-top: 0.5rem;
}
</style>