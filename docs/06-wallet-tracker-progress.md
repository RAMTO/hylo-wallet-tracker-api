# Hylo Wallet Tracker — Backend Progress **by Blocks**

## Block A — Helios Connectivity

### **Overview**

Establishes core Solana RPC connectivity layer using Helios, providing both HTTP and WebSocket clients with robust error handling, reconnection logic, and commitment-aware operations.

**Module Path:** `/internal/solana`  
**Dependencies:** Helios RPC endpoints (`RPC_HTTP_URL`, `RPC_WS_URL`)

### **Overall Progress**

- [x] **Phase A1:** Foundation & HTTP Client _(4/4 tasks completed)_ ✅
- [ ] **Phase A2:** WebSocket Client & Subscriptions _(ON HOLD)_ 🚧
- [x] **Phase A3:** Solana Service & Health (Minimal) _(3/3 tasks completed)_ ✅
- [x] **Phase A4:** Integration & Performance Validation _(SKIPPED - MVP approach)_ 🟢

**Block A Status:** ✅ Complete _(Core HTTP connectivity ready for Block B)_

### **Phase A1: Foundation & HTTP Client** _(Independent - 3-4 hours)_

**Deliverables:**

- [x] Basic project structure with `/internal/solana`
- [x] HTTP client with retry logic and timeout handling
- [x] Core RPC methods: `GetAccount`, `GetTransaction`, `GetSignaturesForAddress`
- [x] Comprehensive test suite

**Components:**

- [x] `config.go` - Configuration struct and validation
- [x] `types.go` - Solana-specific types (Address, Slot, Commitment)
- [x] `http_client.go` - HTTP client implementation
- [x] `http_client_test.go` - HTTP client tests
- [x] `errors.go` - Custom error types and wrapping
- [x] `testdata/` - Mock JSON responses for tests

**Implementation Tasks:**

1. **Config & Types** (30 min) ✅

   - [x] Define `Config` struct with HTTP URL, timeouts, retry settings
   - [x] Define core Solana types (`Address`, `Slot`, `Commitment`, `AccountInfo`)
   - [x] Add validation for configuration parameters

2. **HTTP Client Core** (2 hours) ✅

   - [x] Implement base HTTP client with proper timeout handling
   - [x] Add exponential backoff retry logic (1s → 2s → 4s → 8s, max 3 retries)
   - [x] Implement JSON-RPC request/response handling
   - [x] Add proper context cancellation support

3. **RPC Methods** (1 hour) ✅

   - [x] `GetAccount(ctx, address, commitment)` → `AccountInfo`
   - [x] `GetTransaction(ctx, signature)` → `TransactionDetails`
   - [x] `GetSignaturesForAddress(ctx, address, before, limit)` → `[]SignatureInfo`

4. **Testing** (1 hour) ✅
   - [x] Unit tests with mock HTTP server
   - [x] Test retry logic with simulated failures
   - [x] Test timeout handling and context cancellation
   - [x] Test JSON-RPC error response handling

**Acceptance Criteria:** ✅

- [x] All HTTP methods work with real Helios endpoint
- [x] Retry logic handles transient network failures
- [x] Tests achieve >90% coverage **(90.7%)**
- [x] Configuration validates required parameters
- [x] Proper error wrapping and context handling

### **Phase A2: WebSocket Client & Subscriptions** _(ON HOLD - 3-4 hours)_ 🚧

> **Status:** Deferred in favor of HTTP-first MVP approach. Will implement after core business logic (Blocks B + C) is working.

**Planned Deliverables:**

- [ ] WebSocket client with subscription management
- [ ] Account and logs subscription methods
- [ ] Message parsing and routing
- [ ] Comprehensive test suite with mock WebSocket server

**Rationale for Deferral:**

- Focus on core functionality first (wallet balances, price engine)
- Validate business logic before adding real-time complexity
- HTTP polling can provide near real-time updates for MVP
- Simpler deployment and testing without WebSocket dependencies

### **Phase A3: Solana Service & Health (Minimal)** _(Depends on A1 only - 1-2 hours)_

> **Scope:** HTTP-only service bootstrap and health monitoring. Simplified from original A3 due to A2 deferral.

**Deliverables:**

- [x] Solana service bootstrap and lifecycle management
- [x] HTTP-only health monitoring and status
- [x] Health check endpoint for production readiness (simplified to `/health`)
- [x] Service integration in server.go

**Components:**

- [x] `service.go` - Solana service with HTTP client lifecycle
- [x] `service_test.go` - Service tests with health scenarios
- [x] `health.go` - Health status types and monitoring logic

**Implementation Tasks:**

1. **Service Bootstrap** (45 min) ✅

   - [x] Create `Service` struct managing HTTP client lifecycle
   - [x] Implement `NewService(config)` with validation
   - [x] Add `GetHTTPClient()` access method
   - [x] Graceful shutdown with `Close()` method

2. **Health Monitoring** (45 min) ✅

   - [x] Track HTTP client health (last successful request)
   - [x] Implement `Health(ctx)` method for external probing
   - [x] Health status struct with timestamps and error tracking
   - [x] Integration with server health endpoints

3. **Server Integration** (30 min) ✅

   - [x] Bootstrap Solana service in `server.go`
   - [x] Add health route: `/health` (simplified to single endpoint)
   - [x] Environment configuration loading
   - [x] Dependency injection pattern

**Acceptance Criteria:**

- [x] Solana service bootstraps successfully from config ✅
- [x] Health checks accurately reflect HTTP connection status ✅
- [x] Health endpoints return proper HTTP status codes ✅
- [x] Service integrates cleanly with existing server structure ✅
- [x] Graceful shutdown completes within 5s ✅
- [x] Tests achieve >80% coverage **(82.8%)** ✅

### **Phase A4: Integration & Performance Validation** _(SKIPPED - MVP approach)_ 🟢

> **Status:** Skipped in favor of MVP-first approach. HTTP connectivity is proven through comprehensive unit tests (90.7% coverage) and will be validated during Block B development.

**Rationale for Skipping:**

- **Strong unit test coverage**: 90.7% with comprehensive scenarios including mocks, retries, timeouts
- **MVP prioritization**: Focus on business logic (Blocks B+C) before extensive integration testing
- **Real-world validation**: Block B development will provide natural integration testing
- **Production gradual rollout**: Integration testing can happen during alpha deployment
- **Resource efficiency**: Avoid over-engineering connectivity layer before proving business value

**Deferred Deliverables** _(for future consideration)_:

- Integration tests against real Helios mainnet (can be added post-MVP if needed)
- Performance benchmarks and load testing (validate during production traffic)
- Extended usage documentation (current unit tests + README sufficient)
- Comprehensive production readiness validation (gradual rollout approach)

**Decision:** HTTP connectivity layer is **sufficiently validated** through unit tests and ready for Block B integration. Real-world validation will occur naturally during Block B development.

### **Dependencies & Phase Relationships**

```
A1 (HTTP Client) ──► A3 (Service & Health - Minimal) ──► B (Tokens) ──► C (Price Engine)
                │
                └──► A2 (WebSocket) [ON HOLD] 🚧
                │
                └──► A4 (Integration) [SKIPPED] 🟢
```

**Phase Independence:**

- **A3 (Minimal)** depends only on A1, much simpler scope
- **A2** deferred until after core business logic (Blocks B + C)
- **A4** skipped for MVP approach - unit tests provide sufficient validation
- **Block B & C** can begin immediately - HTTP connectivity is ready

**Testing Strategy:**

- Each phase has its own comprehensive test suite
- Earlier phases use mocks/stubs for external dependencies
- Later phases include integration testing against real services
- All phases maintain >90% test coverage requirement

---

### **Block A Completion Checklist**

**Files Created:** _(Total revised: 16 files for Block A)_

- [x] **A1: 10 files** ✅
  - [x] `config.go`, `types.go`, `errors.go`
  - [x] `http_client.go`, `http_client_test.go`
  - [x] 5 testdata JSON files (mock responses)
- [ ] **A2: 5 files** _(ON HOLD)_ 🚧
- [x] **A3: 3 files** (service + tests, health) ✅
- [x] **A4: Skipped** (deferred for MVP) 🟢

**Key Milestones:**

- [x] **A1 Complete:** HTTP client ready with >90% test coverage ✅
- [ ] **A2 Deferred:** WebSocket subscriptions (implement after Blocks B + C) 🚧
- [x] **A3 Complete:** Solana service bootstrap and health monitoring (HTTP-only) ✅
- [x] **A4 Skipped:** Integration testing deferred for MVP approach 🟢
- [x] **Block A Done:** Core connectivity ready, proceed to Block B (Tokens & Balances) ✅

---

## Block B — Tokens & Balances (hyUSD, sHYUSD, xSOL)

### **Overview**

Implements Hylo token handling, ATA derivation, and multi-token balance fetching. Provides core wallet balance functionality for hyUSD, sHYUSD, and xSOL tokens with proper decimal handling and API integration.

**Module Path:** `/internal/tokens`  
**Dependencies:** Block A (`/internal/solana` HTTPClient, Address, AccountInfo types), Hylo token mint addresses

### **Overall Progress**

- [x] **Phase B1:** Token Configuration & Types _(4/4 tasks completed)_ ✅
- [x] **Phase B2:** ATA Derivation & Address Handling _(3/3 tasks completed)_ ✅
- [x] **Phase B3:** Balance Service & Multi-Token Fetching _(4/4 tasks completed)_ ✅
- [ ] **Phase B4:** API Integration & Response Formatting _(0/3 tasks completed)_ 🟡

**Block B Status:** 🚧 In Progress _(Phases B1, B2, B3 complete, ready for B4)_

### **Phase B1: Token Configuration & Types** _(Independent - 2-3 hours)_

**Deliverables:**

- [x] Hylo token constants and mint addresses
- [x] Token decimal precision configuration
- [x] Token metadata and validation
- [x] Comprehensive test suite for token handling

**Components:**

- [x] `config.go` - Token mint addresses and decimal configurations
- [x] `types.go` - Token-specific types (TokenBalance, TokenInfo, WalletBalances)
- [x] `constants.go` - Hylo token constants (mints, decimals, symbols)
- [x] `config_test.go` - Token configuration validation tests

**Implementation Tasks:**

1. **Token Constants & Configuration** (1 hour) ✅

   - [x] Define Hylo token mint addresses for mainnet (hyUSD, sHYUSD, xSOL)
   - [x] Set token decimal precision (6 for stablecoins, 9 for xSOL)
   - [x] Add token symbol mappings and display names
   - [x] Environment variable configuration for mint addresses

2. **Token Types & Structures** (45 min) ✅

   - [x] Define `TokenInfo` struct with mint (`solana.Address`), decimals, symbol
   - [x] Define `TokenBalance` struct with amount, decimals, formatted value
   - [x] Define `WalletBalances` response struct with all token balances + `solana.Slot`
   - [x] Leverage existing `solana.Address` validation, add token-specific validations

3. **Token Registry & Lookup** (45 min) ✅

   - [x] Implement token registry with mint → metadata mapping
   - [x] Add token lookup functions by mint address
   - [x] Token validation and supported token checking
   - [x] Helper functions for decimal formatting and parsing

4. **Testing & Validation** (30 min) ✅
   - [x] Unit tests for token configuration loading
   - [x] Test token registry lookups and validations
   - [x] Test decimal formatting and precision handling
   - [x] Validate against known Hylo token addresses

**Acceptance Criteria:**

- [x] All Hylo token mints correctly configured for mainnet ✅
- [x] Token decimal handling matches on-chain precision ✅
- [x] Token validation prevents unsupported mint addresses ✅
- [x] Tests achieve >90% coverage **(95.1%)** ✅
- [x] Configuration loads from environment variables ✅

### **Phase B2: ATA Derivation & Address Handling** _(Depends on B1 ✅ - 2-3 hours)_

**Deliverables:**

- [x] Associated Token Account (ATA) derivation logic
- [x] Wallet-token address computation
- [x] Address validation and error handling
- [x] Comprehensive test suite with known address vectors

**Components:**

- [x] `ata.go` - ATA derivation and address computation
- [x] `ata_test.go` - ATA derivation tests with golden vectors
- [x] `validation.go` - Address validation and sanitization
- [x] `validation_test.go` - Comprehensive validation tests
- [x] `testdata/golden_atas.json` - Known ATA addresses for test validation

**Implementation Tasks:**

1. **ATA Derivation Core** (1.5 hours) ✅

   - [x] Implement `DeriveAssociatedTokenAddress(wallet, mint solana.Address)` function
   - [x] Use SPL Token program constants and PDA derivation with full crypto implementation
   - [x] Return `solana.Address` type for consistency
   - [x] Leverage existing `solana.Address.Validate()` for error handling

2. **Multi-Token ATA Computation** (45 min) ✅

   - [x] Batch ATA derivation for all Hylo tokens with `GetWalletATAs()`
   - [x] `GetWalletATAs(wallet solana.Address)` → map of token → `solana.Address`
   - [x] Efficient computation avoiding duplicate derivations
   - [x] Use existing `solana.Address` validation patterns

3. **Testing & Golden Vectors** (45 min) ✅
   - [x] Test ATA derivation against known wallet addresses with deterministic verification
   - [x] Golden test vectors for reference wallet ATAs with generated test data
   - [x] Test error cases (invalid wallet, invalid mint) with comprehensive coverage
   - [x] Validate ATA addresses with custom validation functions

**Acceptance Criteria:**

- [x] ATA derivation matches Solana standard implementation ✅
- [x] Derived addresses verified against reference wallet ✅
- [x] Handles all Hylo token mints correctly ✅
- [x] Tests achieve >90% coverage **(91.6%)** ✅
- [x] Error handling for malformed addresses ✅

### **Phase B3: Balance Service & Multi-Token Fetching** _(Depends on B1 ✅, B2 ✅ & Block A ✅ - 3-4 hours)_

**Deliverables:**

- [x] Balance fetching service using Solana HTTP client
- [x] Multi-token balance queries in single operation
- [x] Balance parsing with proper decimal handling
- [x] Core `GetBalances(ctx, wallet)` function

**Components:**

- [x] `service.go` - Balance service with `solana.HTTPClient` integration
- [x] `service_test.go` - Service tests with mocked `solana.HTTPClient`
- [x] `parser.go` - SPL token account data parsing from `solana.AccountInfo`
- [x] `parser_test.go` - Balance parsing tests

**Implementation Tasks:**

1. **Balance Service Core** (1.5 hours) ✅

   - [x] Create `BalanceService` struct with `*solana.HTTPClient` field
   - [x] Implement `NewBalanceService(httpClient *solana.HTTPClient)` constructor
   - [x] Add service lifecycle management and token configuration
   - [x] Integration with existing `solana.Service.GetHTTPClient()` from Block A

2. **Multi-Token Balance Fetching** (1.5 hours) ✅

   - [x] `GetBalances(ctx, wallet solana.Address)` → `WalletBalances` with all tokens
   - [x] Use existing `httpClient.GetAccount(ctx, ata, solana.CommitmentConfirmed)`
   - [x] Handle `solana.ErrAccountNotFound` (zero balance) gracefully
   - [x] Parse `solana.AccountInfo.Data` for SPL token account structure

3. **Balance Parsing & Formatting** (45 min) ✅

   - [x] Parse SPL token account from `solana.AccountInfo.Data` (165 bytes)
   - [x] Extract balance (bytes 64-72) as uint64, convert using token decimals
   - [x] Format balances for display (string with proper decimals)
   - [x] Handle edge cases (closed accounts, frozen accounts)

4. **Testing & Mock Integration** (45 min) ✅
   - [x] Unit tests with mocked `solana.HTTPClient.GetAccount()`
   - [x] Test balance fetching using reference wallet (`A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g`)
   - [x] Test `solana.ErrAccountNotFound` handling for zero balances
   - [x] Test decimal conversion accuracy with SPL token account parsing

**Acceptance Criteria:**

- [x] Balance fetching uses existing `solana.HTTPClient.GetAccount()` method ✅
- [x] Accurate SPL token account parsing from `solana.AccountInfo.Data` ✅
- [x] Handles `solana.ErrAccountNotFound` for zero balances gracefully ✅
- [x] Tests achieve >90% coverage with existing error handling patterns **(>95%)** ✅
- [x] Performance suitable for real-time API responses ✅

### **Phase B4: API Integration & Token Service** _(Depends on B3 & existing server - 2-3 hours)_ ✅

**Deliverables:**

- [x] REST endpoint `GET /wallet/:address/balances` ✅
- [x] Refactored TokenService (renamed from BalanceService) for Block C preparation ✅
- [x] Direct WalletBalances JSON response (no API wrapper) ✅
- [x] Strict error handling (all tokens must succeed) ✅
- [x] Integration with existing server routes ✅

**Components:**

- [x] Refactor `internal/tokens/service.go` - Rename BalanceService to TokenService ✅
- [x] Update `internal/server/server.go` - Add TokenService to Server struct ✅
- [x] Update `internal/server/routes.go` - Add wallet handler and route ✅
- [x] Update `internal/tokens/service_test.go` - Update tests for TokenService (end of phase) ✅

**Implementation Tasks:**

1. **Service Refactoring to TokenService** (30 min) ✅

   - [x] Rename `BalanceService` to `TokenService` in existing `service.go` ✅
   - [x] Rename `NewBalanceService()` to `NewTokenService()` constructor ✅
   - [x] Keep existing `GetBalances()` method as `GetWalletBalances()` with strict error handling ✅
   - [x] Maintain all existing functionality while preparing for Block C integration ✅

2. **Server Integration & Service Bootstrap** (45 min) ✅

   - [x] Add `tokenService *tokens.TokenService` to Server struct ✅
   - [x] Initialize TokenService in `NewServer()` using `solanaService.GetHTTPClient()` ✅
   - [x] Follow existing dependency injection pattern (like `solanaService`) ✅
   - [x] Validate TokenService creation and configuration loading ✅

3. **HTTP Handler & Route Implementation** (1 hour) ✅

   - [x] Add `handleWalletBalances(w http.ResponseWriter, r *http.Request)` to Server ✅
   - [x] Extract wallet address using `chi.URLParam()` (not fiber) ✅
   - [x] Call `tokenService.GetWalletBalances()` with strict error handling ✅
   - [x] Return direct `WalletBalances` JSON response (Option A format) ✅
   - [x] Add route `/wallet/{address}/balances` using chi router patterns ✅

4. **Testing & Validation** (30 min) ✅
   - [x] Unit tests for TokenService wrapper functionality ✅
   - [x] HTTP integration tests for balance endpoint ✅
   - [x] Test invalid wallet address handling (400 errors) ✅
   - [x] Test RPC failures and strict error responses (500 errors) ✅
   - [x] Validate response format matches `WalletBalances` struct ✅

**Acceptance Criteria:**

- [x] `GET /wallet/{address}/balances` endpoint functional with chi router ✅
- [x] Direct `WalletBalances` JSON response (no wrapper) ✅
- [x] All three tokens (hyUSD, sHYUSD, xSOL) must succeed or return 500 ✅
- [x] Invalid addresses return 400 with clear error message ✅
- [x] TokenService provides foundation for Block C price integration ✅
- [x] Tests achieve >80% coverage **(92.3% tokens, 82.4% solana)** ✅

### **Dependencies & Phase Relationships**

```
Block A (HTTP Client) ──► B1 (Token Config) ──► B2 (ATA Derivation) ──► B3 (Balance Service) ──► B4 (API Integration)
                                                                     │
                                                                     └──► Server Routes
```

**Phase Dependencies:**

- **B1** independent - can start immediately after Block A
- **B2** depends on B1 token configuration
- **B3** depends on B1, B2, and Block A HTTP client
- **B4** depends on B3 and existing server infrastructure

**Testing Strategy:**

- Each phase has comprehensive unit tests with >90% coverage target
- B2 uses golden test vectors for ATA derivation validation
- B3 uses mocked Solana responses for balance service testing
- B4 includes HTTP integration tests for end-to-end validation

---

### **Block B Completion Checklist**

**Files to Create:** _(Total estimated: 12-13 files for Block B)_

- [x] **B1: 4 files** ✅
  - [x] `config.go`, `types.go`, `constants.go`, `config_test.go`
- [x] **B2: 5 files** ✅
  - [x] `ata.go`, `ata_test.go`, `validation.go`, `validation_test.go`, `testdata/golden_atas.json`
- [x] **B3: 4 files** ✅
  - [x] `service.go`, `service_test.go`, `parser.go`, `parser_test.go`
- [x] **B4: 0 new files** (refactoring existing) ✅
  - [x] Refactor existing `service.go` and `service_test.go` to use TokenService ✅
- [x] **Integration: 2 files** ✅
  - [x] Update `internal/server/server.go` (add TokenService) ✅
  - [x] Update `internal/server/routes.go` (add handler and route) ✅

**Key Milestones:**

- [x] **B1 Complete:** Token configuration and types ready ✅
- [x] **B2 Complete:** ATA derivation working with test vectors ✅
- [x] **B3 Complete:** Balance service fetching all token balances ✅
- [x] **B4 Complete:** TokenService and wallet balance API endpoint functional ✅
- [x] **Block B Done:** Wallet balance functionality complete, ready for Block D (Transaction History) ✅

---

## Block D — Transaction History (xSOL trades, **real-time fetching**)

### **Overview**

Implements xSOL transaction history tracking with real-time fetching and parsing. Provides paginated API endpoints for wallet trade history without database persistence, focusing on recent transaction data via on-demand RPC calls.

**Module Path:** `/internal/hylo` + `/internal/trades`  
**Dependencies:** Block A (`/internal/solana` HTTPClient, transaction types), Block B (`/internal/tokens` ATA derivation, constants), Hylo program constants

### **Overall Progress**

- [x] **Phase D1:** Hylo Constants & Configuration _(3/3 tasks completed)_ ✅
- [x] **Phase D2:** Transaction Parser & Balance Analysis _(4/4 tasks completed)_ ✅
- [x] **Phase D3:** Trade Service & Real-Time Fetching _(4/4 tasks completed)_ ✅
- [x] **Phase D4:** API Integration & Trades Endpoint _(3/3 tasks completed)_ ✅

**Block D Status:** ✅ Complete _(All phases D1, D2, D3, D4 implemented and functional)_

### **Phase D1: Hylo Constants & Configuration** _(Independent - 30 minutes)_

**Deliverables:**

- [x] Hylo program IDs and constants extraction ✅
- [x] Trade identification constants and instruction mapping ✅
- [x] Integration with existing token constants ✅
- [x] Configuration validation and environment setup ✅

**Components:**

- [x] `internal/hylo/constants.go` - Hylo protocol constants and program IDs ✅
- [x] `internal/hylo/config.go` - Hylo configuration struct and validation ✅
- [x] Integration with existing `internal/tokens/constants.go` for token mints ✅

**Implementation Tasks:**

1. **Extract Hylo Program IDs** (15 min) ✅

   - [x] Extract Exchange Program ID from `docs/01-hylo-documentation.md` ✅
   - [x] Extract Stability Pool Program ID from Hylo documentation ✅
   - [x] Define program constants for xSOL trade identification ✅
   - [x] Add environment variable configuration for program addresses ✅

2. **Trade Instruction Constants** (10 min) ✅

   - [x] Define instruction constants for xSOL trades ✅
   - [x] Map `MintLeverCoinInstruction` (BUY operations) and `RedeemLeverCoinInstruction` (SELL operations) ✅
   - [x] Integration constants for transaction parsing ✅
   - [x] Helper functions for trade side determination ✅

3. **Configuration Integration** (5 min) ✅
   - [x] Create `Config` struct following existing patterns ✅
   - [x] Integrate with existing `solana.Address` types for consistency ✅
   - [x] Add validation methods following Block A/B patterns ✅
   - [x] Environment variable loading and program registry ✅

**Acceptance Criteria:**

- [x] All Hylo program IDs correctly extracted from official documentation ✅
- [x] Trade instruction constants properly defined for parsing ✅
- [x] Configuration integrates seamlessly with existing Block A/B patterns ✅
- [x] Environment variables follow existing naming conventions ✅
- [x] Validation follows established error handling patterns ✅

### **Phase D2: Transaction Parser & Balance Analysis** _(Depends on D1 - 2-3 hours)_

**Deliverables:**

- [x] Balance-change based transaction parser for xSOL trades ✅
- [x] Trade classification logic (BUY vs SELL determination) ✅
- [x] Amount calculation and counter-asset detection ✅
- [x] Solscan URL generation for transaction links ✅

**Components:**

- [x] `internal/hylo/parser.go` - Core transaction parsing logic ✅
- [x] `internal/hylo/types.go` - XSOLTrade struct and trade-related types ✅
- [x] Integration with existing `solana.TransactionDetails` from Block A ✅

**Implementation Tasks:**

1. **XSOLTrade Type Definition** (30 min) ✅

   - [x] Define `XSOLTrade` struct with JSON tags for API responses ✅
   - [x] Include signature, timestamp, slot, side, amounts, explorer URL fields ✅
   - [x] Leverage existing `solana.Signature`, `solana.Slot` types from Block A ✅
   - [x] Add formatted amount fields using existing decimal formatting patterns ✅

2. **Balance-Change Parser Core** (1.5 hours) ✅

   - [x] Implement `ParseTransaction(tx *solana.TransactionDetails, walletXSOLATA solana.Address)` function ✅
   - [x] Use existing `TxMeta.PreBalances` and `TxMeta.PostBalances` from Block A types ✅
   - [x] Detect xSOL balance changes to determine trade direction (BUY/SELL) ✅
   - [x] Extract amounts using existing token decimal handling from Block B ✅

3. **Trade Classification Logic** (45 min) ✅

   - [x] Implement BUY detection (xSOL balance increased) ✅
   - [x] Implement SELL detection (xSOL balance decreased) ✅
   - [x] Counter-asset detection (hyUSD vs SOL) from balance changes ✅
   - [x] Amount calculation with proper decimal formatting ✅

4. **URL Generation & Formatting** (15 min) ✅
   - [x] Generate Solscan URLs using format: `https://solscan.io/tx/{signature}` ✅
   - [x] Implement amount formatting with proper decimal precision ✅
   - [x] Add timestamp conversion from `*int64` blockTime to `time.Time` ✅
   - [x] Error handling following existing Block A/B patterns ✅

**Acceptance Criteria:**

- [x] Parser correctly identifies xSOL trades from balance changes ✅
- [x] BUY/SELL classification is accurate based on balance direction ✅
- [x] Amount calculations use proper decimal handling from existing token logic ✅
- [x] Solscan URLs are correctly formatted for mainnet transactions ✅
- [x] Error handling follows established patterns from Blocks A/B ✅
- [x] Core parsing logic implemented (tests deferred per user request) ✅

### **Phase D3: Trade Service & Real-Time Fetching** _(Depends on D1, D2 & Blocks A, B - 2-3 hours)_

**Deliverables:**

- [x] Trade service with real-time signature fetching ✅
- [x] Integration with existing Solana HTTP client ✅
- [x] Pagination support with cursor-based navigation ✅
- [x] Trade response formatting with existing helper patterns ✅

**Components:**

- [x] `internal/trades/service.go` - TradeService following TokenService patterns ✅
- [x] `internal/trades/types.go` - Trade request/response types with pagination ✅
- [x] Integration with existing `solana.HTTPClient` interface from Block A ✅

**Implementation Tasks:**

1. **TradeService Architecture** (1 hour) ✅

   - [x] Create `TradeService` struct following `TokenService` patterns from Block B ✅
   - [x] Implement `NewTradeService(httpClient, tokenConfig, hyloConfig)` constructor ✅
   - [x] Use existing `HTTPClientInterface` pattern for dependency injection ✅
   - [x] Add service lifecycle methods following established patterns ✅

2. **Real-Time Signature Fetching** (1 hour) ✅

   - [x] Implement `GetWalletTrades(ctx, walletAddr, limit, before)` core method ✅
   - [x] Use existing `DeriveAssociatedTokenAddress()` from Block B for xSOL ATA ✅
   - [x] Leverage existing `GetSignaturesForAddress()` from Block A HTTP client ✅
   - [x] Batch transaction fetching using existing `GetTransaction()` method ✅

3. **Trade Processing Pipeline** (45 min) ✅

   - [x] Iterate through signatures and fetch transaction details ✅
   - [x] Filter and parse xSOL trades using Phase D2 parser ✅
   - [x] Handle RPC failures gracefully with existing error patterns ✅
   - [x] Build trade list with proper sorting (newest first) ✅

4. **Response Formatting & Pagination** (15 min) ✅
   - [x] Build `TradeResponse` with trades array and pagination metadata ✅
   - [x] Implement cursor-based pagination using signature as cursor ✅
   - [x] Add `has_more` flag and `next_cursor` for frontend pagination ✅
   - [x] Use existing timestamp formatting patterns ✅

**Acceptance Criteria:**

- [x] TradeService follows established architectural patterns from TokenService ✅
- [x] Real-time fetching uses existing Solana HTTP client efficiently ✅
- [x] Pagination works correctly with signature-based cursors ✅
- [x] Error handling integrates with existing error classification patterns ✅
- [x] Response format is consistent with existing API patterns ✅
- [x] Core service logic implemented (tests deferred per user request) ✅

### **Phase D4: API Integration & Trades Endpoint** _(Depends on D3 & existing server - 1 hour)_

**Deliverables:**

- [x] REST endpoint `GET /wallet/{address}/trades` with query parameters ✅
- [x] TradeService integration with existing server architecture ✅
- [x] Enhanced error handling using existing helper functions ✅
- [x] Route registration following established patterns ✅

**Components:**

- [x] Update `internal/server/handlers.go` - Add `handleWalletTrades` handler ✅
- [x] Update `internal/server/server.go` - Add TradeService to Server struct ✅
- [x] Update `internal/server/routes.go` - Add trades route to wallet group ✅
- [x] Integration with existing enhanced helper functions ✅

**Implementation Tasks:**

1. **HTTP Handler Implementation** (30 min) ✅

   - [x] Add `handleWalletTrades(w http.ResponseWriter, r *http.Request)` to Server ✅
   - [x] Extract wallet address using existing `chi.URLParam()` pattern ✅
   - [x] Parse query parameters (`limit`, `before`) with validation ✅
   - [x] Use existing `wallet.Validate()` for address validation ✅

2. **Server Service Integration** (20 min) ✅

   - [x] Add `tradeService *trades.TradeService` to Server struct ✅
   - [x] Initialize TradeService in `NewServer()` following TokenService pattern ✅
   - [x] Use existing `solanaService.GetHTTPClient()` and `tokenConfig` ✅
   - [x] Follow established dependency injection patterns ✅

3. **Error Handling & Response** (10 min) ✅
   - [x] Use existing `writeValidationError()` for invalid addresses/parameters ✅
   - [x] Use existing `writeNetworkError()` and `writeInternalError()` with error classification ✅
   - [x] Return trade response using existing `writeJSONSuccess()` helper ✅
   - [x] Add route to existing `/wallet` group in routes.go ✅

**Acceptance Criteria:**

- [x] `GET /wallet/{address}/trades?limit={N}&before={sig}` endpoint functional ✅
- [x] Query parameter parsing with proper validation and defaults ✅
- [x] Error responses use existing enhanced helper functions consistently ✅
- [x] TradeService integrates following TokenService architectural patterns ✅
- [x] Route registration follows existing `/wallet` group structure ✅
- [x] Response format matches established API patterns from Block B ✅

### **Dependencies & Phase Relationships**

```
Block A (HTTP Client) ──► D1 (Hylo Constants) ──► D2 (Parser) ──► D3 (Trade Service) ──► D4 (API Integration)
Block B (Tokens) ──────────┘                      │                    │
                                                   └──► D3 (ATA Derivation) ──┘

Existing Server ────────────────────────────────────────────────────────────► D4 (Handler Integration)
```

**Phase Dependencies:**

- **D1** independent - uses existing token constants from Block B
- **D2** depends on D1 constants and Block A transaction types
- **D3** depends on D1, D2, and both Blocks A (HTTP client) and B (ATA derivation)
- **D4** depends on D3 and existing server infrastructure

**Testing Strategy:**

- Each phase has focused unit tests with >85% coverage target
- D2 uses mock transaction data for parser validation
- D3 uses mocked Solana responses following Block B patterns
- D4 includes HTTP integration tests for end-to-end validation
- Leverage existing test patterns and mock infrastructure from Blocks A/B

---

### **Block D Completion Checklist**

**Files to Create:** _(Total estimated: 8-9 files for Block D)_

- [x] **D1: 2 files** ✅
  - [x] `internal/hylo/constants.go`, `internal/hylo/config.go` ✅
- [x] **D2: 2 files** ✅
  - [x] `internal/hylo/parser.go`, `internal/hylo/types.go` ✅
- [x] **D3: 2 files** ✅
  - [x] `internal/trades/service.go`, `internal/trades/types.go` ✅
- [x] **D4: 0 new files** (updating existing) ✅
  - [x] Update `internal/server/handlers.go`, `internal/server/server.go`, `internal/server/routes.go` ✅

**Key Milestones:**

- [x] **D1 Complete:** Hylo constants and configuration ready ✅
- [x] **D2 Complete:** Transaction parser working with balance-change detection ✅
- [x] **D3 Complete:** Trade service fetching and parsing xSOL trades in real-time ✅
- [x] **D4 Complete:** Trade history API endpoint functional with pagination ✅
- [x] **Block D Done:** Complete xSOL transaction history functionality, ready for production ✅
