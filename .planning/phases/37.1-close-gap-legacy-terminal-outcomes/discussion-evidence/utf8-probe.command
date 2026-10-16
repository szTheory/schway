GOCACHE=/tmp/ai-lang-verification-gocache go test -overlay=/tmp/schway-371-utf8-overlay.json ./internal/compiler/cgen -run '^TestSchway371ReasonUTF8AdmissionProbe$' -count=1 -v -timeout=120s
