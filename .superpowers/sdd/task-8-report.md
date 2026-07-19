### Task 8 Report: KeyService

**Status:** DONE

**Commit:** `e7e1195` feat(service): add KeyService with key generation, creation, and status query

**Build:** `go build ./internal/service/` -- clean, no errors, no warnings

**File created:** `internal/service/key_service.go`

**What was implemented:**
- `KeyService` struct with configurable prefix, key length, and suffix length
- `NewKeyService(db)` constructor (defaults: `sk-` prefix, 16-byte key, 4-char suffix)
- `WithConfig()` for overriding defaults
- `generateRawKey()` -- crypto/rand random hex key generation
- `hashKey()` -- SHA256 hashing for secure DB storage
- `CreateKey(req)` -- full create flow: generate key, compute hash, persist to DB
- `FindByRawKey(rawKey)` -- hash-based lookup returning `nil, nil` for not-found
- `GetKeyStatus(rawKey)` -- human-readable status result with formatted timestamps

**Consumed types:** `model.Key`, `model.KeyBillingMode`, `model.KeyStatus`, `gorm.DB`

**Concerns:** None
