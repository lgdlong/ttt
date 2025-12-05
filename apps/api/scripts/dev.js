/**
 * Go dev wrapper for Turborepo integration
 * Handles GOCACHE path issues on Windows
 */
const { spawn } = require('child_process')
const path = require('path')
const os = require('os')

// Set GOCACHE to user's home directory if not defined
const gocache = process.env.GOCACHE || path.join(os.homedir(), '.cache', 'go-build')

console.log(`🚀 Starting Go API in development mode...`)
console.log(`   GOCACHE: ${gocache}`)

// Run go run with proper environment
const child = spawn('go', ['run', './cmd/api/main.go'], {
  stdio: 'inherit',
  cwd: path.join(__dirname, '..'),
  env: {
    ...process.env,
    GOCACHE: gocache,
  },
  shell: true,
})

child.on('error', (error) => {
  console.error('❌ Failed to start Go API:', error.message)
  process.exit(1)
})

child.on('close', (code) => {
  process.exit(code || 0)
})

// Handle graceful shutdown
process.on('SIGINT', () => {
  child.kill('SIGINT')
})

process.on('SIGTERM', () => {
  child.kill('SIGTERM')
})
