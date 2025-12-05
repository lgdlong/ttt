/**
 * Go test wrapper for Turborepo integration
 * Handles GOCACHE path issues on Windows
 */
const { spawn } = require('child_process')
const path = require('path')
const os = require('os')

// Set GOCACHE to user's home directory if not defined
const gocache = process.env.GOCACHE || path.join(os.homedir(), '.cache', 'go-build')

console.log(`🧪 Running Go tests...`)
console.log(`   GOCACHE: ${gocache}`)

// Run go test with proper environment
const child = spawn('go', ['test', '-v', './...'], {
  stdio: 'inherit',
  cwd: path.join(__dirname, '..'),
  env: {
    ...process.env,
    GOCACHE: gocache,
  },
  shell: true,
})

child.on('error', (error) => {
  console.error('❌ Failed to run tests:', error.message)
  process.exit(1)
})

child.on('close', (code) => {
  if (code === 0) {
    console.log('✅ All tests passed!')
  } else {
    console.error('❌ Some tests failed')
  }
  process.exit(code || 0)
})
