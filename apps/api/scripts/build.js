/**
 * Go build wrapper for Turborepo integration
 * Handles GOCACHE path issues on Windows
 */
const { execSync } = require('child_process')
const path = require('path')
const fs = require('fs')
const os = require('os')

// Set GOCACHE to user's home directory if not defined
const gocache = process.env.GOCACHE || path.join(os.homedir(), '.cache', 'go-build')

// Ensure tmp directory exists
const tmpDir = path.join(__dirname, '..', 'tmp')
if (!fs.existsSync(tmpDir)) {
  fs.mkdirSync(tmpDir, { recursive: true })
}

// Determine output binary name based on OS
const outputBinary = process.platform === 'win32' ? 'tmp/main.exe' : 'tmp/main'

// Build command
const buildCmd = `go build -o ${outputBinary} ./cmd/api/main.go`

console.log(`📦 Building Go API...`)
console.log(`   GOCACHE: ${gocache}`)

try {
  execSync(buildCmd, {
    stdio: 'inherit',
    cwd: path.join(__dirname, '..'),
    env: {
      ...process.env,
      GOCACHE: gocache,
    },
  })
  console.log(`✅ Go build successful: ${outputBinary}`)
} catch (error) {
  console.error('❌ Go build failed')
  process.exit(1)
}
