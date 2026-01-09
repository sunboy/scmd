#!/usr/bin/env node

/**
 * npm post-install script for scmd
 * Downloads the appropriate binary for the current platform
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

// Configuration
const REPO = 'sunboy/scmd';
const GITHUB_API = 'https://api.github.com';
const GITHUB_RELEASES = 'https://github.com';

// Platform mapping
const PLATFORM_MAPPING = {
  darwin: 'macOS',
  linux: 'linux',
  win32: 'windows'
};

const ARCH_MAPPING = {
  x64: 'amd64',
  arm64: 'arm64'
};

// Colors for console output
const colors = {
  reset: '\x1b[0m',
  blue: '\x1b[34m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m'
};

function log(message, color = 'blue') {
  console.log(`${colors[color]}==>${colors.reset} ${message}`);
}

function error(message) {
  console.error(`${colors.red}Error:${colors.reset} ${message}`);
  process.exit(1);
}

function getPlatform() {
  const platform = PLATFORM_MAPPING[process.platform];
  if (!platform) {
    error(`Unsupported platform: ${process.platform}`);
  }
  return platform;
}

function getArch() {
  const arch = ARCH_MAPPING[process.arch];
  if (!arch) {
    error(`Unsupported architecture: ${process.arch}`);
  }
  return arch;
}

function getPackageVersion() {
  const packageJson = require('./package.json');
  return packageJson.version;
}

async function getLatestVersion() {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${REPO}/releases/latest`,
      method: 'GET',
      headers: {
        'User-Agent': 'scmd-npm-installer'
      }
    };

    https.get(options, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        try {
          const release = JSON.parse(data);
          resolve(release.tag_name);
        } catch (e) {
          reject(e);
        }
      });
    }).on('error', reject);
  });
}

async function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    const request = (url) => {
      https.get(url, { headers: { 'User-Agent': 'scmd-npm-installer' } }, (res) => {
        // Handle redirects
        if (res.statusCode === 302 || res.statusCode === 301) {
          request(res.headers.location);
          return;
        }

        if (res.statusCode !== 200) {
          reject(new Error(`Failed to download: ${res.statusCode}`));
          return;
        }

        const totalSize = parseInt(res.headers['content-length'], 10);
        let downloadedSize = 0;
        let lastPercent = 0;

        res.on('data', (chunk) => {
          downloadedSize += chunk.length;
          const percent = Math.floor((downloadedSize / totalSize) * 100);

          if (percent !== lastPercent && percent % 10 === 0) {
            process.stdout.write(`\r  Downloading... ${percent}%`);
            lastPercent = percent;
          }
        });

        res.pipe(file);
        file.on('finish', () => {
          file.close();
          process.stdout.write('\r  Downloading... 100%\n');
          resolve();
        });
      }).on('error', (err) => {
        fs.unlink(dest, () => {});
        reject(err);
      });
    };

    request(url);
  });
}

async function verifyChecksum(archivePath, checksumUrl) {
  const crypto = require('crypto');

  log('Verifying checksum...');

  // Download checksums file
  const checksumPath = path.join(path.dirname(archivePath), 'checksums.txt');
  await downloadFile(checksumUrl, checksumPath);

  // Read checksums file
  const checksums = fs.readFileSync(checksumPath, 'utf8');
  const filename = path.basename(archivePath);
  const checksumLine = checksums.split('\n').find(line => line.includes(filename));

  if (!checksumLine) {
    error(`Checksum not found for ${filename}`);
  }

  const expectedChecksum = checksumLine.split(/\s+/)[0];

  // Calculate actual checksum
  const fileBuffer = fs.readFileSync(archivePath);
  const actualChecksum = crypto.createHash('sha256').update(fileBuffer).digest('hex');

  if (expectedChecksum !== actualChecksum) {
    error('Checksum verification failed');
  }

  log('Checksum verified', 'green');
  fs.unlinkSync(checksumPath);
}

async function extractArchive(archivePath, destDir) {
  log('Extracting...');

  const platform = process.platform;
  const isZip = archivePath.endsWith('.zip');

  if (isZip) {
    // Use unzip for Windows
    if (platform === 'win32') {
      execSync(`powershell -command "Expand-Archive -Path '${archivePath}' -DestinationPath '${destDir}' -Force"`);
    } else {
      execSync(`unzip -q "${archivePath}" -d "${destDir}"`);
    }
  } else {
    // Use tar for Unix-like systems
    execSync(`tar -xzf "${archivePath}" -C "${destDir}"`);
  }
}

async function main() {
  try {
    log('Installing scmd...');

    const platform = getPlatform();
    const arch = getArch();
    const version = getPackageVersion();

    log(`Platform: ${platform} ${arch}`);
    log(`Version: ${version}`);

    // Construct URLs
    const archiveName = platform === 'windows'
      ? `scmd_${version}_${platform}_${arch}.zip`
      : `scmd_${version}_${platform}_${arch}.tar.gz`;

    const baseUrl = `${GITHUB_RELEASES}/${REPO}/releases/download/v${version}`;
    const archiveUrl = `${baseUrl}/${archiveName}`;
    const checksumUrl = `${baseUrl}/checksums.txt`;

    // Create directories
    const binDir = path.join(__dirname, 'bin');
    const tmpDir = path.join(__dirname, 'tmp');
    fs.mkdirSync(binDir, { recursive: true });
    fs.mkdirSync(tmpDir, { recursive: true });

    const archivePath = path.join(tmpDir, archiveName);

    // Download binary
    log(`Downloading from ${archiveUrl}...`);
    await downloadFile(archiveUrl, archivePath);

    // Verify checksum
    await verifyChecksum(archivePath, checksumUrl);

    // Extract archive
    await extractArchive(archivePath, tmpDir);

    // Move binary to bin directory
    const binaryName = platform === 'windows' ? 'scmd.exe' : 'scmd';
    const sourceBinary = path.join(tmpDir, binaryName);
    const destBinary = path.join(binDir, binaryName);

    if (fs.existsSync(destBinary)) {
      fs.unlinkSync(destBinary);
    }

    fs.renameSync(sourceBinary, destBinary);

    // Make executable (Unix-like systems)
    if (platform !== 'windows') {
      fs.chmodSync(destBinary, 0o755);
    }

    // Clean up
    fs.rmSync(tmpDir, { recursive: true, force: true });

    log('scmd installed successfully!', 'green');
    console.log('');
    log('Next steps:', 'blue');
    console.log('  1. Verify installation: scmd --version');
    console.log('  2. Install llama.cpp for offline usage: brew install llama.cpp');
    console.log('  3. Try it out: scmd /explain "what is a goroutine"');
    console.log('');

  } catch (err) {
    error(`Installation failed: ${err.message}`);
  }
}

// Only run if this is the main module
if (require.main === module) {
  main();
}

module.exports = { main };
