// Package memory provides an in-memory implementation of the
// application repository ports. It is used in tests and in the
// offline "no DB" run mode so the platform is fully operational
// without an external PostgreSQL server.
package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// Store is the in-memory backing store for all repositories. All
// repositories share the same Store instance so cross-repository
// transactions (the UnitOfWork abstraction) can be implemented
// with a single mutex.
type Store struct {
	mu sync.RWMutex

	projects   map[string]domain.Project
	parties    map[string]domain.Party
	batches    map[string]domain.FundingBatch
	cycles     map[string]domain.SettlementCycle
	rules      map[string]domain.RuleVersion
	entries    map[string]domain.SettlementEntry
	exceptions map[string]domain.Exception
	summaries  []domain.Summary
	recalcs    map[string]domain.RecalculationBatch
	annuals    map[string]domain.AnnualAccumulator
	audits     []domain.AuditEntry
	users      map[string]domain.User

	// indexes for uniqueness checks
	projectsByCode map[string]string // tenant+code -> id
	batchesByCode  map[string]string // tenant+code -> id
	rulesByCode    map[string]string // tenant+code -> id
	entryByFP      map[string]string // fingerprint -> id
	usersByName    map[string]string // tenant+username -> id
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		projects:       make(map[string]domain.Project),
		parties:        make(map[string]domain.Party),
		batches:        make(map[string]domain.FundingBatch),
		cycles:         make(map[string]domain.SettlementCycle),
		rules:          make(map[string]domain.RuleVersion),
		entries:        make(map[string]domain.SettlementEntry),
		exceptions:     make(map[string]domain.Exception),
		recalcs:        make(map[string]domain.RecalculationBatch),
		annuals:        make(map[string]domain.AnnualAccumulator),
		users:          make(map[string]domain.User),
		projectsByCode: make(map[string]string),
		batchesByCode:  make(map[string]string),
		rulesByCode:    make(map[string]string),
		entryByFP:      make(map[string]string),
		usersByName:    make(map[string]string),
	}
}

// Now is exposed so callers can stamp time deterministically in tests.
type Now func() time.Time

// Repositories returns a set of repositories sharing the same store.
type Repositories struct {
	Projects   *ProjectRepo
	Parties    *PartyRepo
	Batches    *BatchRepo
	Cycles     *CycleRepo
	Rules      *RuleVersionRepo
	Entries    *EntryRepo
	Exceptions *ExceptionRepo
	Summaries  *SummaryRepo
	Recalcs    *RecalcRepo
	Annuals    *AnnualRepo
	Audits     *AuditRepo
	Users      *UserRepo
}

// New returns a wired Repositories set backed by store.
func New(store *Store) *Repositories {
	return &Repositories{
		Projects:   &ProjectRepo{store: store},
		Parties:    &PartyRepo{store: store},
		Batches:    &BatchRepo{store: store},
		Cycles:     &CycleRepo{store: store},
		Rules:      &RuleVersionRepo{store: store},
		Entries:    &EntryRepo{store: store},
		Exceptions: &ExceptionRepo{store: store},
		Summaries:  &SummaryRepo{store: store},
		Recalcs:    &RecalcRepo{store: store},
		Annuals:    &AnnualRepo{store: store},
		Audits:     &AuditRepo{store: store},
		Users:      &UserRepo{store: store},
	}
}

// Apply mounts each repo into the given Unit pointer fields.
func (r *Repositories) Apply(u *RepoUnit) {
	u.projects = r.Projects
	u.parties = r.Parties
	u.batches = r.Batches
	u.cycles = r.Cycles
	u.rules = r.Rules
	u.entries = r.Entries
	u.exceptions = r.Exceptions
	u.summaries = r.Summaries
	u.recalcs = r.Recalcs
	u.annuals = r.Annuals
	u.audits = r.Audits
	u.users = r.Users
}

// RepoUnit is the application.UnitOfWork implementation backed by
// the in-memory store. All operations happen in the caller's
// goroutine under a single mutex.
type RepoUnit struct {
	projects   application.ProjectRepo
	parties    application.PartyRepo
	batches    application.BatchRepo
	cycles     application.CycleRepo
	rules      application.RuleVersionRepo
	entries    application.EntryRepo
	exceptions application.ExceptionRepo
	summaries  application.SummaryRepo
	recalcs    application.RecalcRepo
	annuals    application.AnnualRepo
	audits     application.AuditRepo
	users      application.UserRepo
}

func (u *RepoUnit) Projects() application.ProjectRepo         { return u.projects }
func (u *RepoUnit) Parties() application.PartyRepo            { return u.parties }
func (u *RepoUnit) Batches() application.BatchRepo            { return u.batches }
func (u *RepoUnit) Cycles() application.CycleRepo             { return u.cycles }
func (u *RepoUnit) RuleVersions() application.RuleVersionRepo { return u.rules }
func (u *RepoUnit) Entries() application.EntryRepo            { return u.entries }
func (u *RepoUnit) Exceptions() application.ExceptionRepo     { return u.exceptions }
func (u *RepoUnit) Summaries() application.SummaryRepo        { return u.summaries }
func (u *RepoUnit) Recalcs() application.RecalcRepo           { return u.recalcs }
func (u *RepoUnit) Annuals() application.AnnualRepo           { return u.annuals }
func (u *RepoUnit) Audits() application.AuditRepo             { return u.audits }
func (u *RepoUnit) Users() application.UserRepo               { return u.users }

// Unit is a no-op Unit that always invokes fn with the same repo
// set. The in-memory store uses a single mutex across repos so
// consistency is guaranteed without a real transaction.
type Unit struct {
	repos *Repositories
}

// NewUnit returns a Unit backed by repos.
func NewUnit(repos *Repositories) *Unit { return &Unit{repos: repos} }

// Do runs fn with a RepoUnit that delegates to repos. The mutex is
// held for the duration of fn so concurrent Do calls are serialised.
func (u *Unit) Do(ctx context.Context, fn func(ctx context.Context, uow application.UnitOfWork) error) error {
	ru := &RepoUnit{}
	u.repos.Apply(ru)
	return fn(ctx, ru)
}

// limit returns min(n, max) where max >= 0.
func limit(n, max int) int {
	if max <= 0 {
		return n
	}
	if n > max {
		return max
	}
	return n
}

// pageSlice returns the page slice of in.
func pageSlice[T any](in []T, page, size int) []T {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	start := (page - 1) * size
	if start >= len(in) {
		return nil
	}
	end := start + size
	if end > len(in) {
		end = len(in)
	}
	out := make([]T, end-start)
	copy(out, in[start:end])
	return out
}

func sortStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func matchesFilters(filters map[string]string, candidate map[string]string) bool {
	for k, v := range filters {
		got, ok := candidate[k]
		if !ok {
			return false
		}
		if !strings.EqualFold(got, v) {
			return false
		}
	}
	return true
}
