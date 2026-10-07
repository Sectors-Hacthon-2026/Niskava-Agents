const assert = require('assert');
const path = require('path');
const os = require('os');
const { getTargetBinaryPath, getPlatformAssetName, getNiskavaHome } = require('../bin/resolver.js');

// Test 1: User-space isolation (guaranteed writable, zero EACCES risk)
const home = os.homedir() || process.env.HOME;
const targetPath = getTargetBinaryPath('0.1.2');

assert(targetPath.startsWith(home), `Target binary must reside in user home directory (${home}), got: ${targetPath}`);
assert(targetPath.includes('.niskava'), `Target binary must be inside .niskava folder, got: ${targetPath}`);
assert(!targetPath.includes('node_modules'), `Target binary must NOT be in node_modules, got: ${targetPath}`);

// Test 2: Platform asset names mapping
assert.strictEqual(getPlatformAssetName('linux', 'x64'), 'niskava-linux-amd64');
assert.strictEqual(getPlatformAssetName('linux', 'arm64'), 'niskava-linux-arm64');
assert.strictEqual(getPlatformAssetName('darwin', 'arm64'), 'niskava-darwin-arm64');
assert.strictEqual(getPlatformAssetName('darwin', 'x64'), 'niskava-darwin-amd64');
assert.strictEqual(getPlatformAssetName('win32', 'x64'), 'niskava-windows-amd64.exe');

// Test 3: Niskava home directory
assert.strictEqual(getNiskavaHome(), path.join(home, '.niskava'));

// Test 4: Package.json files and version check
const pkg = require('../package.json');
assert.strictEqual(pkg.version, '0.2.1', `Expected version to be bumped to 0.2.1, got: ${pkg.version}`);
assert(pkg.files.includes('backend/engine/requirements.txt'), 'package.json files must explicitly include backend/engine/requirements.txt');
assert(pkg.files.includes('!**/.pytest_cache'), 'package.json files must exclude .pytest_cache');

// Test 5: syncEngineToUserSpace and file exclusion tests
const fs = require('fs');
const { syncEngineToUserSpace, copyDirRecursiveSync, checkPythonRuntime } = require('../bin/index.js');

const tempTestDir = path.join(os.tmpdir(), `niskava-test-${Date.now()}`);
fs.mkdirSync(tempTestDir, { recursive: true });

try {
    // Create mock source tree
    const mockRoot = path.join(tempTestDir, 'mock-repo');
    const mockEngine = path.join(mockRoot, 'backend', 'engine');
    fs.mkdirSync(mockEngine, { recursive: true });
    fs.writeFileSync(path.join(mockEngine, 'runner.py'), '#!/usr/bin/env python3\nprint("test")\n');
    fs.writeFileSync(path.join(mockEngine, 'requirements.txt'), 'numpy>=1.26.0\n');
    
    // Create items that should be excluded
    const mockPycache = path.join(mockEngine, '__pycache__');
    fs.mkdirSync(mockPycache, { recursive: true });
    fs.writeFileSync(path.join(mockPycache, 'runner.cpython-311.pyc'), 'fake bytecode');
    fs.writeFileSync(path.join(mockEngine, 'stray.pyc'), 'fake bytecode');

    // Create target user space
    const mockUserHome = path.join(tempTestDir, 'mock-home');
    fs.mkdirSync(mockUserHome, { recursive: true });

    const success = syncEngineToUserSpace(mockRoot, mockUserHome);
    assert.strictEqual(success, true, 'syncEngineToUserSpace should return true for valid source tree');

    // Verify files were copied to both ~/.niskava/engine and ~/.niskava/backend/engine
    const dest1 = path.join(mockUserHome, 'engine', 'runner.py');
    const dest2 = path.join(mockUserHome, 'backend', 'engine', 'runner.py');
    assert(fs.existsSync(dest1), `Expected ${dest1} to exist`);
    assert(fs.existsSync(dest2), `Expected ${dest2} to exist`);

    // Verify bytecode and __pycache__ are excluded
    const pycacheDest = path.join(mockUserHome, 'engine', '__pycache__');
    const pycDest = path.join(mockUserHome, 'engine', 'stray.pyc');
    assert(!fs.existsSync(pycacheDest), '__pycache__ should NOT be copied');
    assert(!fs.existsSync(pycDest), '*.pyc files should NOT be copied');

    // Test 6: Python runtime check
    const py = checkPythonRuntime(mockUserHome, mockRoot);
    // On systems with Python 3.11+, py should be non-null
    if (py) {
        assert(typeof py === 'string' && py.length > 0, 'checkPythonRuntime should return a non-empty string path');
    }
} finally {
    // Cleanup temporary test directory
    try {
        fs.rmSync(tempTestDir, { recursive: true, force: true });
    } catch (_) {}
}

console.log('✓ All NPM resolver, engine-sync, and user-space isolation tests passed successfully!');


