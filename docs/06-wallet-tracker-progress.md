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

### **Phase B4: API Integration & Token Service** _(Depends on B3 & existing server - 2-3 hours)_

**Deliverables:**

- [ ] REST endpoint `GET /wallet/:address/balances`
- [ ] Refactored TokenService (renamed from BalanceService) for Block C preparation
- [ ] Direct WalletBalances JSON response (no API wrapper)
- [ ] Strict error handling (all tokens must succeed)
- [ ] Integration with existing server routes

**Components:**

- [ ] Refactor `internal/tokens/service.go` - Rename BalanceService to TokenService
- [ ] Update `internal/server/server.go` - Add TokenService to Server struct
- [ ] Update `internal/server/routes.go` - Add wallet handler and route
- [ ] Update `internal/tokens/service_test.go` - Update tests for TokenService (end of phase)

**Implementation Tasks:**

1. **Service Refactoring to TokenService** (30 min) 🟡

   - [ ] Rename `BalanceService` to `TokenService` in existing `service.go`
   - [ ] Rename `NewBalanceService()` to `NewTokenService()` constructor
   - [ ] Keep existing `GetBalances()` method as `GetWalletBalances()` with strict error handling
   - [ ] Maintain all existing functionality while preparing for Block C integration

2. **Server Integration & Service Bootstrap** (45 min) 🟡

   - [ ] Add `tokenService *tokens.TokenService` to Server struct
   - [ ] Initialize TokenService in `NewServer()` using `solanaService.GetHTTPClient()`
   - [ ] Follow existing dependency injection pattern (like `solanaService`)
   - [ ] Validate TokenService creation and configuration loading

3. **HTTP Handler & Route Implementation** (1 hour) 🟡

   - [ ] Add `handleWalletBalances(w http.ResponseWriter, r *http.Request)` to Server
   - [ ] Extract wallet address using `chi.URLParam()` (not fiber)
   - [ ] Call `tokenService.GetWalletBalances()` with strict error handling
   - [ ] Return direct `WalletBalances` JSON response (Option A format)
   - [ ] Add route `/wallet/{address}/balances` using chi router patterns

4. **Testing & Validation** (30 min) 🟡
   - [ ] Unit tests for TokenService wrapper functionality
   - [ ] HTTP integration tests for balance endpoint
   - [ ] Test invalid wallet address handling (400 errors)
   - [ ] Test RPC failures and strict error responses (500 errors)
   - [ ] Validate response format matches `WalletBalances` struct

**Acceptance Criteria:**

- [ ] `GET /wallet/{address}/balances` endpoint functional with chi router
- [ ] Direct `WalletBalances` JSON response (no wrapper)
- [ ] All three tokens (hyUSD, sHYUSD, xSOL) must succeed or return 500
- [ ] Invalid addresses return 400 with clear error message
- [ ] TokenService provides foundation for Block C price integration
- [ ] Tests achieve >80% coverage

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
- [ ] **B4: 0 new files** (refactoring existing) 🟡
  - [ ] Refactor existing `service.go` and `service_test.go` to use TokenService
- [ ] **Integration: 2 files** 🟡
  - [ ] Update `internal/server/server.go` (add TokenService)
  - [ ] Update `internal/server/routes.go` (add handler and route)

**Key Milestones:**

- [x] **B1 Complete:** Token configuration and types ready ✅
- [x] **B2 Complete:** ATA derivation working with test vectors ✅
- [x] **B3 Complete:** Balance service fetching all token balances ✅
- [ ] **B4 Complete:** TokenService and wallet balance API endpoint functional
- [ ] **Block B Done:** Wallet balance functionality complete, ready for Block C (Price Engine)
