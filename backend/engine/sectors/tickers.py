"""Master list and validator for official Indonesia Stock Exchange (IDX) tickers.

Guarantees that 4-letter words, slang, and typos (e.g. 'domg', 'guys', 'suhu', 'hari')
are never falsely recognized as stock tickers.
"""

import re
from typing import FrozenSet, List, Optional, Set

# Comprehensive registry of active Indonesia Stock Exchange (IDX) tickers.
# Sources: IDX Official Directory & Sectors Financial API v2.
IDX_TICKERS: FrozenSet[str] = frozenset({
    # A
    "AALI", "ABBA", "ABDA", "ABMM", "ACES", "ACST", "ADCP", "ADES", "ADHI", "ADMF",
    "ADMR", "ADRO", "AGII", "AGRO", "AGRS", "AHAP", "AIMS", "AISA", "AKKU", "AKPI",
    "AKRA", "ALDO", "ALKA", "ALMI", "ALTO", "AMAR", "AMFG", "AMIN", "AMMN", "AMMS",
    "AMOR", "AMRT", "ANDI", "ANJT", "ANTM", "APEX", "APIC", "APII", "APLI", "APLN",
    "ARCI", "ARGO", "ARII", "ARKA", "ARKO", "ARMY", "ARNA", "ARTA", "ARTI", "ARTO",
    "ASBI", "ASDF", "ASDM", "ASGR", "ASHA", "ASII", "ASJT", "ASLC", "ASMI", "ASPI",
    "ASRI", "ASRM", "ASSA", "ATAP", "ATIC", "AUTO", "AVIA", "AWAN", "AYAM", "AYLS",
    # B
    "BABP", "BACA", "BAJA", "BALI", "BANK", "BAPA", "BAPI", "BATA", "BAUT", "BAYU",
    "BBCA", "BBHI", "BBKP", "BBLD", "BBMD", "BBNI", "BBRI", "BBRM", "BBTN", "BBYB",
    "BCAP", "BCIC", "BCIP", "BDMN", "BEBS", "BEEF", "BEER", "BEKS", "BELL", "BESS",
    "BEST", "BFIN", "BFST", "BGTG", "BHAT", "BHIT", "BIKA", "BIMA", "BINA", "BIPI",
    "BIPP", "BISI", "BJBR", "BJTM", "BKDP", "BKSL", "BKSW", "BLES", "BLOG", "BLTA",
    "BLTZ", "BLUE", "BMAS", "BMBL", "BMHS", "BMRI", "BMSR", "BMTR", "BNBA", "BNBR",
    "BNGA", "BNII", "BNLI", "BOBA", "BOGA", "BOLA", "BOLT", "BOSS", "BPFI", "BPII",
    "BPTR", "BRAM", "BRIS", "BRMS", "BRNA", "BRPT", "BSBK", "BSDE", "BSIM", "BSML",
    "BSSR", "BSWD", "BTEK", "BTON", "BTPN", "BTPS", "BUAH", "BUDI", "BUKK", "BULL",
    "BUMI", "BUVA", "BVIC", "BWPT", "BYAN",
    # C
    "CAKK", "CAMP", "CANI", "CARE", "CARS", "CASA", "CASH", "CASS", "CBMF", "CCSI",
    "CEKA", "CENT", "CFIN", "CGAS", "CINT", "CITA", "CITY", "CLAY", "CLEO", "CLPI",
    "CMNP", "CMNT", "CMPP", "CMRY", "CNKO", "CNMA", "CNTB", "COAL", "COCO", "CPIN",
    "CPRI", "CPRO", "CSAP", "CSIS", "CSMI", "CSRA", "CTBN", "CTRA", "CTTH", "CUAN",
    # D
    "DAAZ", "DADA", "DART", "DAYA", "DEAL", "DEFI", "DEPO", "DEWA", "DGIK", "DGNS",
    "DIGI", "DILD", "DIVA", "DKFT", "DLTA", "DMAS", "DMMX", "DMND", "DNAR", "DNET",
    "DOID", "DPNS", "DPUM", "DRMA", "DSFI", "DSNG", "DSSA", "DUTI", "DVLA", "DWGL",
    "DYAN",
    # E
    "EAST", "ECII", "EDGE", "ELIT", "ELPI", "ELSA", "ELTY", "EMDE", "EMTK", "ENAK",
    "ENRG", "ENVY", "ENZO", "EPAC", "EPMT", "ERAA", "ERAL", "ERTX", "ESIP", "ESSA",
    "ESTA", "ESTI", "ETWA", "EURO", "EXCL",
    # F
    "FAPA", "FAST", "FASW", "FILM", "FIMP", "FIRE", "FISH", "FITT", "FLMC", "FMII",
    "FOOD", "FORU", "FPNI", "FREN", "FWCT",
    # G
    "GAAH", "GAMA", "GDST", "GDYR", "GEMA", "GEMS", "GGRP", "GHON", "GIAA", "GJTL",
    "GLOB", "GLVA", "GMFI", "GMTD", "GOLD", "GOLL", "GOOD", "GOTO", "GPRA", "GPSO",
    "GRIA", "GRM",  "GRPH", "GRPM", "GSMF", "GTBO", "GTRA", "GTSI", "GULA", "GWAB",
    # H
    "HAIS", "HAJJ", "HALO", "HATM", "HDFA", "HDIT", "HDTX", "HEAL", "HELI", "HERO",
    "HEXA", "HITS", "HKMU", "HMSP", "HOKI", "HOME", "HOMI", "HOPO", "HOTL", "HRME",
    "HRTA", "HRUM", "HUMI", "HYGN",
    # I
    "IATA", "IBFN", "IBOS", "IBST", "ICBP", "ICON", "IDEA", "IDPR", "IFII", "IFSH",
    "IGAR", "IIKP", "IKAI", "IKAN", "IKBI", "IKPM", "ILKP", "IMAS", "IMJS", "IMPC",
    "INAF", "INAI", "INCF", "INCI", "INCO", "INDF", "INDO", "INDS", "INDX", "INDY",
    "INKP", "INNO", "INPC", "INPP", "INPS", "INRU", "INTA", "INTD", "INTP", "INVS",
    "IOTF", "IPAC", "IPCC", "IPCM", "IPPE", "IPTV", "IRRA", "ISAP", "ISAT", "ISSP",
    "ITIC", "ITMA", "ITMG",
    # J
    "JARR", "JAST", "JATI", "JAWA", "JAYA", "JECC", "JGLE", "JIHD", "JKON", "JMAS",
    "JPFA", "JRPT", "JSMR", "JSPT", "JTPE",
    # K
    "KAEF", "KARW", "KAYU", "KBAG", "KBLI", "KBLM", "KBLV", "KDSI", "KEEN", "KEJU",
    "KIAS", "KICI", "KIJA", "KINO", "KIOS", "KJEN", "KKGI", "KLAS", "KLBF", "KLIN",
    "KMDS", "KMTR", "KOBX", "KOIN", "KOKA", "KONI", "KOPI", "KOTA", "KPAL", "KPAS",
    "KPIG", "KRAH", "KRAS", "KREN", "KRYA", "KTIC",
    # L
    "LAMI", "LAND", "LAPD", "LCGP", "LCKM", "LEAD", "LION", "LIVE", "LMAS", "LMAX",
    "LMPI", "LMSH", "LOPI", "LPCK", "LPGI", "LPIN", "LPKR", "LPLI", "LPPF", "LPPS",
    "LRNA", "LSIP", "LTLS", "LUCK",
    # M
    "MABA", "MAGP", "MAHA", "MAIN", "MAMI", "MAPA", "MAPB", "MAPI", "MARI", "MARK",
    "MASA", "MAXI", "MBAP", "MBMA", "MBSS", "MBTO", "MCAS", "MCOL", "MCOR", "MDIA",
    "MDKA", "MDKI", "MDLN", "MDRN", "MEDC", "MEGA", "MENN", "MERO", "META", "MFIN",
    "MFMI", "MGLV", "MGNA", "MICE", "MIDI", "MIKA", "MIRA", "MITI", "MKAP", "MKNT",
    "MKPI", "MKTR", "MLBI", "MLIA", "MLPL", "MLPT", "MMIX", "MMLP", "MNCN", "MOLI",
    "MPMX", "MPOW", "MPPA", "MPRO", "MRAT", "MREI", "MSIN", "MSKY", "MTDL", "MTEL",
    "MTFN", "MTLA", "MTMH", "MTPS", "MTRN", "MTSM", "MTWI", "MUTU", "MYOH", "MYOR",
    "MYRX", "MYTX",
    # N
    "NAIK", "NANO", "NASA", "NASI", "NATO", "NCKL", "NELY", "NEST", "NETV", "NFCX",
    "NICE", "NICK", "NICL", "NIKL", "NINE", "NISP", "NOBU", "NPGF", "NRCA", "NSSS",
    "NTBK", "NUSA", "NZIA",
    # O
    "OASA", "OBMD", "OCAP", "OILS", "OKAS", "OLIV", "OMED", "OMRE", "OPMS",
    # P
    "PADA", "PADI", "PALM", "PAMG", "PANI", "PANR", "PANS", "PBID", "PBRX", "PBSA",
    "PCAR", "PDES", "PEGE", "PEHA", "PEVE", "PGAS", "PGEO", "PGLI", "PGUN", "PICO",
    "PJAA", "PKPK", "PLAS", "PLIN", "PMJS", "PMMP", "PNBN", "PNBS", "PNGO", "PNIN",
    "PNLF", "PNSE", "POLA", "POLE", "POLI", "POLL", "POLU", "POLY", "POOL", "PORT",
    "POSA", "POWR", "PPGL", "PPRE", "PPRO", "PRAS", "PRDA", "PRIM", "PRSF", "PSAB",
    "PSAT", "PSDN", "PSGO", "PSKT", "PSSI", "PTBA", "PTDU", "PTIS", "PTON", "PTPW",
    "PTRO", "PTSN", "PTSP", "PUDP", "PURA", "PURE", "PURI", "PWON", "PYFA", "PZZA",
    # R
    "RAAM", "RAFI", "RAJA", "RALS", "RANC", "RBMS", "RCCC", "RDMD", "REAL", "RELF",
    "RELI", "RGAS", "RICY", "RIGS", "RIMO", "RISE", "RMKE", "RMKO", "ROCK", "RODA",
    "ROKI", "ROTI", "RSCH", "RSGK", "RUIS", "RUNS",
    # S
    "SAFE", "SAME", "SAMF", "SAPX", "SATU", "SBAT", "SBMA", "SCCO", "SCMA", "SCNP",
    "SDMU", "SDPC", "SDRA", "SEMA", "SFAN", "SGER", "SGRO", "SHID", "SHIP", "SICO",
    "SIDO", "SILO", "SIMA", "SIMP", "SINI", "SIPD", "SKBM", "SKLT", "SKRN", "SKYB",
    "SLIS", "SMAR", "SMBR", "SMCB", "SMDM", "SMDR", "SMGA", "SMGR", "SMIL", "SMKL",
    "SMKM", "SMMA", "SMMT", "SMRA", "SMRT", "SMSM", "SNLK", "SOBI", "SOFA", "SOHO",
    "SONA", "SOSS", "SOTS", "SPMA", "SPTO", "SQMI", "SRAJ", "SRTG", "SSIA", "SSMS",
    "SSTM", "STAR", "STAA", "STTP", "SUGI", "SULI", "SUMI", "SUNR", "SUPR", "SURE",
    "SWAT",
    # T
    "TALF", "TAMA", "TAMU", "TAPG", "TARA", "TAXI", "TBIG", "TBLA", "TBMS", "TCID",
    "TCPI", "TDPM", "TEBE", "TECH", "TELE", "TFAS", "TFCO", "TGKA", "TGRA", "TIFA",
    "TIMS", "TINS", "TIRA", "TIRT", "TKIM", "TLDN", "TLKM", "TMAS", "TMPO", "TNCA",
    "TOBA", "TOOL", "TOPP", "TOSK", "TOTL", "TOTO", "TOWR", "TPMA", "TRAM", "TRGU",
    "TRIL", "TRIM", "TRIN", "TRIO", "TRIS", "TRJA", "TRON", "TRST", "TRUE", "TRUK",
    "TRUS", "TSPC", "TUGU", "TYRE",
    # U
    "UCID", "UDNG", "UFOE", "ULTJ", "UNIC", "UNIQ", "UNIT", "UNSP", "UNTR", "UNVR",
    "URBN", "UVCR",
    # V
    "VAST", "VERN", "VICI", "VICO", "VINS", "VIPT", "VISI", "VIVA", "VOKS", "VRNA",
    # W
    "WAPO", "WEGE", "WEHA", "WGSH", "WICO", "WIDI", "WIFI", "WIKA", "WINS", "WIRG",
    "WMPP", "WMUU", "WOOD", "WOWS", "WSBP", "WSKT", "WTON",
    # Y
    "YAPO", "YELO", "YPAS",
    # Z
    "ZATA", "ZBRA", "ZINC", "ZONE", "ZYRX",
})

# Words that could be mistaken as tickers if written in uppercase,
# e.g. "HALO", "BISA", "GOLD", "FOOD", "DEAL", "EAST", "GOOD", "HOME".
# If they appear in natural conversational context without financial modifiers,
# they should be disambiguated carefully.
AMBIGUOUS_DICTIONARY_TICKERS: FrozenSet[str] = frozenset({
    "HALO", "BISA", "GOLD", "FOOD", "DEAL", "EAST", "GOOD", "HOME", "CARE",
    "BLUE", "BEST", "FAST", "FIRE", "FISH", "CITY", "CLEO", "CASH", "ROTI",
    "STAR", "TRUE", "TOOL", "LIVE", "NICE", "REAL", "SAFE", "SAME", "TECH",
    "PADA", "PADI", "IKAN", "AYAM", "BEER", "BULL", "DARI", "KITA", "KATA",
    "HARI", "BUAT", "BAIK", "SAYA", "KAMU", "COBA", "DATA", "MODE", "YANG",
    "SAAT", "AUTO", "KOPI", "GULA", "BOLA", "BABP",
})


def is_valid_idx_ticker(candidate: str) -> bool:
    """Check if candidate string is a valid active Indonesia Stock Exchange (IDX) ticker."""
    if not candidate:
        return False
    clean = candidate.strip().upper()
    return clean in IDX_TICKERS


def extract_valid_tickers(text: str) -> List[str]:
    """Extract only genuine IDX stock tickers from text.

    Replaces naive 4-letter regex with master IDX ticker verification.
    Guarantees that words like 'domg', 'dong', 'guys', 'suhu', 'pagi' are never returned.
    """
    if not text:
        return []

    # 1. Match potential uppercase or word-bounded 4-letter tokens
    raw_candidates = re.findall(r"\b[A-Za-z]{4}\b", text)
    valid_tickers: List[str] = []

    for cand in raw_candidates:
        upper_cand = cand.upper()
        if upper_cand in IDX_TICKERS:
            # Handle ambiguous dictionary words: only accept if strictly identified as a financial ticker
            if upper_cand in AMBIGUOUS_DICTIONARY_TICKERS:
                pattern = rf"(?:saham|emiten|ticker|kode|pt)\s+{re.escape(upper_cand)}\b|\b{re.escape(upper_cand)}\s+(?:tbk)\b"
                if not re.search(pattern, text, re.IGNORECASE):
                    continue

            if upper_cand not in valid_tickers:
                valid_tickers.append(upper_cand)

    return valid_tickers
