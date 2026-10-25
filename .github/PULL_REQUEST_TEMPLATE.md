## What changed

Describe the problem and the implementation.

## Verification

- [ ] `go test ./internal/...`
- [ ] `go vet ./internal/...`
- [ ] `cd frontend && npm run build`
- [ ] I checked whether this changes local data access or persistence.
- [ ] Any estimate is clearly distinguished from official/measured data.
- [ ] No prompt, source-code, credential, raw command, or tool-result content is newly persisted without explicit design review.

## Resource-intelligence boundary

Does this change only observe/advise, or can it change agent/provider behavior? Explain any new permission or mutation.
