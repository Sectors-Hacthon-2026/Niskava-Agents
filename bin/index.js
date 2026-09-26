#!/usr/bin/env node

/**
 * NISKAVA AGENT — NPM Universal CLI Launcher (User-Space Resilient)
 * Bridges Node.js npx/npm execution with native Go Core and Python Quant Engine.
 * 
 * Invariants:
 * 1. ZERO write operations inside node_modules (prevents EACCES permission denied).
 * 2. All downloads and runtime artifacts reside strictly in user home (~/.niskava/bin).
 * 3. Supports HTTP 302/301 redirects with User-Agent header for GitHub Release assets.
 * 4. Automatic fallback to local Go compilation if Go compiler is present in host.
 * 5. Full terminal stdio and POSIX signal forwarding (SIGINT/SIGTERM).
 */

const { spawn, spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');
const https = require('https');
const { getTargetBinaryPath, getPlatformAssetName, getNiskavaHome } = require('./resolver');

const ROOT_DIR = path.resolve(__dirname, '..');
const PKG_VERSION = require('../package.json').version;
const GITHUB_REPO = 'Sectors-Hacthon-2026/Niskava-Agents';

const isWindows = process.platform === 'win32';

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
        console.log(`\x1b[36m[↓] Downloading Niskava native binary for ${process.platform}/${process.arch}:\x1b[0m\n    ${url}`);

        // Ensure destination directory exists
        const destDir = path.dirname(destPath);
        if (!fs.existsSync(destDir)) {
            fs.mkdirSync(destDir, { recursive: true });
        }

        const tempPath = `${destPath}.tmp.${Date.now()}`;

        function followRedirect(currentUrl, redirectCount = 0) {
            if (redirectCount > 5) {
                return reject(new Error('Too many redirects while downloading binary.'));
            }

            const reqOptions = {
                headers: {
                    'User-Agent': 'niskava-npm-launcher',
                    'Accept': 'application/octet-stream'
                }
            };

            https.get(currentUrl, reqOptions, (res) => {
                if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
                    return followRedirect(res.headers.location, redirectCount + 1);
                }

                if (res.statusCode !== 200) {
                    return reject(new Error(`HTTP Download failed with status ${res.statusCode}`));
                }

                const fileStream = fs.createWriteStream(tempPath);
                res.pipe(fileStream);

                fileStream.on('finish', () => {
                    fileStream.close(() => {
                        try {
                            if (!isWindows) {
                                fs.chmodSync(tempPath, 0o755);
                            }
                            // Atomic rename
                            fs.renameSync(tempPath, destPath);
                            console.log('\x1b[32m[✓] Binary downloaded and cached successfully in ~/.niskava/bin\x1b[0m\n');
                            resolve(destPath);
                        } catch (renameErr) {
                            reject(renameErr);
                        }
                    });
                });

                fileStream.on('error', (err) => {
                    fs.unlink(tempPath, () => {});
                    reject(err);
                });
            }).on('error', (err) => {
                fs.unlink(tempPath, () => {});
                reject(err);
            });
        }

        followRedirect(url);
    });
}

async function ensureBinary() {
    const targetBinary = getTargetBinaryPath(PKG_VERSION);

    // 1. Check if binary is already cached in user home (~/.niskava/bin/)
    if (fs.existsSync(targetBinary)) {
        try {
            if (!isWindows) fs.chmodSync(targetBinary, 0o755);
            return targetBinary;
        } catch (_) {
            return targetBinary;
        }
    }

    // 2. Check if a pre-compiled binary exists in local repo workspace (for dev mode)
    const localRepoBinary = path.join(ROOT_DIR, 'bin', isWindows ? 'niskava.exe' : 'niskava');
    if (fs.existsSync(localRepoBinary)) {
        try {
            if (!isWindows) fs.chmodSync(localRepoBinary, 0o755);
            return localRepoBinary;
        } catch (_) {
            return localRepoBinary;
        }
    }

    // 3. Attempt to download precompiled platform binary from GitHub Releases
    const releaseAsset = getPlatformAssetName();
    try {
        await downloadBinary(releaseAsset, targetBinary);
        if (fs.existsSync(targetBinary)) {
            return targetBinary;
        }
    } catch (err) {
        console.warn(`\x1b[33m[!] GitHub prebuilt binary download failed: ${err.message}\x1b[0m`);
    }

    // 4. Fallback: Local compilation if Go compiler exists and source files are present
    const goModPath = path.join(ROOT_DIR, 'go.mod');
    if (fs.existsSync(goModPath)) {
        console.log('\x1b[33m[!] Checking Go compiler on host...\x1b[0m');
        const goCheck = spawnSync('go', ['version'], { encoding: 'utf8' });
        if (goCheck.status === 0) {
            console.log('\x1b[32m[✓] Go compiler found. Compiling native executable...\x1b[0m');
            const destDir = path.dirname(targetBinary);
            if (!fs.existsSync(destDir)) fs.mkdirSync(destDir, { recursive: true });

            const build = spawnSync('go', ['build', '-ldflags', '-s -w', '-o', targetBinary, './cmd/niskava'], {
                cwd: ROOT_DIR,
                stdio: 'inherit'
            });
            if (build.status === 0 && fs.existsSync(targetBinary)) {
                if (!isWindows) fs.chmodSync(targetBinary, 0o755);
                return targetBinary;
            }
        }
    }

    console.error('\x1b[31m[✗] Error: Niskava native executable could not be acquired.\x1b[0m');
    console.error('    Please download the binary manually from:');
    console.error(`    https://github.com/${GITHUB_REPO}/releases/tag/v${PKG_VERSION}\n`);
    console.error(`    and place it into: ${targetBinary}\n`);
    process.exit(1);
}

async function main() {
    const executable = await ensureBinary();

    // Guidance if Python 3.11+ is missing (for quantitative calculations)
    const pythonCmd = checkPythonRuntime();
    if (!pythonCmd && !process.env.NISKAVA_PYTHON_PATH) {
        console.warn('\x1b[33m[!] Note: Python 3.11+ was not detected on PATH.\x1b[0m');
        console.warn('    Deep quantitative calculations and anomaly recon require Python 3.11+.\x1b[0m');
        if (process.platform === 'linux') {
            console.warn('    → To install on Ubuntu/Debian: sudo apt install python3 python3-venv python3-pip\x1b[0m\n');
        } else if (process.platform === 'darwin') {
            console.warn('    → To install on macOS: brew install python@3.12\x1b[0m\n');
        } else if (process.platform === 'win32') {
            console.warn('    → To install on Windows: winget install Python.Python.3.12\x1b[0m\n');
        }
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
