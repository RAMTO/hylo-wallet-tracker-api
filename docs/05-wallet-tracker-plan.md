# Hylo Wallet Tracker — Backend Plan **by Blocks**

> Scope: Backend-only, Solana mainnet-beta. Read-only services to provide prices (SOL, xSOL), balances (hyUSD, sHYUSD, xSOL), and xSOL trade history. Transport: **SSE**. Explorer: **Solscan**. Backfill: **from genesis**.

> **Price Data Source**: **DexScreener API** for SOL/USD pricing (replaces oracle complexity with simple HTTP API calls). Combined with on-chain Hylo state for xSOL price computation.

---

## Locked Decisions

- **Realtime:** SSE (`/events`, events: `price`, `balances`, `trades`, `ping 15s`)
- **RPC:** Helios (HTTP + WS); commitments: `confirmed` live / `finalized` backfill
- **Limits:** 60 req/min/IP; SSE per‑IP max 3; idle timeout 120s
- **Ref wallet (tests):** `A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g`
- **Package structure:** All services in `/internal/` (follows Go conventions for application-internal code)

---

## Shared Components

**Scope:** `/internal/common` — Shared code used across all services.  
**Modules:**

- **Config** — Environment loading, validation, Hylo constants (program IDs, mints)
- **Observability** — Structured logging, Prometheus metrics setup
- **Types** — Common Solana types (`Address`, `Signature`, `Slot`, `Commitment`)
- **Utils** — Address validation, decimal formatting, error types, Solscan URL builder

---

## Block A — Helios Connectivity

**Scope:** `/internal/solana` — HTTP + WS clients, retries/backoff, heartbeats, commitment handling.  
**Inputs:** `RPC_HTTP_URL`, `RPC_WS_URL`.  
**Outputs:** `GetAccount`, `AccountSubscribe`, `LogsSubscribe`, `GetSignaturesForAddress`, `GetTransaction`.  
**Accept:** Stable subscribe/unsubscribe; reconnect with jitter; health probe green.

**Function Usage:**

- `GetAccount` — Price engine reads Hylo state accounts (xSOL calculation data)
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
**Inputs:** Hylo program IDs, state accounts, DexScreener API for SOL/USD price.  
**Outputs:** `PriceService.XSOL()` → `{ xsol_sol, xsol_usd }`; cached results, price snapshots.  
**API:** `GET /price` → `{ sol_usd, xsol_sol, xsol_usd, updated_at }`.  
**Accept:** Matches docs/SDK reference vectors within rounding tolerance; DexScreener API reliability with fallback handling.

**Implementation Approach:**

**Phase C1: DexScreener Integration** _(2-3 days)_

- HTTP client for DexScreener API (`https://api.dexscreener.com`)
- SOL/USD price fetching with best pair selection (highest liquidity SOL/USDC or SOL/USDT pairs)
- Response caching (30-60 second TTL) and error handling with exponential backoff
- Price validation (sanity checks for reasonable SOL price range $50-$1000)

**Phase C2: Hylo State Reader** _(2-3 days)_

- Use existing `GetAccount` from Block A to read Hylo state accounts
- Parse Hylo state data for xSOL computation (leverage ratio, total SOL, total xSOL)
- Implement xSOL price formula: `xSOL_price_USD = SOL_price_USD * leverage_ratio`
- State validation and error handling following established patterns

**Phase C3: Price Service & Caching** _(1-2 days)_

- `PriceService` struct with `DexScreenerClient` and `Solana.HTTPClient` dependencies
- Combined price computation: fetch SOL price (DexScreener) + Hylo state (on-chain)
- In-memory caching with TTL expiration and cache-aside pattern
- Fallback strategies (cached prices during API failures, degraded service modes)

---

## Block D — Historical Backfill (xSOL trades, **from genesis**)

**Scope:** `/internal/indexer` — Crawl signatures, decode **Mint xSOL (BUY)** / **Redeem xSOL (SELL)** via IDL/discriminators + pre/post balances.  
**Inputs:** Hylo program IDs, wallet address, `sync_cursors`.  
**Outputs:** UPSERT rows to `xsol_trades`; `explorer_url` using Solscan; cursor persisted.  
**API:** `GET /wallet/:address/trades?cursor=<sig>&limit=25`.  
**Accept:** For reference wallet, counts & totals match Solscan over full history; re-run is idempotent.
