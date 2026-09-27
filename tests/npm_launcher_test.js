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
assert.strictEqual(pkg.version, '0.1.4', `Expected version to be bumped to 0.1.4, got: ${pkg.version}`);
assert(pkg.files.includes('backend/engine/requirements.txt'), 'package.json files must explicitly include backend/engine/requirements.txt');
assert(pkg.files.includes('!**/.pytest_cache'), 'package.json files must exclude .pytest_cache');

console.log('✓ All NPM resolver and user-space isolation tests passed successfully!');

