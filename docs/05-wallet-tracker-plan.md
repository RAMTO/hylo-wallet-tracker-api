# Hylo Wallet Tracker — Backend Plan **by Blocks**

> Scope: Backend-only, Solana mainnet-beta. Read-only services to provide prices (SOL, xSOL), balances (hyUSD, sHYUSD, xSOL), and xSOL trade history. Transport: **SSE**. Explorer: **Solscan**. Backfill: **from genesis**.

---

## Locked Decisions

- **Realtime:** SSE (`/events`, events: `price`, `balances`, `trades`, `ping 15s`)
- **RPC:** Helios (HTTP + WS); commitments: `confirmed` live / `finalized` backfill
- **Limits:** 60 req/min/IP; SSE per‑IP max 3; idle timeout 120s
- **Ref wallet (tests):** `A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g`
- **Package structure:** All services in `/internal/` (follows Go conventions for application-internal code)

---

## Block A — Helios Connectivity

**Scope:** `/internal/solana` — HTTP + WS clients, retries/backoff, heartbeats, commitment handling.  
**Inputs:** `RPC_HTTP_URL`, `RPC_WS_URL`.  
**Outputs:** `GetAccount`, `AccountSubscribe`, `LogsSubscribe`, `GetSignaturesForAddress`, `GetTransaction`.  
**Accept:** Stable subscribe/unsubscribe; reconnect with jitter; health probe green.

**Function Usage:**

- `GetAccount` — Price engine reads Pyth SOL/USD + Hylo state accounts
- `AccountSubscribe` — Indexer watches specific wallet ATAs (hyUSD, sHYUSD, xSOL) + Hylo state accounts
- `LogsSubscribe` — Indexer monitors Hylo program transaction logs
- `GetSignaturesForAddress` — Historical backfill gets wallet's xSOL trade signatures
- `GetTransaction` — Historical backfill fetches transaction details for parsing

---

## Block B — Tokens & Balances (hyUSD, sHYUSD, xSOL)

**Scope:** `/internal/tokens` — Hylo mints, decimals loader, ATA derivation, balance fetch.  
**Inputs:** Hylo mint addresses.  
**Outputs:** `GetBalances(ctx, wallet)` → `{hyusd, shyusd, xsol, slot}`.  
**API:** `GET /wallet/:address/balances`.  
**Accept:** Exact match to RPC for reference wallet (token decimals).

---

## Block C — Hylo State & xSOL Price Engine

**Scope:** `/internal/hylo` + `/internal/price` — Load Hylo state; implement formulas to compute **xSOL in SOL** and **USD**.  
**Inputs:** Hylo program IDs, state accounts, Pyth SOL/USD.  
**Outputs:** `PriceService.XSOL()` → `{ xsol_sol, xsol_usd }`; snapshots to DB, cache.  
**API:** `GET /price` (phase 2) → `{ sol_usd, xsol_sol, xsol_usd }`.  
**Accept:** Matches docs/SDK reference vectors within rounding tolerance.

---

## Block D — Historical Backfill (xSOL trades, **from genesis**)

**Scope:** `/internal/indexer` — Crawl signatures, decode **Mint xSOL (BUY)** / **Redeem xSOL (SELL)** via IDL/discriminators + pre/post balances.  
**Inputs:** Hylo program IDs, wallet address, `sync_cursors`.  
**Outputs:** UPSERT rows to `xsol_trades`; `explorer_url` using Solscan; cursor persisted.  
**API:** `GET /wallet/:address/trades?cursor=<sig>&limit=25`.  
**Accept:** For reference wallet, counts & totals match Solscan over full history; re-run is idempotent.
