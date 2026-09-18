# 02 — Target Persona & Jobs-to-be-Done (JTBD)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Dokumen ini mendefinisikan siapa pengguna utama Niskava Agent, masalah nyata yang mereka hadapi, serta kerangka kerja *Jobs-to-be-Done (JTBD)* yang memandu desain fitur.

---

## 1. Persona Pengguna Utama

### Persona 1: Arya Pratama (Active Retail Swing Trader)
* **Demografi**: Usia 28 tahun, trader aktif di IDX dengan holding period 3–15 hari, menggunakan terminal Mac/Linux dan smartphone.
* **Perilaku**: Memantau running trade di jam bursa, aktif di grup Telegram saham dan forum Stockbit.
* **Pain Points**:
  * Sering melihat saham melonjak tiba-tiba (+8% dalam 1 jam) dengan volume raksasa.
  * Takut terjebak FOMO pada saham gorengan (*pump and dump*) tanpa katalis fundamental.
  * Googling berita memakan waktu terlalu lama; saat berita selesai dibaca, harga sudah kembali turun (*reversal*).
* **Kebutuhan Solusi**:
  * Alat terminal instan (`niskava investigate TICKER`) yang dalam hitungan detik memberitahu: *Apakah lonjakan ini didukung berita resmi atau murni spekulasi tanpa bukti?*

### Persona 2: Clara Widjaja (Junior Equity Research Associate)
* **Demografi**: Usia 25 tahun, bekerja di perusahaan sekuritas lokal di kawasan SCBD Jakarta.
* **Perilaku**: Menyusun draft morning note dan brief aksi korporasi untuk Senior Analyst setiap pukul 07:30 WIB.
* **Pain Points**:
  * Menghabiskan 1–2 jam setiap malam untuk membaca belasan dokumen PDF keterbukaan informasi di IDXnet dan mencocokkannya dengan chart pergerakan saham.
  * Rentan membuat kesalahan ketik data rasio keuangan jika dilakukan copy-paste manual.
* **Kebutuhan Solusi**:
  * Dashboard visual yang otomatis mengompilasi kronologi kejadian (*timeline events*), perbandingan kinerja emiten terhadap subsektornya, serta kartu bukti (*evidence cards*) yang siap disalin ke laporan riset.

### Persona 3: Dimas Setiawan (Financial Journalist & Fact-Checker)
* **Demografi**: Usia 33 tahun, jurnalis desk pasar modal di media berita finansial nasional.
* **Perilaku**: Menulis berita investigatif mengenai isu korporasi, gagal bayar utang, restrukturisasi, atau transaksi afiliasi.
* **Pain Points**:
  * Sering menerima tip atau desas-desus di media sosial bahwa *"Emiten X terancam bangkrut karena gagal bayar obligasi"*.
  * Butuh pembuktian cepat: Apakah rumor tersebut bertentangan dengan rasio likuiditas kas di laporan keuangan resmi?
* **Kebutuhan Solusi**:
  * Fitur verifikasi otomatis dengan status `CONTRADICTED` yang menunjukkan bukti numerik (Quick ratio, total kas bersih) untuk membantah disinformasi pasar secara objektif.

---

## 2. Kerangka Kerja Jobs-to-be-Done (JTBD)

| Persona | Situasi (When...) | Dorongan (I want to...) | Hasil yang Diharapkan (So I can...) |
|---|---|---|---|
| **Retail Trader** | Ketika melihat volume perdagangan saham melonjak tidak wajar di tengah jam bursa. | Mengetahui apakah ada pengumuman resmi atau katalis nyata di balik lonjakan tersebut dalam < 15 detik. | Mengambil keputusan disiplin tanpa terjebak rumor palsu dan manipulasi harga. |
| **Equity Analyst** | Ketika menyiapkan bahan due diligence atau laporan riset mingguan untuk klien institusi. | Mengotomatisasi penelusuran korelasi antara aksi korporasi emiten dan reaksi pasar historis. | Menghemat 70% waktu riset manual dan fokus pada analisis valuasi strategis. |
| **Financial Journalist** | Ketika beredar rumor viral mengenai skandal atau risiko kebangkrutan emiten tertentu. | Memvalidasi rumor tersebut secara silang terhadap data neraca resmi Sectors dan pengumuman bursa. | Menerbitkan berita faktual berbasis bukti dan menghindari penyebaran hoaks pasar modal. |
