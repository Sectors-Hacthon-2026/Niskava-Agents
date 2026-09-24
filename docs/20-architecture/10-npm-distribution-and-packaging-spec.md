# Spesifikasi Distribusi & Packaging NPM — Niskava Agent

> **Status:** Draft / Planned Implementation Roadmap  
> **Target Package:** `niskava` (atau `@niskava/agent`) di [npmjs.com](https://www.npmjs.com)  
> **Perintah Pengguna:** `npx niskava` atau `npm install -g niskava`  
> **Tanggal Dokumen:** 24 September 2026

---

## 1. Pendahuluan & Tujuan

Dokumen ini adalah spesifikasi arsitektur dan panduan teknis implementasi untuk mendistribusikan **Niskava Agent** ke ekosistem Node.js/NPM. Tujuannya adalah memberikan pengalaman onboarding instan (*zero friction*) bagi analis, trader, pengembang, dan juri hackathon dengan perintah sederhana:

```bash
# Menjalankan wizard setup interaktif instan
npx niskava setup

# Menjalankan terminal AI & Web Workspace
npx niskava
```

Mengingat Niskava Agent menggunakan arsitektur **Tripartite Hybrid (Go Core + Python Engine + Local SQLite WAL)**, proses packaging via NPM membutuhkan penanganan khusus agar binary Go dan dependensi Python (NumPy, Pydantic, NetworkX) dapat berjalan mulus di semua sistem operasi (Linux, macOS, Windows).

---

## 2. Analisis Kesiapan Saat Ini (Current Baseline Audit)

### Komponen yang Sudah Siap (Ready):
* [x] **Go Core Single Executable (`cmd/niskava`):** Binary Go sangat cepat dan tidak memiliki dependensi runtime C (Zero-CGO via `modernc.org/sqlite`).
* [x] **Embedded Web Workspace:** Antarmuka web sudah ter-bundle langsung di dalam binary Go melalui handler HTTP, tidak memerlukan server web eksternal seperti Nginx atau Node server tambahan.
* [x] **Universal ReAct Engine:** Agen Python sudah berbasis protokol standar OpenAI-compatible JSON HTTP, mendukung berbagai model dan router secara dinamis.

### Komponen yang Belum Ada / Harus Disiapkan (Gap Analysis):
* [ ] **Hardcoded Relative Path pada Python Engine:** Di `config.go`, `EnginePath` masih bernilai `./backend/engine`. Jika user menjalankan `niskava` dari folder acak via NPM global, folder `./backend/engine` tidak akan ditemukan.
* [ ] **Root `package.json` & Node Launcher:** Belum ada konfigurasi packaging npm dan skrip launcher JavaScript di level root repositori.
* [ ] **Otomasi Virtual Environment Python:** Pengguna NPM mengharapkan instalasi otomatis. Belum ada mekanisme untuk membuat Python virtual environment (`.venv`) dan memasang `requirements.txt` secara transparan di latar belakang.
* [ ] **Multi-Platform Precompiled Binaries:** Belum ada pipeline otomatis untuk mengompilasi binary Go untuk multi-OS (Linux amd64/arm64, macOS Intel/Apple Silicon, Windows amd64).
* [ ] **Konfigurasi `.npmignore`:** Berkas cache, database lokal, dan log uji coba harus dikecualikan agar ukuran paket NPM tetap kecil (< 2 MB untuk wrapper).

---

## 3. Arsitektur Distribusi NPM (The NPM Wrapper Architecture)

Niskava mengadopsi pola **Hybrid Binary Wrapper** (standar industri yang digunakan oleh *Supabase CLI, Turborepo, Wrangler, Biome, dan esbuild*):

```
┌─────────────────────────────────────────────────────────────┐
│                       PENGGUNA NPM                          │
│               `npx niskava` / `npm i -g niskava`            │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                 NODE.JS LAUNCHER (bin/niskava.js)           │
│  - Deteksi OS (darwin / linux / win32)                      │
│  - Deteksi Arsitektur CPU (x64 / arm64)                     │
│  - Periksa ketersediaan binary Go lokal                     │
│  - Jika belum ada, download binary dari GitHub Releases     │
│  - Spawn subprocess binary Go (meneruskan argv & stdio)     │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                       GO CORE BINARY                        │
│  - Deteksi Dynamic Engine Path (~/.niskava/engine)          │
│  - Deteksi Python Runtime (~/.niskava/venv/bin/python3)     │
│  - Jalankan REST/SSE Gateway, Web Workspace, & Terminal TUI │
└──────────────────────────────┬──────────────────────────────┘
                               │ IPC (JSON Lines)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    PYTHON AGENT ENGINE                      │
│  - Berada di: ~/.niskava/engine/ atau bundled node_modules  │
│  - Virtualenv otomatis: ~/.niskava/venv/                    │
└─────────────────────────────────────────────────────────────┘
```

---

## 4. Rencana Kerja Detail (Step-by-Step Implementation Checklist)

### Fase 1: Dynamic Engine Path & Runtime Resolution di Go Core
Sebelum membuat paket NPM, binary Go Core harus dibuat agnostik terhadap direktori kerja (`CWD`).

1. **Resolusi Lokasi Python Engine Dinamis:**
   Ubah resolusi `EnginePath` di `backend/core/config/config.go` dan `backend/core/ipc/ipc.go` dengan urutan prioritas:
   * **Level 1:** Variabel lingkungan `$NISKAVA_ENGINE_PATH` (jika didefinisikan manual).
   * **Level 2:** Direktori lokal proyek `./backend/engine` (untuk mode development).
   * **Level 3:** Direktori relatif terhadap executable binary Go (`<binary_dir>/../backend/engine` atau `<binary_dir>/engine`).
   * **Level 4:** Direktori global pengguna `~/.niskava/engine/`.
2. **Auto-Provisioning Engine ke `~/.niskava/engine`:**
   Saat menjalankan perintah `niskava setup`, jika engine belum ada di `~/.niskava/engine`, binary otomatis menyalin berkas-berkas skrip Python dari bundle ke folder tersebut.

---

### Fase 2: Otomasi Setup Python Virtual Environment
Pengguna NPM tidak boleh dibebani kompilasi Python manual.

1. **Auto-Venv Provisioner di Go / Script:**
   Tambahkan logika pada perintah `niskava setup` (atau script `scripts/postinstall.js`):
   ```bash
   # 1. Deteksi python3 (versi >= 3.11) di PATH sistem
   python3 --version

   # 2. Buat virtual environment terisolasi di direktori user
   python3 -m venv ~/.niskava/venv

   # 3. Pasang seluruh dependensi resmi (NumPy, Requests, Pydantic, NetworkX, Trafilatura)
   ~/.niskava/venv/bin/pip install --upgrade pip
   ~/.niskava/venv/bin/pip install -r ~/.niskava/engine/requirements.txt
   ```
2. **Update IPC Resolver (`backend/core/ipc/ipc.go`):**
   Prioritaskan `~/.niskava/venv/bin/python3` (atau Windows: `~/.niskava/venv/Scripts/python.exe`) sebagai kandidat interpreter utama.

---

### Fase 3: Struktur Berkas NPM & Node Launcher

1. **Berkas Root `package.json`:**
   ```json
   {
     "name": "niskava",
     "version": "1.0.0",
     "description": "Autonomous Financial OSINT & Market Intelligence Orchestration Platform for IDX",
     "author": "Niskava Team",
     "license": "MIT",
     "homepage": "https://github.com/Sectors-Hacthon-2026/Niskava-Agents",
     "repository": {
       "type": "git",
       "url": "https://github.com/Sectors-Hacthon-2026/Niskava-Agents.git"
     },
     "keywords": [
       "idx",
       "stock-market",
       "financial-osint",
       "ai-agent",
       "sectors-api",
       "autonomous-research",
       "cli"
     ],
     "main": "./bin/launcher.js",
     "bin": {
       "niskava": "./bin/launcher.js"
     },
     "files": [
       "bin/launcher.js",
       "scripts/postinstall.js",
       "backend/engine/",
       "README.md",
       "LICENSE"
     ],
     "scripts": {
       "postinstall": "node ./scripts/postinstall.js"
     },
     "engines": {
       "node": ">=18.0.0"
     }
   }
   ```

2. **Berkas Launcher `bin/launcher.js`:**
   ```javascript
   #!/usr/bin/env node
   const { spawn } = require('child_process');
   const path = require('path');
   const os = require('os');
   const fs = require('fs');

   // Deteksi path binary lokal berdasarkan OS & arsitektur
   function getBinaryPath() {
     const platform = os.platform(); // 'linux', 'darwin', 'win32'
     const arch = os.arch();         // 'x64', 'arm64'
     const ext = platform === 'win32' ? '.exe' : '';
     
     // 1. Cek binary lokal di direktori package
     const localBin = path.join(__dirname, `niskava-${platform}-${arch}${ext}`);
     if (fs.existsSync(localBin)) return localBin;

     // 2. Cek binary global di ~/.niskava/bin
     const globalBin = path.join(os.homedir(), '.niskava', 'bin', `niskava${ext}`);
     if (fs.existsSync(globalBin)) return globalBin;

     return path.join(__dirname, `niskava${ext}`);
   }

   const binaryPath = getBinaryPath();
   if (!fs.existsSync(binaryPath)) {
     console.error(`\x1b[31m[ERROR] Niskava binary tidak ditemukan di: ${binaryPath}\x1b[0m`);
     console.error('Silakan jalankan ulang: npm install -g niskava');
     process.exit(1);
   }

   // Jalankan binary Go dengan meneruskan seluruh argumen CLI
   const child = spawn(binaryPath, process.argv.slice(2), {
     stdio: 'inherit',
     env: process.env
   });

   child.on('exit', (code, signal) => {
     if (signal) process.kill(process.pid, signal);
     else process.exit(code || 0);
   });
   ```

3. **Berkas `.npmignore`:**
   ```text
   # VCS & CI
   .git/
   .github/
   .gitignore

   # Dev & Tests
   tests/
   backend/engine/tests/
   *.test.go
   .pytest_cache/
   __pycache__/
   *.pyc

   # Local runtime & Databases (Kedaulatan Data Lokal)
   .env
   .venv/
   *.db
   *.db-wal
   *.db-shm

   # Intermediate builds
   bin/*.exe
   bin/niskava
   clients/web/node_modules/
   ```

---

### Fase 4: Otomasi Kompilasi Multi-Platform (GoReleaser)
Untuk mendistribusikan binary siap pakai tanpa mengharuskan pengguna memiliki compiler Go, siapkan `.goreleaser.yaml`:

```yaml
version: 2
project_name: niskava

builds:
  - id: niskava-cli
    main: ./cmd/niskava
    binary: niskava
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - name_template: "{{ .ProjectName }}-{{ .Os }}-{{ .Arch }}"
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

release:
  github:
    owner: Sectors-Hacthon-2026
    name: Niskava-Agents
```

---

### Fase 5: Skrip `scripts/postinstall.js` (Download Binary & Setup Venv)
Saat pengguna mengetik `npm i -g niskava`, script ini otomatis:
1. Mendeteksi OS & arsitektur komputer pengguna.
2. Mengunduh tarball binary Go yang sesuai dari GitHub Releases terbaru.
3. Mengekstrak binary ke folder `~/.niskava/bin/niskava`.
4. Menyalin skrip Python `backend/engine/` ke `~/.niskava/engine/`.
5. Menampilkan panduan:
   ```
   ⚡ Niskava Agent berhasil terpasang!
   Ketik 'niskava setup' untuk mengonfigurasi API Key & Model AI.
   ```

---

## 5. Prosedur Uji Coba & Verifikasi Lokal Sebelum Publish

Sebelum mempublikasikan package ke registry publik npm:

```bash
# 1. Pack kering (dry-run) untuk melihat ukuran dan daftar file yang masuk
npm pack --dry-run

# 2. Test instalasi lokal menggunakan npm link
npm link

# 3. Uji coba memanggil binary dari folder acak
cd /tmp
niskava --help
niskava setup

# 4. Lepas link setelah pengujian selesai
npm unlink -g niskava
```

---

## 6. Prosedur Publikasi ke NPM Registry

1. **Login ke NPM:**
   ```bash
   npm login
   ```
2. **Verifikasi Nama Package Tersedia:**
   ```bash
   npm view niskava
   # Jika nama 'niskava' sudah dipakai orang lain, gunakan scope: '@niskava/agent'
   ```
3. **Publish ke Registry:**
   ```bash
   # Untuk unscoped package:
   npm publish --access public

   # Untuk scoped package:
   npm publish --access public
   ```

---

## 7. Kesimpulan & Manfaat Strategis untuk Hackathon

Mengaktifkan distribusi via NPM memberikan keunggulan kompetitif besar:
1. **Adopsi Instan bagi Juri:** Juri tidak perlu setup Go SDK, Git clone, atau konfigurasi manual. Cukup menjalankan `npx niskava` di terminal mereka.
2. **Kesesuaian dengan Hukum 4 (Kedaulatan Data Lokal):** Seluruh data, sesi, grafik memori, dan cache tetap tersimpan di mesin lokal pengguna (`~/.niskava/`), NPM hanya berfungsi sebagai mekanisme pengiriman binary.
3. **Standar Profesional:** Menunjukkan kematangan *developer experience* (DX) tingkat enterprise.
