GOCACHE=/tmp/ai-lang-verification-gocache go test -overlay=/tmp/schway-371-reason-overlay.json ./internal/compiler/cgen -run '^TestSchway371ReasonByteProbe$' -count=1 -v -timeout=120s
