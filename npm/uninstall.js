#!/usr/bin/env node

/**
 * npm pre-uninstall script for scmd
 * Cleans up the binary
 */

const fs = require('fs');
const path = require('path');

const colors = {
  reset: '\x1b[0m',
  blue: '\x1b[34m',
  yellow: '\x1b[33m'
};

function log(message, color = 'blue') {
  console.log(`${colors[color]}==>${colors.reset} ${message}`);
}

function main() {
  log('Uninstalling scmd...');

  const binDir = path.join(__dirname, 'bin');

  if (fs.existsSync(binDir)) {
    fs.rmSync(binDir, { recursive: true, force: true });
    log('Binary removed');
  }

  console.log('');
  log('Note: User data in ~/.scmd has been preserved.', 'yellow');
  log('To completely remove scmd data, run:', 'yellow');
  console.log('  rm -rf ~/.scmd');
  console.log('');
}

if (require.main === module) {
  main();
}

module.exports = { main };
