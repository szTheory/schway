# Deferred items

## Pre-existing independent core-validator mismatch

**Owner:** Core validation / Phase 07 legacy fixture follow-up.

The Phase 34 and Phase 36 frozen corpora both include
testdata/phase07/duplicate_function_name.schway. CheckFile admits its core,
while an independent corevalidate.Validate call reports
core.duplicate_function_id for
s1:phase07.duplicate_function_name:fn:helper.
This plan captures the result and exact core bytes in the pre-emitter census.
The mismatch predates this plan and does not alter the UTF-8 admission or route
census work, so no checker or fixture behavior was changed here.
