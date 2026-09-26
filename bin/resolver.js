/**
 * NISKAVA AGENT — Path & Platform Asset Resolver
 * Ensures user-space data sovereignty and prevents permission errors (EACCES).
 */

const path = require('path');
const os = require('os');

function getHomeDir() {
    // If run under sudo, prefer the actual invoking user's home directory if available
    if (process.env.SUDO_USER && process.env.SUDO_USER !== 'root') {
        const isLinux = process.platform === 'linux';
        const isDarwin = process.platform === 'darwin';
        if (isLinux) {
            const sudoHome = `/home/${process.env.SUDO_USER}`;
            return sudoHome;
        }
        if (isDarwin) {
            const sudoHome = `/Users/${process.env.SUDO_USER}`;
            return sudoHome;
        }
    }
    return os.homedir() || process.env.HOME || process.env.USERPROFILE || '.';
}

function getNiskavaHome() {
    return path.join(getHomeDir(), '.niskava');
}

function getPlatformAssetName(platform = process.platform, arch = process.arch) {
    if (platform === 'win32') return 'niskava-windows-amd64.exe';
    if (platform === 'darwin') {
        return arch === 'arm64' ? 'niskava-darwin-arm64' : 'niskava-darwin-amd64';
    }
    if (platform === 'linux') {
        return arch === 'arm64' ? 'niskava-linux-arm64' : 'niskava-linux-amd64';
    }
    return platform === 'win32' ? 'niskava.exe' : 'niskava';
}

function getTargetBinaryPath(version) {
    const isWindows = process.platform === 'win32';
    const binName = isWindows ? `niskava-v${version}.exe` : `niskava-v${version}`;
    return path.join(getNiskavaHome(), 'bin', binName);
}

module.exports = {
    getHomeDir,
    getNiskavaHome,
    getPlatformAssetName,
    getTargetBinaryPath
};
