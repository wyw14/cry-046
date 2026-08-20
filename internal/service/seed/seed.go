// Package seed provides demo data seeding for the platform.
// The seeders are idempotent: running them twice does not overwrite
// existing rows. They never delete or update demo data — they only
// create rows that do not exist yet.
package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// Seeder seeds demo data into the given application services.
type Seeder struct {
	projects   *application.ProjectsApp
	parties    *application.PartiesApp
	batches    *application.BatchesApp
	cycles     *application.CyclesApp
	rules      *application.RulesApp
	users      *application.UsersApp
	imports    *application.ImportsApp
	evaluator  *application.EvaluateApp
	summariser *application.SummaryApp
	exceptions *application.ExceptionsApp
	clock      application.Clock
}

// New constructs a Seeder.
func New(
	projects *application.ProjectsApp,
	parties *application.PartiesApp,
	batches *application.BatchesApp,
	cycles *application.CyclesApp,
	rules *application.RulesApp,
	users *application.UsersApp,
	imports *application.ImportsApp,
	evaluator *application.EvaluateApp,
	summariser *application.SummaryApp,
	exceptions *application.ExceptionsApp,
	clock application.Clock,
) *Seeder {
	if clock == nil {
		clock = systemClock{}
	}
	return &Seeder{
		projects: projects, parties: parties, batches: batches, cycles: cycles,
		rules: rules, users: users, imports: imports, evaluator: evaluator,
		summariser: summariser, exceptions: exceptions, clock: clock,
	}
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// Result summarises what the seeder did.
type Result struct {
	Projects     int
	Parties      int
	Batches      int
	Cycles       int
	RuleVersions int
	Users        int
	Entries      application.UpsertSummary
	Evaluated    application.EvaluateResult
	Recalculated application.RecalcResult
}

// Seed runs the demo seed.
func (s *Seeder) Seed(ctx context.Context, tenantID string, projectCount, batchCount int) (Result, error) {
	var r Result
	now := s.clock.Now()

	// Users.
	userIDs := make(map[string]string, 4)
	users := []struct {
		username, display, email string
		role                     domain.Role
	}{
		{"admin", "管理员", "admin@local", domain.RoleAdmin},
		{"operator", "运营人员", "operator@local", domain.RoleOperator},
		{"assignee", "处理人", "assignee@local", domain.RoleAssignee},
		{"reviewer", "复核人", "reviewer@local", domain.RoleReviewer},
	}
	for _, u := range users {
		existing, err := s.users.GetByUsername(ctx, tenantID, u.username)
		if err == nil {
			userIDs[u.username] = existing.ID
			continue
		}
		if !domain.IsNotFound(err) {
			return r, err
		}
		created, err := s.users.Create(ctx, application.CreateUserInput{
			TenantID: tenantID, Username: u.username,
			DisplayName: u.display, Email: u.email, Role: u.role,
			PasswordHash: "demo:" + u.username,
		})
		if err != nil {
			return r, err
		}
		userIDs[u.username] = created.ID
		r.Users++
	}
	adminID := userIDs["admin"]
	assigneeID := userIDs["assignee"]

	// Projects.
	projectIDs := make([]string, 0, projectCount)
	for i := 0; i < projectCount; i++ {
		code := fmt.Sprintf("WS-%d", 20260001+i)
		list, _, err := s.projects.List(ctx, application.ListQuery{
			TenantID: tenantID, PageSize: 1, Filters: map[string]string{"code": code},
		})
		if err == nil && len(list) > 0 {
			projectIDs = append(projectIDs, list[0].ID)
			continue
		}
		p, err := s.projects.Create(ctx, application.CreateProjectInput{
			TenantID:     tenantID,
			Code:         code,
			Name:         fmt.Sprintf("公益项目 %d", i+1),
			Sponsor:      "示例资助方",
			AnnualBudget: int64(10_000_000 + i*1_000_000), // 10万-12万元
			StartYear:    2026,
			EndYear:      2027,
			Metadata:     map[string]string{"demo": "true"},
		})
		if err != nil {
			return r, err
		}
		projectIDs = append(projectIDs, p.ID)
		r.Projects++
	}

	// Parties.
	partySpecs := []struct {
		name    string
		ptype   domain.PartyType
		contact string
	}{
		{"示例资助方", domain.PartyTypeDonor, "sponsor@local"},
		{"示例执行方", domain.PartyTypeImplementer, "impl@local"},
		{"示例受益方", domain.PartyTypeBeneficiary, "+8613800000001"},
		{"示例中间方", domain.PartyTypeIntermediary, "inter@local"},
	}
	partyIDs := make(map[string]string, len(partySpecs))
	for _, ps := range partySpecs {
		list, _, err := s.parties.List(ctx, application.ListQuery{TenantID: tenantID, PageSize: 100, Filters: map[string]string{"name": ps.name}})
		if err == nil && len(list) > 0 {
			partyIDs[ps.name] = list[0].ID
			continue
		}
		created, err := s.parties.Create(ctx, application.CreatePartyInput{
			TenantID: tenantID, Name: ps.name, Type: ps.ptype, Contact: ps.contact,
			Metadata: map[string]string{"demo": "true"},
		})
		if err != nil {
			return r, err
		}
		partyIDs[ps.name] = created.ID
		r.Parties++
	}

	// Batches + cycles + rules + entries + exceptions + summary per project.
	for i, pid := range projectIDs {
		for j := 0; j < batchCount; j++ {
			batchCode := fmt.Sprintf("FB-2026-%02d-%02d", i+1, j+1)
			var batch domain.FundingBatch
			list, _, err := s.batches.List(ctx, application.ListQuery{
				TenantID: tenantID, PageSize: 100,
				Filters: map[string]string{"code": batchCode, "project_id": pid},
			})
			if err == nil && len(list) > 0 {
				batch = list[0]
			} else {
				created, err := s.batches.Create(ctx, application.CreateBatchInput{
					TenantID:            tenantID,
					ProjectID:           pid,
					Code:                batchCode,
					DonorPartyID:        partyIDs["示例资助方"],
					ImplementerPartyID:  partyIDs["示例执行方"],
					IntermediaryPartyID: partyIDs["示例中间方"],
					TotalAmount:         int64(2_000_000 + j*500_000),
					Currency:            "CNY",
					DisbursedAt:         now.AddDate(0, -3-j, 0),
					Metadata:            map[string]string{"demo": "true"},
				})
				if err != nil {
					return r, err
				}
				batch = created
				r.Batches++
			}

			// Cycle.
			cycles, _, err := s.cycles.List(ctx, application.ListQuery{
				TenantID: tenantID, PageSize: 100,
				Filters: map[string]string{"funding_batch_id": batch.ID},
			})
			var cycle domain.SettlementCycle
			if err == nil && len(cycles) > 0 {
				cycle = cycles[0]
			} else {
				created, err := s.cycles.Create(ctx, application.CreateCycleInput{
					TenantID:       tenantID,
					ProjectID:      pid,
					FundingBatchID: batch.ID,
					Year:           2026,
					Quarter:        j%4 + 1,
					StartDate:      now.AddDate(0, -3-j, 0),
					EndDate:        now.AddDate(0, -j-1, 0),
				})
				if err != nil {
					return r, err
				}
				cycle = created
				r.Cycles++
			}

			// Rule version. Codes are unique within (tenant, project) so
			// include the project index to allow multiple projects to each
			// have their own rule version.
			rvCode := fmt.Sprintf("RV-2026-%02d-Q%d", i+1, j%4+1)
			rvList, _, err := s.rules.List(ctx, application.ListQuery{
				TenantID: tenantID, PageSize: 100, Filters: map[string]string{"code": rvCode, "project_id": pid},
			})
			var rv domain.RuleVersion
			if err == nil && len(rvList) > 0 {
				rv = rvList[0]
			} else {
				created, err := s.rules.Create(ctx, application.CreateRuleVersionInput{
					TenantID:    tenantID,
					Code:        rvCode,
					ProjectID:   pid,
					Description: "演示规则版本",
					Rules: []domain.RuleDefinition{
						{ID: domain.NewID(), Code: "AMOUNT_ZERO", Description: "金额为0", Severity: domain.SeverityHigh, Category: "amount", Expression: "amount == 0", DeadlineHours: 48},
						{ID: domain.NewID(), Code: "AMOUNT_NEG", Description: "金额为负", Severity: domain.SeverityCritical, Category: "amount", Expression: "amount < 0", DeadlineHours: 24},
						{ID: domain.NewID(), Code: "CURRENCY_NOT_CNY", Description: "币种非CNY", Severity: domain.SeverityLow, Category: "currency", Expression: "currency == USD", DeadlineHours: 72},
						{ID: domain.NewID(), Code: "OCCURRED_TOO_OLD", Description: "发生日期过旧", Severity: domain.SeverityMedium, Category: "date", Expression: "occurred_before 2025-01-01", DeadlineHours: 96},
					},
				})
				if err != nil {
					return r, err
				}
				rv = created
				r.RuleVersions++
			}
			if rv.Status == domain.RuleVersionStatusDraft {
				rv, err = s.rules.Publish(ctx, tenantID, rv.ID)
				if err != nil {
					return r, err
				}
			}

			// Entries.
			entryInputs := []application.ImportEntryInput{
				{
					TenantID: tenantID, CycleID: cycle.ID, BatchID: batch.ID, ProjectID: pid,
					SourceID: fmt.Sprintf("S-%d-%d-1", i, j), Source: domain.EntrySourceImport,
					PayeePartyID: partyIDs["示例受益方"], PayerPartyID: partyIDs["示例资助方"],
					Amount: 100_00, Currency: "CNY",
					OccurredAt: now.AddDate(0, -3-j, 0),
					Metadata:   map[string]string{"note": "正常明细"},
				},
				{
					TenantID: tenantID, CycleID: cycle.ID, BatchID: batch.ID, ProjectID: pid,
					SourceID: fmt.Sprintf("S-%d-%d-2", i, j), Source: domain.EntrySourceImport,
					PayeePartyID: partyIDs["示例受益方"], PayerPartyID: partyIDs["示例资助方"],
					Amount: 0, Currency: "CNY",
					OccurredAt: now.AddDate(0, -3-j, 0),
					Metadata:   map[string]string{"note": "金额为零"},
				},
				{
					TenantID: tenantID, CycleID: cycle.ID, BatchID: batch.ID, ProjectID: pid,
					SourceID: fmt.Sprintf("S-%d-%d-3", i, j), Source: domain.EntrySourceImport,
					PayeePartyID: partyIDs["示例受益方"], PayerPartyID: partyIDs["示例资助方"],
					Amount: 200_00, Currency: "CNY",
					OccurredAt: now.AddDate(-2, 0, 0),
					Metadata:   map[string]string{"note": "发生日期过早"},
				},
			}
			summary, _, err := s.imports.ImportEntries(ctx, adminID, entryInputs)
			if err != nil {
				return r, err
			}
			r.Entries.Created += summary.Created
			r.Entries.Updated += summary.Updated
			r.Entries.Skipped += summary.Skipped

			// Evaluate.
			ev, err := s.evaluator.EvaluateCycle(ctx, application.EvaluateCycleInput{
				TenantID: tenantID, CycleID: cycle.ID, RuleVersionID: rv.ID, ActorID: adminID,
			})
			if err != nil {
				return r, err
			}
			r.Evaluated.ScannedEntries += ev.ScannedEntries
			r.Evaluated.CreatedExceptions += ev.CreatedExceptions

			// Assign one pending exception to the demo assignee for the workspace demo.
			if assigneeID != "" {
				excList, _, err := s.exceptions.List(ctx, application.ExceptionListQuery{
					ListQuery: application.ListQuery{TenantID: tenantID, PageSize: 100},
					CycleID:   cycle.ID,
					Status:    string(domain.ExceptionStatusPending),
				})
				_ = err
				for _, ex := range excList {
					if _, err := s.exceptions.Assign(ctx, application.AssignInput{
						TenantID:    tenantID,
						ExceptionID: ex.ID,
						AssigneeID:  assigneeID,
						AuthorID:    adminID,
						Note:        "演示分派",
					}); err == nil {
						break
					}
				}
			}
		}
	}

	// Recalculate summaries for the latest cycle of each project.
	for _, pid := range projectIDs {
		cycles, _, err := s.cycles.List(ctx, application.ListQuery{
			TenantID: tenantID, PageSize: 100, Filters: map[string]string{"project_id": pid},
		})
		if err != nil || len(cycles) == 0 {
			continue
		}
		cycle := cycles[0]
		rvList, _, err := s.rules.List(ctx, application.ListQuery{
			TenantID: tenantID, PageSize: 100, Filters: map[string]string{"project_id": pid, "status": string(domain.RuleVersionStatusPublished)},
		})
		if err != nil || len(rvList) == 0 {
			continue
		}
		rv := rvList[0]
		recalc, err := s.summariser.Recalculate(ctx, application.RecalcInput{
			TenantID: tenantID, CycleID: cycle.ID, RuleVersionID: rv.ID,
			ActorID: adminID, TriggerReason: "demo seed recalculation",
		})
		if err != nil {
			return r, err
		}
		r.Recalculated = recalc
	}

	return r, nil
}
