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
const os = require('os');
const { getTargetBinaryPath, getPlatformAssetName, getNiskavaHome, getGoBinBinaryPath } = require('./resolver');

const ROOT_DIR = path.resolve(__dirname, '..');
const PKG_VERSION = require('../package.json').version;
const GITHUB_REPO = 'Sectors-Hacthon-2026/Niskava-Agents';

const isWindows = process.platform === 'win32';

function copyDirRecursiveSync(srcDir, destDir) {
    if (!fs.existsSync(srcDir)) return;
    if (!fs.existsSync(destDir)) {
        fs.mkdirSync(destDir, { recursive: true });
    }
    const entries = fs.readdirSync(srcDir, { withFileTypes: true });
    for (const entry of entries) {
        const srcPath = path.join(srcDir, entry.name);
        const destPath = path.join(destDir, entry.name);

        if (entry.name === '__pycache__' || entry.name === '.pytest_cache' || entry.name.endsWith('.pyc')) {
            continue;
        }

        if (entry.isDirectory()) {
            copyDirRecursiveSync(srcPath, destPath);
        } else if (entry.isFile()) {
            try {
                fs.copyFileSync(srcPath, destPath);
            } catch (_) {}
        }
    }
}

function syncEngineToUserSpace(rootDir, niskavaHome = getNiskavaHome()) {
    const srcEngine = path.join(rootDir, 'backend', 'engine');
    if (!fs.existsSync(srcEngine)) return false;

    try {
        const targetEngine = path.join(niskavaHome, 'engine');
        const targetBackendEngine = path.join(niskavaHome, 'backend', 'engine');

        copyDirRecursiveSync(srcEngine, targetEngine);
        copyDirRecursiveSync(srcEngine, targetBackendEngine);
        return true;
    } catch (_) {
        return false;
    }
}

function ensureUserVenv(pythonCmd, niskavaHome = getNiskavaHome(), rootDir = ROOT_DIR) {
    if (!pythonCmd) return null;

    const venvDir = path.join(niskavaHome, 'venv');
    const venvPy = isWindows 
        ? path.join(venvDir, 'Scripts', 'python.exe')
        : path.join(venvDir, 'bin', 'python3');

    if (fs.existsSync(venvPy)) {
        return venvPy;
    }

    try {
        if (!fs.existsSync(venvDir)) {
            const venvRes = spawnSync(pythonCmd, ['-m', 'venv', venvDir], {
                timeout: 30000,
                stdio: 'ignore'
            });
            if (venvRes.status !== 0 || !fs.existsSync(venvPy)) {
                return pythonCmd;
            }
        }

        let reqPath = path.join(rootDir, 'backend', 'engine', 'requirements.txt');
        if (!fs.existsSync(reqPath)) {
            reqPath = path.join(niskavaHome, 'engine', 'requirements.txt');
        }
        if (fs.existsSync(reqPath)) {
            spawnSync(venvPy, ['-m', 'pip', 'install', '-r', reqPath, '--quiet'], {
                timeout: 60000,
                stdio: 'ignore'
            });
        }
        return venvPy;
    } catch (_) {
        return pythonCmd;
    }
}

function checkPythonRuntime(niskavaHome = getNiskavaHome(), rootDir = ROOT_DIR) {
    // 1. Check user-space ~/.niskava/venv first before system PATH
    const venvDir = path.join(niskavaHome, 'venv');
    const venvPy = isWindows 
        ? path.join(venvDir, 'Scripts', 'python.exe')
        : path.join(venvDir, 'bin', 'python3');
    if (fs.existsSync(venvPy)) {
        try {
            const check = spawnSync(venvPy, ['-c', "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')"], { encoding: 'utf8', timeout: 5000 });
            if (check.status === 0) {
                return venvPy;
            }
        } catch (_) {}
    }

    const candidates = isWindows ? ['python', 'py'] : ['python3', 'python'];
    for (const cmd of candidates) {
        try {
            const check = spawnSync(cmd, ['-c', "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')"], { encoding: 'utf8', timeout: 5000 });
            if (check.status === 0 && check.stdout) {
                const parts = check.stdout.trim().split('.');
                const major = parseInt(parts[0], 10);
                const minor = parseInt(parts[1], 10);
                if (major >= 3 && minor >= 11) {
                    const bootstrapped = ensureUserVenv(cmd, niskavaHome, rootDir);
                    return bootstrapped || cmd;
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
                            const stat = fs.statSync(tempPath);
                            if (stat.size < 1024 * 1024) {
                                try { fs.unlinkSync(tempPath); } catch (_) {}
                                return reject(new Error(`Downloaded binary is incomplete or truncated (${stat.size} bytes). Minimum size is 1MB.`));
                            }
                            if (!isWindows) {
                                fs.chmodSync(tempPath, 0o755);
                            }
                            // Atomic rename
                            fs.renameSync(tempPath, destPath);
                            console.log('\x1b[32m[✓] Binary downloaded and cached successfully in ~/.niskava/bin\x1b[0m\n');
                            resolve(destPath);
                        } catch (renameErr) {
                            try { fs.unlinkSync(tempPath); } catch (_) {}
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
    const goBinBinary = getGoBinBinaryPath();
    const localRepoBinary = path.join(ROOT_DIR, 'bin', isWindows ? 'niskava.exe' : 'niskava');
    const localRootBinary = path.join(ROOT_DIR, isWindows ? 'niskava.exe' : 'niskava');

    // 1. Check if a newly compiled binary exists from `go install ./cmd/niskava` or local `go build`
    const localCandidates = [goBinBinary, localRepoBinary, localRootBinary];
    let newestLocal = null;
    let newestMtime = 0;

    for (const cand of localCandidates) {
        if (fs.existsSync(cand)) {
            try {
                const stat = fs.statSync(cand);
                if (stat.size >= 1024 * 1024 && stat.mtimeMs > newestMtime) {
                    newestMtime = stat.mtimeMs;
                    newestLocal = cand;
                }
            } catch (_) {}
        }
    }

    if (newestLocal) {
        let useNewestLocal = true;
        if (fs.existsSync(targetBinary)) {
            try {
                const targetStat = fs.statSync(targetBinary);
                if (targetStat.mtimeMs >= newestMtime) {
                    useNewestLocal = false;
                }
            } catch (_) {}
        }

        if (useNewestLocal) {
            try {
                const destDir = path.dirname(targetBinary);
                if (!fs.existsSync(destDir)) fs.mkdirSync(destDir, { recursive: true });
                fs.copyFileSync(newestLocal, targetBinary);
                if (!isWindows) fs.chmodSync(targetBinary, 0o755);
            } catch (_) {}
            if (!isWindows) fs.chmodSync(newestLocal, 0o755);
            return newestLocal;
        }
    }

    // 2. Check if binary is already cached in user home (~/.niskava/bin/)
    if (fs.existsSync(targetBinary)) {
        try {
            const stat = fs.statSync(targetBinary);
            if (stat.size >= 1024 * 1024) {
                if (!isWindows) fs.chmodSync(targetBinary, 0o755);
                return targetBinary;
            }
            console.warn(`\x1b[33m[!] Cached binary in ~/.niskava/bin is truncated (${stat.size} bytes). Re-downloading...\x1b[0m`);
            try { fs.unlinkSync(targetBinary); } catch (_) {}
        } catch (_) {
            try { fs.unlinkSync(targetBinary); } catch (_) {}
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
    const userHome = getNiskavaHome();

    // 1. Sync Python quant engine to user space ~/.niskava/engine
    syncEngineToUserSpace(ROOT_DIR, userHome);

    // 2. Resolve Python runtime (prioritizing isolated user-space venv)
    const pythonCmd = checkPythonRuntime(userHome, ROOT_DIR);
    if (!pythonCmd && !process.env.NISKAVA_PYTHON_PATH && !process.env.NISKAVA_PYTHON_BIN) {
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
    const enginePath = path.join(userHome, 'engine');
    const child = spawn(executable, args, {
        cwd: process.cwd(),
        stdio: 'inherit',
        env: {
            ...process.env,
            NISKAVA_ROOT: ROOT_DIR,
            NISKAVA_ENGINE_PATH: enginePath,
            ...(pythonCmd ? {
                NISKAVA_PYTHON_PATH: pythonCmd,
                NISKAVA_PYTHON_BIN: pythonCmd,
                NISKAVA_PYTHON: pythonCmd
            } : {})
        }
    });

    child.on('error', (err) => {
        console.error(`\x1b[31mFailed to start Niskava process: ${err.message}\x1b[0m`);
        process.exit(1);
    });

    child.on('close', (code) => {
        process.exit(code || 0);
    });

    if (!isWindows) {
        process.on('SIGINT', () => {
            if (child && !child.killed) child.kill('SIGINT');
        });
        process.on('SIGTERM', () => {
            if (child && !child.killed) child.kill('SIGTERM');
        });
    }
}

if (require.main === module) {
    main().catch((err) => {
        console.error(`\x1b[31mUnexpected launcher error: ${err.message}\x1b[0m`);
        process.exit(1);
    });
}

module.exports = {
    syncEngineToUserSpace,
    ensureUserVenv,
    checkPythonRuntime,
    copyDirRecursiveSync,
    ensureBinary,
    main
};
