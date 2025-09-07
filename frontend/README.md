# Journal MCP Frontend

Modern Vue.js 3 frontend for the Journal MCP task management and journaling application.

## 🚀 Features

- **Dashboard**: Overview of tasks, recent activity, and quick actions
- **Task Management**: Create, edit, view, and organize tasks with filtering
- **Analytics**: Comprehensive insights with charts and productivity metrics
- **Settings**: Configuration for GitHub integration, backups, and preferences
- **Real-time Updates**: WebSocket integration for live data updates
- **Offline-First**: All dependencies bundled, no external CDN requirements
- **Responsive Design**: Works on desktop and mobile devices

## 🛠️ Technology Stack

- **Vue.js 3** with Composition API
- **Vue Router 4** for routing
- **Pinia** for state management
- **Vite** for fast development and building
- **Modern CSS** with responsive design
- **Chart.js** for analytics visualizations (planned)

## 📁 Project Structure

```
frontend/
├── src/
│   ├── components/         # Reusable Vue components
│   ├── pages/             # Page components (Dashboard, Tasks, etc.)
│   │   ├── Dashboard.vue  # Main dashboard
│   │   ├── Tasks.vue      # Task management
│   │   ├── Analytics.vue  # Analytics and insights
│   │   └── Settings.vue   # Configuration
│   ├── api/               # API client and services
│   ├── utils/             # Utility functions
│   ├── App.vue            # Main app component
│   ├── main.js            # Application entry point
│   └── style.css          # Global styles
├── assets/                # Static assets
├── dist/                  # Built frontend (embedded in Go binary)
├── index.html             # Main HTML file
├── package.json           # Dependencies and scripts
└── vite.config.js         # Vite configuration
```

## 🏃‍♂️ Development

### Prerequisites

- Node.js 16+ and npm
- Go 1.19+ (for backend)

### Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Start backend (in separate terminal)
cd ..
go run ./cmd/journal-mcp --web
```

The frontend will be available at http://localhost:5173 with API proxy to http://localhost:8080.

### Available Scripts

```bash
npm run dev      # Start development server with hot reload
npm run build    # Build for production
npm run preview  # Preview production build locally
```

## 🔧 Configuration

### API Proxy

The development server proxies API requests to the backend:

```javascript
// vite.config.js
server: {
  proxy: {
    '/api': 'http://localhost:8080'
  }
}
```

### Build Options

```javascript
// vite.config.js
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
}
```

## 📱 Pages Overview

### Dashboard
- Task summary statistics
- Recent activity timeline
- Quick action buttons
- Real-time updates

### Tasks
- Filterable task list
- Task creation and editing
- Status management
- Tag and priority support

### Analytics
- Task completion metrics
- Productivity insights
- Visual charts and graphs
- Time-based analysis

### Settings
- GitHub integration setup
- Backup configuration
- General preferences
- Data export/import

## 🎨 Styling

Uses modern CSS with:
- CSS Grid and Flexbox for layouts
- Custom properties for theming
- Responsive design patterns
- Minimal external dependencies

## 📦 Production Build

The frontend is built into a single bundle and embedded in the Go binary:

```bash
# Build frontend
npm run build

# Build complete application (from project root)
make build
```

The built files are embedded using Go's `embed` package and served by the web server.

## 🔌 API Integration

Uses a centralized API service (`src/api/index.js`) for all backend communication:

```javascript
import api from '@/api'

// Example usage
const tasks = await api.getTasks({ status: 'active' })
const analytics = await api.getAnalytics({ period: 'month' })
```

## 🚧 Development Status

- ✅ Basic project structure
- ✅ Core pages (Dashboard, Tasks, Analytics, Settings)
- ✅ Responsive design
- ✅ API service architecture
- ✅ Build system and embedding
- 🚧 Real API integration (using mock data)
- 🚧 WebSocket real-time updates
- 🚧 Chart.js integration
- 🚧 Advanced task management features

## 🤝 Contributing

When working on the frontend:

1. Follow Vue.js 3 Composition API patterns
2. Use the established project structure
3. Keep dependencies minimal and offline-capable
4. Test responsive design on multiple screen sizes
5. Ensure all features work with mock data before API integration