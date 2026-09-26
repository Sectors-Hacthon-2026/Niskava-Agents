#!/usr/bin/env node

/**
 * NISKAVA AGENT — NPM Universal CLI Launcher
 * Bridges Node.js npx/npm execution with native Go Core and Python Quant Engine.
 * Supports auto-downloading precompiled binaries from GitHub Releases.
 */

const { spawn, spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');
const https = require('https');

const ROOT_DIR = path.resolve(__dirname, '..');
const BIN_DIR = __dirname;
const PKG_VERSION = require('../package.json').version;
const GITHUB_REPO = 'Sectors-Hacthon-2026/Niskava-Agents';

const isWindows = process.platform === 'win32';
const isDarwin = process.platform === 'darwin';
const isLinux = process.platform === 'linux';
const arch = process.arch;

function getBinaryName() {
    if (isWindows) return 'niskava-windows-amd64.exe';
    if (isDarwin) {
        return arch === 'arm64' ? 'niskava-darwin-arm64' : 'niskava-darwin-amd64';
    }
    if (isLinux) {
        return arch === 'arm64' ? 'niskava-linux-arm64' : 'niskava-linux-amd64';
    }
    return isWindows ? 'niskava.exe' : 'niskava';
}

const targetBinaryName = isWindows ? 'niskava.exe' : 'niskava';
const localTarget = path.join(BIN_DIR, targetBinaryName);

function checkPythonRuntime() {
    const candidates = isWindows ? ['python', 'py'] : ['python3', 'python'];
    for (const cmd of candidates) {
        try {
            const check = spawnSync(cmd, ['-c', "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')"], { encoding: 'utf8' });
            if (check.status === 0 && check.stdout) {
                const parts = check.stdout.trim().split('.');
                const major = parseInt(parts[0], 10);
                const minor = parseInt(parts[1], 10);
                if (major >= 3 && minor >= 11) {
                    return cmd;
                }
            }
        } catch (_) {}
    }
    return null;
}

function downloadBinary(assetName, destPath) {
    return new Promise((resolve, reject) => {
        const url = `https://github.com/${GITHUB_REPO}/releases/download/v${PKG_VERSION}/${assetName}`;
        console.log(`\x1b[36m[↓] Downloading Niskava native binary from GitHub Releases:\x1b[0m\n    ${url}`);

        function followRedirect(currentUrl, redirectCount = 0) {
            if (redirectCount > 5) {
                return reject(new Error('Too many redirects while downloading binary.'));
            }

            https.get(currentUrl, (res) => {
                if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
                    return followRedirect(res.headers.location, redirectCount + 1);
                }

                if (res.statusCode !== 200) {
                    return reject(new Error(`HTTP Download failed with status ${res.statusCode}`));
                }

                const fileStream = fs.createWriteStream(destPath);
                res.pipe(fileStream);

                fileStream.on('finish', () => {
                    fileStream.close(() => {
                        if (!isWindows) {
                            fs.chmodSync(destPath, 0o755);
                        }
                        console.log('\x1b[32m[✓] Download and installation complete.\x1b[0m\n');
                        resolve();
                    });
                });

                fileStream.on('error', (err) => {
                    fs.unlink(destPath, () => {});
                    reject(err);
                });
            }).on('error', reject);
        }

        followRedirect(url);
    });
}

async function ensureBinary() {
    if (fs.existsSync(localTarget)) {
        try {
            if (!isWindows) fs.chmodSync(localTarget, 0o755);
            return localTarget;
        } catch (_) {
            return localTarget;
        }
    }

    // Attempt 1: Download precompiled binary from GitHub Releases
    const releaseAsset = getBinaryName();
    try {
        await downloadBinary(releaseAsset, localTarget);
        if (fs.existsSync(localTarget)) {
            return localTarget;
        }
    } catch (err) {
        console.warn(`\x1b[33m[!] GitHub prebuilt binary download failed: ${err.message}\x1b[0m`);
    }

    // Attempt 2: Fallback to local Go compilation if Go compiler is present in host
    console.log('\x1b[33m[!] Checking Go compiler on host...\x1b[0m');
    const goCheck = spawnSync('go', ['version'], { encoding: 'utf8' });
    if (goCheck.status === 0 && fs.existsSync(path.join(ROOT_DIR, 'go.mod'))) {
        console.log('\x1b[32m[✓] Go compiler found. Compiling native executable...\x1b[0m');
        const build = spawnSync('go', ['build', '-ldflags', '-s -w', '-o', localTarget, './cmd/niskava'], {
            cwd: ROOT_DIR,
            stdio: 'inherit'
        });
        if (build.status === 0 && fs.existsSync(localTarget)) {
            if (!isWindows) fs.chmodSync(localTarget, 0o755);
            return localTarget;
        }
    }

    console.error('\x1b[31m[✗] Error: Niskava native executable could not be acquired.\x1b[0m');
    console.error('    Please install Go (https://go.dev) or download the binary manually from:');
    console.error(`    https://github.com/${GITHUB_REPO}/releases/tag/v${PKG_VERSION}\n`);
    process.exit(1);
}

async function main() {
    const executable = await ensureBinary();

    // Warn if Python 3.11+ is missing (for quantitative calculations)
    const pythonCmd = checkPythonRuntime();
    if (!pythonCmd && !process.env.NISKAVA_PYTHON_PATH) {
        console.warn('\x1b[33m[!] Note: Python 3.11+ was not detected on PATH.\x1b[0m');
        console.warn('    Deep quantitative calculations and anomaly recon require Python 3.11+.\x1b[0m\n');
    }

    const args = process.argv.slice(2);
    const child = spawn(executable, args, {
        cwd: process.cwd(),
        stdio: 'inherit',
        env: {
            ...process.env,
            NISKAVA_ROOT: ROOT_DIR
        }
    });

    child.on('error', (err) => {
        console.error(`\x1b[31mFailed to start Niskava process: ${err.message}\x1b[0m`);
        process.exit(1);
    });

    child.on('close', (code) => {
        process.exit(code || 0);
    });

    process.on('SIGINT', () => {
        if (child && !child.killed) child.kill('SIGINT');
    });
    process.on('SIGTERM', () => {
        if (child && !child.killed) child.kill('SIGTERM');
    });
}

main().catch((err) => {
    console.error(`\x1b[31mUnexpected launcher error: ${err.message}\x1b[0m`);
    process.exit(1);
});
