---
id: 260924-bxh
type: quick
mode: quick-full
status: planned
---

<objective>
Make the QLT-01 registry account for spike 007 in a way consistent with the registry audit and the spike's actual evidence, then restore a green test suite.

Purpose: The new spike directory is discovered by the registry coverage test, which currently fails because the registry has no row for it.
Output: An honest registry disposition for spike 007 and passing focused and full Go test verification.
</objective>

<must_haves>
  <truths>
    <truth>The QLT-01 registry includes a valid disposition row for spike 007.</truth>
    <truth>The disposition accurately reflects that spike 007 has a focused production regression test but no QLT control ID in the shipped control inventory.</truth>
    <truth>The focused registry coverage test and full Go test suite pass.</truth>
  </truths>
  <artifacts>
    <artifact>internal/compiler/session/qlt01_registry.json</artifact>
  </artifacts>
  <key_links>
    <key_link>The row's waiver cites the spike evidence and existing production regression test without fabricating a live control identifier.</key_link>
  </key_links>
</must_haves>

<tasks>
  <task type="auto">
    <name>Task: Register spike 007 with an evidence-grounded disposition</name>
    <files>internal/compiler/session/qlt01_registry.json</files>
    <action>Read the neighboring spike 006 waiver and apply the same schema to spike 007. Record its hazard as the risk that CFG liveness fails to distinguish the unused branch edge from the consuming point for a loan created before a branch. Since `TestEdgeSpecificLiveOut` supplies the regression evidence but is not a `control:` ID returned by the shipped QLT control inventory, use a waiver with a concrete reason and citation to the spike README, that test, and the shipped liveness implementation; do not invent a control ID. Preserve valid JSON and existing registry rows.</action>
    <verify>
      <automated>go test ./internal/compiler/session -run '^TestQLT01RegistryCoversAllFiveSpikes$' -count=1</automated>
      <automated>go test ./...</automated>
    </verify>
    <done>The registry coverage and completeness checks accept a single, justified spike 007 disposition, and all Go packages pass.</done>
  </task>
</tasks>
