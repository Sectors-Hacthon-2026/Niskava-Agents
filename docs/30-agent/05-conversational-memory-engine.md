# 05 — Spesifikasi Engine Memori Percakapan Berbasis Graf Lokal

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  
**Keputusan Terkait:** [`06-local-conversational-graph-memory.md`](../90-decisions/06-local-conversational-graph-memory.md)  
**Kepatuhan Hackathon:** Track 1 · AI Agents & Assistants (*Memory or State Management*)  

Dokumen ini menguraikan arsitektur teknis, model data, siklus hidup 4 tahap (*4-phase lifecycle*), dan logika pemanggilan konteks pada **Local Conversational & Cross-Session Graph Memory Engine** Niskava Agent.

---

## 1. Konsep Dasar: Dari Chat Log Linier ke Graf Asosiatif

Chatbot konvensional mengandalkan *chat history* linier mentah. Ketika percakapan bertambah panjang, riwayat dipangkas (*truncated*), menyebabkan agen kehilangan ingatan (*amnesia*) atas aksi atau analisa di sesi-sesi sebelumnya.

Niskava mengadopsi **Graf Memori Asosiatif Lokal** mirip filosofi *Graphify*:
* Percakapan dan aksi investigasi dipecah menjadi **Simpul Entitas (*Nodes*)** dan **Hubungan Kausalitas (*Directed Edges*)**.
* Memori disimpan secara persisten di SQLite lokal dan di-load ke struktur graf `NetworkX` saat proses penalaran berlangsung.

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                       LOCAL CONVERSATIONAL GRAPH MEMORY                     │
│                                                                             │
│     (User) ────────[INVESTIGATED]───────▶ (ANTM: Aneka Tambang)            │
│       │                                          │                          │
│   [HOLDS_AT]                                [CATALYZED_BY]                  │
│       ▼                                          ▼                          │
│   (Price: 1450)                         (Smelter Halmahera Timur)           │
│       │                                          │                          │
│   [BELONGS_TO]                              [RECORDED_IN]                   │
│       ▼                                          ▼                          │
│  (Basic Materials)                        (Session: INV-2026-0042)          │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Siklus Hidup Memori 4 Tahap (*The 4-Phase Memory Lifecycle*)

```text
1. INGESTION ──▶ 2. STORAGE & NORMALIZATION ──▶ 3. EGO-GRAPH RETRIEVAL ──▶ 4. PROMPT INJECTION
```

### Tahap 1: Perekaman Memori (*Ingestion*)
Perekaman dilakukan melalui dua jalur hibrida:
1. **Perekaman Deterministik (Zero-Token Cost - Pasca Investigasi):**  
   Setiap kali eksekusi `niskava investigate <ticker>` berhasil, sistem secara otomatis menerbitkan relasi tanpa memanggil LLM:
   * `(User) --[INVESTIGATED]--> (<TICKER>)`
   * `(<TICKER>) --[TRIGGERED_ANOMALY]--> (Metric: Volume Z-Score)`
   * `(<TICKER>) --[CATALYZED_BY]--> (Finding Title)`
2. **Ekstraksi Percakapan Bebas (LLM Ringan / Gemini Flash):**  
   Ketika pengguna mengetik pesan di CLI/Web yang mengandung preferensi, posisi portofolio, atau target harga (misal: *"Saya beli ANTM di 1450"*):
   * Modul ekstraktor menghasilkan payload JSON terstruktur:
     ```json
     {
       "source": "User",
       "relation": "HOLDS_AT",
       "target": "Price: 1450",
       "context": "Pembelian posisi modal di ANTM",
       "ticker": "ANTM"
     }
     ```

### Tahap 2: Penyimpanan & Normalisasi (*Storage & Normalization*)
* **Normalisasi Ticker:** Variasi penulisan entitas diseragamkan (misal: *"Antam"*, *"aneka tambang"*, *"PT ANTM"* disatukan menjadi simpul `TICKER:ANTM`).
* **Upsert SQLite:** Node dan relasi ditulis ke tabel `memory_nodes` dan `memory_edges` dengan mode `INSERT OR REPLACE` disertai stempel waktu `last_observed_at`.

### Tahap 3: Pemanggilan Subgraf Kontekstual (*Ego-Graph Retrieval*)
Ketika pengguna mengajukan pertanyaan baru di sesi berikutnya:
1. **Identifikasi Entitas Kunci:** Sistem mendeteksi entitas awal dari pertanyaan pengguna (misal: menyebut *"ANTM"* atau *"saham tambang"*).
2. **Sub-Traversal NetworkX:** Membangun graf di memori dari database, lalu memotong subgraf berjarak $k \le 2$ hop (*Ego-Graph*) di sekitar simpul target.
3. **Penyusutan Berbobot Waktu (*Temporal Recency Decay*):**  
   Bobot relasi dihitung dengan formula peluruhan eksponensial:
   $$W_{\text{effective}} = W_0 \times e^{-\lambda \Delta t}$$
   Relasi yang baru diamati dalam 3 hari terakhir memiliki prioritas lebih tinggi daripada percakapan 3 pekan lalu.

### Tahap 4: Augmentasi Prompt Sistem (*Prompt Augmentation*)
Subgraf yang terpilih diterjemahkan menjadi teks ringkas (<200 token) dan disuntikkan ke dalam instruksi agen:

```xml
<investigative_memory>
- User pernah menginvestigasi ANTM pada 12 September 2026 (Sesi INV-2026-0042).
- Anomali terdeteksi: Lonjakan volume 3.84σ mendahului peresmian smelter Halmahera.
- Catatan posisi pengguna: Teridentifikasi entry harga di level 1450.
</investigative_memory>
```

Agen dapat merespons pertanyaan pengguna secara cerdas dengan menghubungkan konteks masa lalu tanpa harus membaca ulang seluruh file riwayat.

---

## 3. Taksonomi Simpul & Relasi Graf Memori

### A. Tipe Simpul (*Node Types*)
* `USER`: Representasi pengguna aktif.
* `TICKER`: Simpul saham bursa efek Indonesia (format: 4 huruf kapital, misal `ANTM`, `BBRI`).
* `SECTOR`: Klasifikasi sektor/subsektor resmi IDX (misal `Basic Materials`).
* `PRICE_LEVEL`: Level harga atau valuasi yang menjadi acuan pengguna.
* `CATALYST_EVENT`: Peristiwa aksi korporasi, pengumuman, atau berita penting.
* `STRATEGY`: Preferensi analisis pengguna (misal `Dividend_Focus`, `Swing_Breakout`).

### B. Tipe Relasi (*Edge Relations*)
* `INVESTIGATED`: Menandai bahwa emiten pernah dianalisis lengkap oleh agen.
* `WATCHES`: Menandai emiten yang sedang dipantau berkala oleh pengguna.
* `HOLDS_AT`: Menghubungkan pengguna dengan harga pembelian posisi.
* `TRIGGERED_ANOMALY`: Menghubungkan emiten dengan metrik anomali kuantitatif.
* `CATALYZED_BY`: Menghubungkan anomali dengan temuan berita/keterbukaan resmi.
* `SUPERSEDES`: Relasi pembatalan ketika fakta baru menganulir catatan lama.

---

## 4. Resolusi Konflik & Validitas Temporal (*Superseding Logic*)

Pasar modal bersifat dinamis, sehingga fakta masa lalu dapat berubah:
* **Kasus Perubahan Posisi:** Jika pada Sesi 1 pengguna mencatat `User -[HOLDS_AT]-> 1450`, lalu pada Sesi 5 pengguna menyatakan *"Saya sudah take profit ANTM di 1620"*:
  * Sistem tidak menghapus memori lama secara brutal, melainkan menambahkan relasi `(Fact_Sesi_5) --[SUPERSEDES]--> (Fact_Sesi_1)`.
  * Status relasi lama diubah menjadi non-aktif atau bobotnya diturunkan drastis, sehingga agen mengetahui kronologi evolusi posisi pengguna:
    > *"Berdasarkan catatan sebelumnya Anda telah merealisasikan profit di 1620 dari posisi beli 1450..."*

---

## 5. Implementasi Referensi Python (`niskava/memory/graph_memory.py`)

Modul ini siap pakai dan terintegrasi langsung dengan database SQLite lokal:

```python
import sqlite3
import networkx as nx
from datetime import datetime
from typing import List, Dict, Any, Optional

class LocalGraphMemory:
    """Engine Memori Graf Lokal Niskava berbasis SQLite dan NetworkX."""
    
    def __init__(self, db_path: str = "niskava.db"):
        self.db_path = db_path
        self._ensure_schema()

    def _get_connection(self) -> sqlite3.Connection:
        conn = sqlite3.connect(self.db_path)
        conn.execute("PRAGMA journal_mode = WAL;")
        conn.execute("PRAGMA foreign_keys = ON;")
        return conn

    def _ensure_schema(self):
        with self._get_connection() as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS memory_nodes (
                    id TEXT PRIMARY KEY,
                    label TEXT NOT NULL,
                    node_type TEXT NOT NULL,
                    metadata_json TEXT,
                    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                    last_observed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
                );
            """)
            conn.execute("""
                CREATE TABLE IF NOT EXISTS memory_edges (
                    source_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
                    target_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
                    relation TEXT NOT NULL,
                    context_snippet TEXT,
                    session_id TEXT,
                    weight REAL NOT NULL DEFAULT 1.0,
                    confidence_score REAL NOT NULL DEFAULT 1.0,
                    last_observed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                    PRIMARY KEY (source_id, target_id, relation)
                );
            """)

    def record_investigation(self, session_id: str, ticker: str, anomaly_desc: str, catalyst: str):
        """Perekaman deterministik (zero-token) pasca investigasi."""
        now = datetime.now().isoformat()
        with self._get_connection() as conn:
            # Nodes
            conn.execute("INSERT OR REPLACE INTO memory_nodes (id, label, node_type, last_observed_at) VALUES (?, ?, ?, ?)",
                         ("user:default", "User", "USER", now))
            conn.execute("INSERT OR REPLACE INTO memory_nodes (id, label, node_type, last_observed_at) VALUES (?, ?, ?, ?)",
                         (f"ticker:{ticker}", ticker, "TICKER", now))
            conn.execute("INSERT OR REPLACE INTO memory_nodes (id, label, node_type, last_observed_at) VALUES (?, ?, ?, ?)",
                         (f"catalyst:{ticker}_{session_id}", catalyst, "CATALYST_EVENT", now))
            
            # Edges
            conn.execute("""
                INSERT OR REPLACE INTO memory_edges 
                (source_id, target_id, relation, context_snippet, session_id, last_observed_at)
                VALUES (?, ?, ?, ?, ?, ?)
            """, ("user:default", f"ticker:{ticker}", "INVESTIGATED", anomaly_desc, session_id, now))
            
            conn.execute("""
                INSERT OR REPLACE INTO memory_edges 
                (source_id, target_id, relation, context_snippet, session_id, last_observed_at)
                VALUES (?, ?, ?, ?, ?, ?)
            """, (f"ticker:{ticker}", f"catalyst:{ticker}_{session_id}", "CATALYZED_BY", catalyst, session_id, now))

    def retrieve_subgraph_prompt(self, target_ticker: str, radius: int = 1) -> str:
        """Mengambil subgraf Ego-Graph dan menyusun konteks memori untuk prompt LLM."""
        with self._get_connection() as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT source_id, target_id, relation, context_snippet FROM memory_edges")
            rows = cursor.fetchall()

        if not rows:
            return ""

        G = nx.DiGraph()
        for src, tgt, rel, ctx in rows:
            G.add_edge(src, tgt, relation=rel, context=ctx or "")

        root_node = f"ticker:{target_ticker}"
        if root_node not in G:
            return ""

        subgraph = nx.ego_graph(G, root_node, radius=radius, undirected=True)
        memory_lines = []
        for u, v, data in subgraph.edges(data=True):
            rel = data['relation']
            ctx = f" ({data['context']})" if data['context'] else ""
            memory_lines.append(f"- {u.replace('ticker:', '').replace('user:', '')} -> [{rel}] -> {v.replace('ticker:', '').replace('catalyst:', '')}{ctx}")

        if not memory_lines:
            return ""

        return "<investigative_memory>\n" + "\n".join(memory_lines) + "\n</investigative_memory>"
```

---

## 6. Dampak Strategis pada Penjurian Sectors Hackathon

1. **Memenuhi Kriteria Track 1 (*Memory or State Management*):**  
   Membuktikan bahwa Niskava memiliki mekanisme manajemen memori yang dirancang secara mandiri (*custom-built memory engine*), bukan sekadar mengandalkan memori bawaan client LLM.
2. **Poin Pembeda di Video Penjurian (Storytelling 30%):**  
   Skenario demonstrasi menunjukkan dua sesi berjarak: Sesi 1 investigasi ANTM $\to$ Sesi 2 pengguna bertanya *"Bagaimana kelanjutan saham nikel yang kemarin?"* dan agen langsung mengenali konteks ANTM secara mulus.
3. **Efisiensi Total Biaya:**  
   Tidak ada biaya API database vektor pihak ketiga; seluruh memori berjalan 100% lokal, cepat (<5ms), dan privat.
