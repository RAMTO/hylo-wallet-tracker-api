# AI Product Requirements Document (PRD) — Hylo Wallet Tracker (Solana)

> **Purpose:** Decision‑oriented PRD to guide implementation of a read‑only dashboard that tracks Hylo.so wallet activity and metrics on **Solana mainnet**, to be used as context inside **Cursor IDE** during development.

---

## 1) Document Control
- **Product Name:** Hylo Wallet Tracker
- **Owner / DRI:** TBD (You)
- **Stakeholders:** Backend (Go), Frontend (React), Infra/DevOps, Data/Analytics, Security
- **Status:** Draft
- **Version & History:** v0.1 — 2025‑09‑02 (initial)
- **Last Updated:** 2025‑09‑02

---

## 2) Executive Summary
- **Problem Statement:** Builders and users need a lightweight, near‑real‑time view of a wallet’s balances and xSOL trade history plus core protocol metrics (xSOL & SOL prices) for Hylo on Solana.
- **Proposed Solution:** A web app (Go API + React UI) that subscribes to Hylo programs and token accounts on **mainnet‑beta**, computes xSOL pricing from on‑chain state + Pyth SOL/USD, and displays holdings (hyUSD, sHYUSD, xSOL) and the wallet’s xSOL buy/sell transactions.
- **Who Benefits:** DeFi traders, protocol contributors, risk/ops analysts, and support.
- **Why Now:** Hylo adoption and LST yield dynamics create demand for instant visibility on xSOL exposure and stablecoin posture.
- **Success Snapshot:**
  - **Launch:** One wallet at a time, <2s P95 freshness from on‑chain change to UI update; accurate balances and prices; complete xSOL trade list for configurable lookback.
  - **+6 months:** Multi‑wallet compare, alerting, CSV export, and health/leverage panels.

---

## 3) Goals & Non‑Goals
- **Goals (Top 3):**
  1. **Accurate pricing:** Compute and show **xSOL price** (in SOL & USD) alongside **SOL/USD** in near real‑time.
  2. **Holdings view:** Show **current balances** of **hyUSD, sHYUSD, xSOL** for a provided wallet address.
  3. **Trade history:** List all **xSOL buy/sell** transactions for that wallet with timestamps, signatures, and amounts.
- **Non‑Goals:**
  - No devnet/testnet support at launch (mainnet‑beta only).
  - No execution/trading or key management; read‑only.
  - No portfolio performance analytics beyond balances & xSOL trade list.
  - No mobile app; responsive web is sufficient.

---

## 4) Users & Use Cases
- **Primary Users:**
  - DeFi traders checking exposure and fills.
  - Protocol contributors and support verifying state.
  - Analysts monitoring xSOL leverage dynamics for a wallet.
- **Key Use Cases:**
  1. Enter a wallet → see live hyUSD/sHYUSD/xSOL balances and valuations.
  2. Monitor **xSOL price** vs **SOL price** updating in near real‑time.
  3. Inspect **xSOL buys/sells** (mint/redeem) including amounts & signatures.
  4. Copy signature or link out to a block explorer.
- **Context of Use:** Browser app; single‑wallet focus; near real‑time updates over WebSockets.
- **Value Hypothesis:** Removes manual explorer sleuthing; provides one‑pane view grounded in Hylo’s own equations.

---

## 5) Experience & Guardrails
- **Entry Points:** Landing page with a wallet address input; optional wallet connect for convenience (read‑only, no signing required).
- **Primary Views:**
  - **Header bar:** SOL price (USD), xSOL price (SOL & USD), last updated indicator.
  - **Holdings card:** Token balances for hyUSD, sHYUSD, xSOL; USD valuation.
  - **xSOL Trades table:** Time, side (Buy/Sell), amount xSOL, amount paid/received (denom), signature, slot, link to explorer; pagination.
- **Interaction Rules:**
  - Auto‑subscribe on address submit; live toasts on new trades; manual refresh button.
  - All numbers formatted with fixed decimals (token‑aware) and thousand‑separators.
- **Safety & Tone:** Clear **read‑only** banner; disclaimers on data latency and oracle pricing; no financial advice.
- **Accessibility:** Keyboard‑friendly; high‑contrast mode; responsive layout.

---

## 6) Functional Requirements (acceptance criteria noted)
1. **[Must] Show SOL price (USD).** Pulled from Pyth SOL/USD; refreshed continuously. *Acceptance:* SOL price updates appear within **<2s P95** of oracle change; includes timestamp.
2. **[Must] Show xSOL price in SOL & USD.** Computed from on‑chain state using Hylo equations; USD derived via SOL/USD. *Acceptance:* Matches reference calculation within rounding tolerance; updates **<2s P95** after relevant account change.
3. **[Must] Holdings for a wallet.** Display current balances of **hyUSD, sHYUSD, xSOL** (and their USD equivalents). *Acceptance:* Matches on‑chain token account balances at the latest confirmed slot; refreshes on account change.
4. **[Must] xSOL trade history.** List all **Mint xSOL (buy)** and **Redeem xSOL (sell)** transactions for the provided wallet; include time (block time), slot, signature, side, amount xSOL, counter‑asset and amount (if inferable), and link to explorer. *Acceptance:* For a configured lookback window, table includes all matching signatures; new entries appear within **<5s P95** from confirmation.
5. **[Should] Error & empty states.** Wallet with no Hylo activity shows zero balances and empty table; network/oracle/RPC errors render retriable banners.
6. **[Should] Performance & pagination.** Trades table paginates by time (infinite scroll or pages of 25 rows) with server‑side cursors.
7. **[Could] CSV export.** Export trades and balances snapshots for the active wallet.

**Assumptions & Config:**
- Default trade backfill **lookback = 12 months**, configurable (including "from genesis").
- Only **one wallet** tracked at a time in the UI; switching wallets tears down/re‑subscribes.
- **Commitment:** `confirmed` for live view; `finalized` for backfills.

---

## 7) Architecture & Tech Stack (Conceptual)
- **Overall:** Event‑driven read‑only indexer (Go) + REST/WS API (Go) + React front end (PNPM) + PostgreSQL storage.
- **Components:**
  1. **Indexer (Go):**
     - **WebSocket subscriptions** via Helios RPC to: Hylo program logs, relevant program accounts (e.g., HYLO state, token mints/supplies) and the user’s token accounts for hyUSD/sHYUSD/xSOL.
     - **Historical backfill** using `getSignaturesForAddress` and transaction parsing to identify **MintLevercoin / RedeemLevercoin** instructions.
     - **Price engine:** Fetch Pyth SOL/USD; compute **hyUSD NAV (in SOL)** and **xSOL NAV (in SOL)**; convert to USD; compute collateral ratio & effective leverage (for future panels).
  2. **API Gateway (Go, e.g., Fiber/Gin):**
     - Endpoints: `/price`, `/wallet/:address/balances`, `/wallet/:address/trades` (with cursor params), `/health`.
     - **Server‑sent events or WS** stream for live price/balance updates.
  3. **DB (PostgreSQL):**
     - Tables (conceptual): `price_snapshots`, `wallet_balances`, `xsol_trades`, `sync_cursors`, `ingest_failures`.
  4. **Frontend (React + Vite, PNPM):**
     - Hooks for live streams; components for metrics header, holdings card, and trades table; copy‑to‑clipboard and explorer links.
- **Dependencies:** Helios RPC (HTTP + WS), Pyth price feed, Hylo program IDs & mints.
- **Observability:** Structured logs, Prometheus metrics, basic traces; on‑screen timing badges for freshness.

---

## 8) Data & Calculations
- **On‑chain programs & mints (mainnet):** Exchange program, Stability Pool program; token mints for hyUSD, sHYUSD, xSOL.
- **Key derived values:**
  - SOL/USD (from Pyth)
  - hyUSD NAV (SOL), xSOL NAV (SOL), xSOL price (USD), collateral ratio, effective leverage.
- **Holdings:** Fetch ATAs for hyUSD, sHYUSD, xSOL for the provided wallet; show balances + USD equivalents.
- **Trade identification:** Parse transactions where the wallet is signer or source/target in **Mint xSOL** or **Redeem xSOL** instructions; derive side (Buy/Sell) and amounts from instruction args and pre/post token balances. Use Anchor IDL decoding or instruction discriminators.

---

## 9) Interfaces (API draft)
- `GET /price` → `{ sol_usd, xsol_sol, xsol_usd, updated_at }`
- `GET /wallet/:address/balances` → `{ hyusd, shyusd, xsol, values_usd, updated_at }`
- `GET /wallet/:address/trades?cursor=<sig>&limit=25` → `[{ ts, slot, sig, side, amount_xsol, amount_counter, counter_symbol }]`
- `GET /health` → readiness/liveness.
- **Realtime:** `/events` SSE or `/ws` with topics: `price`, `balances`, `trades`.

---

## 10) Performance, Cost, SLOs, Rollout & Metrics
- **Latency Targets:**
  - Price & balances **<2s P95** from on‑chain update to UI render.
  - New trade rows **<5s P95** from confirmation.
- **Throughput/Scale:** Single-wallet focus; <1 req/sec average; spikes on backfill; keep RPC under provider rate limits.
- **Availability:** 99.5% target (read‑only).
- **Cost Guardrails:** Pyth access via on‑chain read; minimize expensive historical RPC by caching cursors and bounding lookbacks.
- **Rollout Plan:**
  1) Alpha (internal): hardcoded wallet; 2) Beta: address input + explorer links; 3) GA: CSV export, env/config hardening.
- **Success Metrics:** Accuracy vs reference, time‑to‑freshness, UI error rate, user‑reported correctness, and export usage.

---

## 11) Security & Privacy
- Read‑only application; no private keys stored.
- Sanitized address inputs; rate‑limit public endpoints; CORS rules.
- Only public, on‑chain data plus derived metrics; no PII.

---

## 12) Open Questions / Assumptions
- **Trade backfill window:** defaulting to **12 months**, but can be set to **genesis**. (Config flag)
- **Decimals & rounding:** Display xSOL to 6–9 decimals (confirm token decimals at integration time); SOL/USD to 2 decimals.
- **Explorer links:** Which explorer as default (SolanaFM/Explorer/Solana Beach)?

---

## 13) Implementation Notes (for Cursor)
- **Backend (Go):**
  - Packages: `github.com/gagliardetto/solana-go` (tx parsing), `nhooyr.io/websocket` or Gorilla for WS/SSE, `pgx` for Postgres.
  - Use Helios WS for `programSubscribe`, `logsSubscribe`, and `accountSubscribe` as applicable.
  - Pyth: read SOL/USD EMA price account; convert to float with confidence handling and staleness checks.
  - Compute NAVs and xSOL price per Hylo formulas; cache in‑memory with TTL.
- **Frontend (React):**
  - State via TanStack Query + SSE/WS; PNPM for install/build.
  - Components: `PriceHeader`, `HoldingsCard`, `TradesTable` with infinite scroll.
- **Database:**
  - `price_snapshots(ts, slot, sol_usd, xsol_sol, xsol_usd)`
  - `wallet_balances(wallet, ts, hyusd, shyusd, xsol)` (latest per wallet)
  - `xsol_trades(sig PK, slot, ts, wallet, side, amount_xsol, counter_symbol, amount_counter)`
- **Testing:** Simulate against a short block range; golden tests for price math vs precomputed vectors.

---

## 14) Glossary
- **hyUSD:** Hylo stablecoin.
- **sHYUSD:** Staked hyUSD (LP token from stability pool).
- **xSOL:** Leveraged SOL exposure issued by Hylo.
- **Collateral Ratio, NAV, Effective Leverage:** Protocol health & exposure metrics derived from reserves and supplies.
