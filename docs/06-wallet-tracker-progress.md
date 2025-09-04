# Hylo Wallet Tracker — Backend Progress **by Blocks**

## Block A — Helios Connectivity

### **Overview**

Establishes core Solana RPC connectivity layer using Helios, providing both HTTP and WebSocket clients with robust error handling, reconnection logic, and commitment-aware operations.

**Module Path:** `/internal/solana`  
**Dependencies:** Helios RPC endpoints (`RPC_HTTP_URL`, `RPC_WS_URL`)

### **Overall Progress**

- [x] **Phase A1:** Foundation & HTTP Client _(4/4 tasks completed)_ ✅
- [ ] **Phase A2:** WebSocket Client & Subscriptions _(0/4 tasks completed)_
- [ ] **Phase A3:** Connection Health & Resilience _(0/4 tasks completed)_
- [ ] **Phase A4:** Integration & Performance Validation _(0/3 tasks completed)_

**Block A Status:** 🟡 In Progress _(4/15 total tasks completed)_

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

1. **Config & Types** (30 min)

   - [ ] Define `Config` struct with HTTP URL, timeouts, retry settings
   - [ ] Define core Solana types (`Address`, `Slot`, `Commitment`, `AccountInfo`)
   - [ ] Add validation for configuration parameters

2. **HTTP Client Core** (2 hours)

   - [ ] Implement base HTTP client with proper timeout handling
   - [ ] Add exponential backoff retry logic (1s → 2s → 4s → 8s, max 3 retries)
   - [ ] Implement JSON-RPC request/response handling
   - [ ] Add proper context cancellation support

3. **RPC Methods** (1 hour)

   - [ ] `GetAccount(ctx, address, commitment)` → `AccountInfo`
   - [ ] `GetTransaction(ctx, signature)` → `TransactionDetails`
   - [ ] `GetSignaturesForAddress(ctx, address, before, limit)` → `[]SignatureInfo`

4. **Testing** (1 hour)
   - [ ] Unit tests with mock HTTP server
   - [ ] Test retry logic with simulated failures
   - [ ] Test timeout handling and context cancellation
   - [ ] Test JSON-RPC error response handling

**Acceptance Criteria:**

- [ ] All HTTP methods work with real Helios endpoint
- [ ] Retry logic handles transient network failures
- [ ] Tests achieve >90% coverage
- [ ] Configuration validates required parameters
- [ ] Proper error wrapping and context handling

### **Phase A2: WebSocket Client & Subscriptions** _(Independent - 3-4 hours)_

**Deliverables:**

- [ ] WebSocket client with subscription management
- [ ] Account and logs subscription methods
- [ ] Message parsing and routing
- [ ] Comprehensive test suite with mock WebSocket server

**Components:**

- [ ] `websocket.go` - WebSocket client implementation
- [ ] `websocket_test.go` - WebSocket tests with mock server
- [ ] `subscription.go` - Subscription lifecycle management
- [ ] `subscription_test.go` - Subscription tests
- [ ] `mock_server_test.go` - Test helper for WebSocket mocking

**Implementation Tasks:**

1. **WebSocket Foundation** (1.5 hours)

   - [ ] Basic WebSocket connection with `nhooyr.io/websocket`
   - [ ] JSON-RPC message sending/receiving over WebSocket
   - [ ] Proper context handling and graceful shutdown

2. **Subscription Management** (1.5 hours)

   - [ ] `AccountSubscribe(ctx, address, commitment)` → subscription ID
   - [ ] `LogsSubscribe(ctx, mentions, commitment)` → subscription ID
   - [ ] `Unsubscribe(ctx, subscriptionId)` → confirmation
   - [ ] Track active subscriptions in memory map

3. **Message Handling** (1 hour)

   - [ ] Parse incoming subscription messages
   - [ ] Route messages to appropriate channels based on subscription ID
   - [ ] Handle subscription confirmations vs. data messages
   - [ ] Buffer messages to prevent blocking

4. **Testing** (1 hour)
   - [ ] Mock WebSocket server for testing
   - [ ] Test subscription lifecycle (subscribe → receive → unsubscribe)
   - [ ] Test message parsing and routing
   - [ ] Test concurrent subscriptions

**Acceptance Criteria:**

- [ ] Successfully subscribes to account changes
- [ ] Successfully subscribes to program logs
- [ ] Messages route correctly to subscribers
- [ ] Subscription cleanup works properly
- [ ] Tests achieve >90% coverage

### **Phase A3: Connection Health & Resilience** _(Depends on A1 + A2 - 2-3 hours)_

**Deliverables:**

- [ ] Connection health monitoring
- [ ] Auto-reconnection with exponential backoff
- [ ] Subscription recovery after reconnection
- [ ] Health check endpoint readiness

**Components:**

- [ ] `manager.go` - Connection manager and health monitoring
- [ ] `manager_test.go` - Manager tests with connection simulation
- [ ] `health.go` - Health check logic
- [ ] `reconnect.go` - Reconnection logic with backoff

**Implementation Tasks:**

1. **Health Monitoring** (1 hour)

   - [ ] Track HTTP client health (last successful request)
   - [ ] Track WebSocket health (connection state, last message)
   - [ ] Implement health check method for external probing

2. **Reconnection Logic** (1 hour)

   - [ ] Exponential backoff with jitter (1s → 2s → 4s → max 30s)
   - [ ] Automatic WebSocket reconnection on disconnect
   - [ ] Re-establish all active subscriptions after reconnect

3. **Connection Manager** (30 min)

   - [ ] Unified interface combining HTTP and WebSocket clients
   - [ ] Coordinate health status across both clients
   - [ ] Graceful shutdown handling

4. **Testing** (30 min)
   - [ ] Test reconnection scenarios with simulated disconnects
   - [ ] Test subscription recovery after reconnection
   - [ ] Test health monitoring accuracy

**Acceptance Criteria:**

- [ ] Auto-reconnects within 30s of disconnect
- [ ] All subscriptions restore after reconnection
- [ ] Health checks accurately reflect connection status
- [ ] No subscription message loss during normal operation
- [ ] Graceful shutdown completes within 5s

### **Phase A4: Integration & Performance Validation** _(Depends on A1-A3 - 1-2 hours)_

**Deliverables:**

- [ ] End-to-end integration tests against real Helios
- [ ] Performance benchmarks and load testing
- [ ] Usage documentation and examples
- [ ] Production readiness validation

**Components:**

- [ ] `integration_test.go` - End-to-end tests against real Helios
- [ ] `benchmark_test.go` - Performance benchmarks
- [ ] `example_test.go` - Usage examples
- [ ] `README.md` - Component documentation

**Implementation Tasks:**

1. **Integration Testing** (45 min)

   - [ ] Test against real Helios mainnet endpoint
   - [ ] Verify all methods work with production data
   - [ ] Test with reference wallet: `A3wpCHTBFHQr7JeGFSA6cbTHJ4rkXgHZ2BLj2rZDyc6g`

2. **Performance Testing** (30 min)

   - [ ] Benchmark HTTP request latency (target <2s P95)
   - [ ] Benchmark WebSocket subscription throughput
   - [ ] Test concurrent subscription handling (10+ subscriptions)

3. **Documentation** (15 min)
   - [ ] Usage examples for each major function
   - [ ] Configuration documentation with environment variables
   - [ ] Error handling guidelines and common patterns

**Acceptance Criteria:**

- [ ] Integration tests pass against real Helios mainnet
- [ ] HTTP requests complete <2s P95
- [ ] WebSocket supports 10+ concurrent subscriptions
- [ ] Documentation covers all public methods
- [ ] Example code demonstrates typical usage patterns

### **Dependencies & Phase Relationships**

```
A1 (HTTP Client) ──┐
                   ├──► A3 (Health & Resilience) ──► A4 (Integration)
A2 (WebSocket) ────┘
```

**Phase Independence:**

- **A1 & A2** can be developed in parallel by different developers
- **A3** requires both A1 & A2 interfaces but can mock them for initial testing
- **A4** validates the complete system works together

**Testing Strategy:**

- Each phase has its own comprehensive test suite
- Earlier phases use mocks/stubs for external dependencies
- Later phases include integration testing against real services
- All phases maintain >90% test coverage requirement

---

### **Block A Completion Checklist**

**Files Created:** _(Total: 15 files)_

- [x] A1: 6 files (config, types, http_client + tests, errors, testdata) ✅
- [ ] A2: 5 files (websocket + tests, subscription + tests, mock_server_test)
- [ ] A3: 4 files (manager + tests, health, reconnect)
- [ ] A4: 4 files (integration_test, benchmark_test, example_test, README)

**Key Milestones:**

- [x] **A1 Complete:** HTTP client ready with >90% test coverage ✅
- [ ] **A2 Complete:** WebSocket subscriptions working with >90% test coverage
- [ ] **A3 Complete:** Auto-reconnection and health monitoring operational
- [ ] **A4 Complete:** Integration tests pass, performance benchmarks meet targets
- [ ] **Block A Done:** All 15 tasks completed, ready for Block B integration
