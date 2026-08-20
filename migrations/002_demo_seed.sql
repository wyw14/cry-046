-- Migration 002_demo_seed.sql
-- Demo seed data. Idempotent: re-running is a no-op. Does NOT overwrite
-- existing rows; the application-level seeder has the same contract.

INSERT INTO users (id, tenant_id, username, display_name, email, role, password_hash)
VALUES
  ('user-admin',    'default', 'admin',    '管理员',  'admin@local',    'admin',    'demo:admin'),
  ('user-operator', 'default', 'operator', '运营人员', 'operator@local', 'operator', 'demo:operator'),
  ('user-assignee', 'default', 'assignee', '处理人',  'assignee@local', 'assignee', 'demo:assignee'),
  ('user-reviewer', 'default', 'reviewer', '复核人',  'reviewer@local', 'reviewer', 'demo:reviewer')
ON CONFLICT (tenant_id, username) DO NOTHING;

INSERT INTO projects (id, tenant_id, code, name, sponsor, annual_budget_cents, start_year, end_year, metadata)
VALUES
  ('proj-ws-2026-01', 'default', 'WS-20260001', '公益项目 1', '示例资助方', 10000000, 2026, 2027, '{"demo":"true"}'),
  ('proj-ws-2026-02', 'default', 'WS-20260002', '公益项目 2', '示例资助方', 11000000, 2026, 2027, '{"demo":"true"}'),
  ('proj-ws-2026-03', 'default', 'WS-20260003', '公益项目 3', '示例资助方', 12000000, 2026, 2027, '{"demo":"true"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO parties (id, tenant_id, name, type, contact, metadata)
VALUES
  ('party-donor',        'default', '示例资助方',   'donor',        'sponsor@local',  '{"demo":"true"}'),
  ('party-implementer',  'default', '示例执行方',   'implementer',  'impl@local',     '{"demo":"true"}'),
  ('party-beneficiary',  'default', '示例受益方',   'beneficiary',  '+8613800000001', '{"demo":"true"}'),
  ('party-intermediary', 'default', '示例中间方',   'intermediary', 'inter@local',    '{"demo":"true"}')
ON CONFLICT DO NOTHING;
