# Service 契约基线

`service_input.json` 和 `service_results.sha256.json` 在本次结构重构开始前，由当时的 `makeTestInput` 和正式 `Service.Simulate` 生成。固定 `StartTime = 1700000000000`，记录 seed `11`、`42`、`20261001` 下整个 `MatchResult` 的 JSON SHA-256。

`service_contract_test.go` 属于外部 `matchengine_test` 包，只使用公开 API。它同时检查完整战报兼容、输入快照只读、结构化错误和 Context 取消。

结构整理不应更新预期摘要。后续有意修改模拟算法时，需先审查战报差异和对应规格，再明确更新基线；不要通过自动重新生成摘要掩盖回归。
