#!/usr/bin/env node

/**
 * NISKAVA AGENT — Postinstall Hook
 * Pre-downloads the matching native binary to user-space cache (~/.niskava/bin/)
 * during npm install so execution is instantaneous and avoids EACCES.
 */

const { getTargetBinaryPath, getPlatformAssetName } = require('../bin/resolver');
const PKG_VERSION = require('../package.json').version;
const GITHUB_REPO = 'Sectors-Hacthon-2026/Niskava-Agents';
const https = require('https');
const fs = require('fs');
const path = require('path');

const isWindows = process.platform === 'win32';

async function preDownload() {
    const targetBinary = getTargetBinaryPath(PKG_VERSION);
    if (fs.existsSync(targetBinary)) {
        return; // Already present
    }

    const assetName = getPlatformAssetName();
    const url = `https://github.com/${GITHUB_REPO}/releases/download/v${PKG_VERSION}/${assetName}`;

    console.log(`\x1b[36m[@niskava/agent] Pre-caching platform binary (${process.platform}/${process.arch})...\x1b[0m`);

    const destDir = path.dirname(targetBinary);
    if (!fs.existsSync(destDir)) {
        try {
            fs.mkdirSync(destDir, { recursive: true });
        } catch (_) {
            // If home dir is inaccessible during restricted postinstall, silently fallback to runtime download
            return;
        }
    }

    const tempPath = `${targetBinary}.tmp.${Date.now()}`;

    function followRedirect(currentUrl, redirectCount = 0) {
        if (redirectCount > 5) return;

        const reqOptions = {
            headers: {
                'User-Agent': 'niskava-postinstall',
                'Accept': 'application/octet-stream'
            }
        };

        https.get(currentUrl, reqOptions, (res) => {
            if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
                return followRedirect(res.headers.location, redirectCount + 1);
            }

            if (res.statusCode !== 200) {
                return; // Gracefully continue; runtime launcher will attempt again or report clean error
            }

            const fileStream = fs.createWriteStream(tempPath);
            res.pipe(fileStream);

            fileStream.on('finish', () => {
                fileStream.close(() => {
                    try {
                        if (!isWindows) fs.chmodSync(tempPath, 0o755);
                        fs.renameSync(tempPath, targetBinary);
                        console.log(`\x1b[32m[@niskava/agent] Binary successfully cached to ${targetBinary}\x1b[0m`);
                    } catch (_) {}
                });
            });

            fileStream.on('error', () => {
                fs.unlink(tempPath, () => {});
            });
        }).on('error', () => {
            fs.unlink(tempPath, () => {});
        });
    }

    followRedirect(url);
}

preDownload().catch(() => {});
