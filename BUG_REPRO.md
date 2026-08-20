# Bug reproduction

在同一租户和结算周期中，先用已发布的旧规则版本评估一笔命中明细，再用新发布但规则编码相同的版本重新评估。当前实现把两次评估视为同一幂等请求，第二次不生成新异常，导致规则版本和快照丢失。

验证：`go test ./internal/application -run '^TestEvaluateCycle_RuleVersionIsPartOfIdempotencyKey$' -count=1`
