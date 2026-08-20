# 公益项目结算异常处置平台

该项目围绕公益资金结算后的核验与整改闭环：导入带来源指纹的结算明细，按已发布规则生成异常，完成分派、认领、补件、复核、解决和关闭，并按异常状态重算合格金额。系统不包含支付、商城或订单交易。

## 模块职责

- `internal/domain`：规则命中、异常状态机、去重指纹、汇总与年度额度不变量。
- `internal/application`：项目/批次/周期、导入、异常协作、复算、工作台与审计用例。
- `internal/repository`：默认可测试内存仓储与 pgx/PostgreSQL 仓储。
- `internal/transport/http`：`/api/v1` Gin API、分页筛选、稳定错误码与 `request_id`。
- `internal/platform`：本地通知、文件、回调、调度、日志与时钟适配器。
- `migrations` / `scripts`：幂等迁移与不会覆盖已有数据的演示种子。
- `web`：Vue 3 + TypeScript + Vite + Pinia 中文前端，包含结算批次、异常清单、处置详情、复算对比、工作台、规则/审计页面。

## 本地启动

1. 复制 `.env.example` 为 `.env`，启动 PostgreSQL 16。
2. 执行 `make migrate-up` 和 `make seed`。
3. 执行 `make run`，后端监听 `http://localhost:8080`。
4. 在 `web` 下执行 `npm install && npm run dev`，访问 `http://localhost:5173`。

也可执行 `docker compose up --build`。所有通知、附件、回调与定时事件使用本地适配器，不访问外部服务。

演示令牌为 `admin`、`operator`、`assignee`、`reviewer`；它们仅是本地演示身份，不是生产密钥。请求需带 `Authorization: Bearer admin` 与 `X-Tenant-ID: default`。

```bash
curl -H "Authorization: Bearer admin" -H "X-Tenant-ID: default" http://localhost:8080/api/v1/projects?page=1&page_size=20
```

## 状态与数据规则

异常状态链为 `pending -> processing -> review -> resolved -> closed`；无法解决时进入 `escalated`，复核人可返工。解决人与复核人分离，非法跳转返回稳定领域错误。未解决异常对应金额只能进入待处理汇总，不能进入合格金额。每次复算保存触发原因、规则版本、明细和异常输入快照；年度额度调整记录独立留痕。导入使用幂等键、来源指纹和唯一约束，写操作带版本号进行乐观并发控制。

## 验证

本机已实际通过：

```text
go build ./...
go test ./...
go test -race ./...
go vet ./...
npm test                 # 5 files / 26 tests
npm run build            # Vite production build
```

可选 PostgreSQL 集成测试使用 `go test -tags=integration ./...`。运行期的 `.env`、附件、缓存、`node_modules` 和构建产物均由 `.gitignore` 排除。
