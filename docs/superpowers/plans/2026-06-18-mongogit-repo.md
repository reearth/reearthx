# mongogit standalone repository — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `github.com/reearth/mongogit` — a standalone, MIT-licensed Go library containing the git-style version object model and its MongoDB backend, importing nothing from reearthx.

**Architecture:** Port three reearthx packages into one new module: `asset/domain/version` → `mongogit/version`; the MongoDB-backed `mongogit` collection → the root `mongogit` package; and the slices of `mongox`/`usecasex`/`rerror`/`util` that `mongogit` depends on → library-owned code (public boundary types + an unexported `mongoCollection` that mirrors `mongox.Collection`'s method set, so the collection logic ports as renames rather than rewrites).

**Tech Stack:** Go, `go.mongodb.org/mongo-driver`, `github.com/samber/lo`, `github.com/google/uuid`, `github.com/chrispappas/golang-generics-set`. Tests use a live MongoDB via a ported `mongotest` helper. CI is GitHub Actions.

## Global Constraints

- Module path: `github.com/reearth/mongogit`.
- License: **MIT**, `Copyright (c) Re:Earth contributors`.
- The module imports **nothing** from `github.com/reearth/reearthx`. After Task 17, `grep -r reearthx` over the repo returns nothing.
- Dependency versions (match reearthx exactly):
  - `go.mongodb.org/mongo-driver v1.17.3`
  - `github.com/samber/lo v1.50.0`
  - `github.com/google/uuid v1.6.0`
  - `github.com/chrispappas/golang-generics-set v1.0.1`
- Go directive in `go.mod`: `go 1.26.0` (toolchain `1.26.2` is fine).
- CI: no hardcoded tunables — Go version, golangci-lint version, MongoDB image come from Actions `vars` with literal fallbacks; tokens from `secrets`.
- `.github/` contains **no** reference to AI assistants, agents, code-generation tools, "superpowers", or any tooling provenance — in workflows, templates, configs, or comments.
- Every comment written or kept in this repo must survive a pass of the `oh-my-claudecode:ai-slop-cleaner` skill (Task 17): no narration, no "this function does X" restatement, no AI fingerprints.
- No `Co-Authored-By` trailer on any commit. Commit messages are plain Conventional Commits.

## Working locations

- **New repo (build here):** `/Users/dexter/active/mongogit` (created in Task 1).
- **Source to port from:** `/Users/dexter/active/reearthx` (referred to below as `<reearthx>`).

## Porting convention (read before starting)

Most files are **ports**, not new code. For a port, a step gives: the source path, the destination path, and the exact transformations to apply. "Copy verbatim" means byte-for-byte except the stated changes. The canonical content lives in `<reearthx>` at the given path — open it, copy it, apply the listed edits. New (non-ported) files show their full content inline.

**Recurring symbol substitution table** (apply wherever the symbol appears, in both source and test ports):

| reearthx symbol | replacement |
|---|---|
| import `…/asset/domain/version` | import `github.com/reearth/mongogit/version` (package name stays `version`) |
| `mongox.Consumer`, `mongox.SimpleConsumer`, `mongox.SliceConsumer`, `mongox.SliceFuncConsumer`, `mongox.BatchConsumer` | same name, package `mongogit` (drop the `mongox.` qualifier within the package) |
| `mongox.And(...)`, `mongox.AppendE(...)` | `and(...)`, `appendE(...)` (unexported, package `mongogit`) |
| `mongox.Index` | `Index` |
| `usecasex.Sort`, `Pagination`, `CursorPagination`, `OffsetPagination`, `PageInfo`, `Cursor` | same name, package `mongogit` |
| `usecasex.NewPageInfo(...)` | `NewPageInfo(...)` |
| `rerror.ErrNotFound` / `ErrInvalidParams` / `ErrAlreadyExists` | `ErrNotFound` / `ErrInvalidParams` / `ErrAlreadyExists` |
| `rerror.ErrInternalBy(err)`, `rerror.ErrInternalByWithContext(ctx, err)` | `ErrInternalBy(err)` |
| `util.Now()` | `clock.Now()` (import `github.com/reearth/mongogit/internal/clock`) |
| `util.MockNow(t)` | `clock.Mock(t)` |
| `util.Map(s, func(v T) V {…})` | `lo.Map(s, func(v T, _ int) V {…})` |
| `i18n.T("archived")` wrapped by `rerror.NewE` | `errors.New("archived")` |
| `log.Errorfc(...)`, `IsTransactionError`, `usecasex.ErrTransaction` | drop (see Task 9 `wrapError`) |

---

## File Structure

```
mongogit/
├── go.mod, go.sum
├── LICENSE                     MIT
├── README.md
├── .gitignore
├── .golangci.yml
├── .github/workflows/ci.yml
├── errors.go                   Err* sentinels, ErrInternalBy, wrapError          [new]
├── consumer.go                 Consumer + Slice/Func/Batch consumers             [port mongox/consumer.go]
├── bson.go                     and, appendE, getE, appendI                       [port mongox/util.go subset]
├── pagination.go               Cursor, *Pagination, Sort, PageInfo, NewPageInfo  [port usecasex pagination+cursor+pageinfo]
├── index.go                    Index, IndexList, Model(s), Normalize, Equal       [port mongox/index_model.go subset]
├── mongo.go                    unexported mongoCollection (find/aggregate/count/  [port mongox/collection.go + pagination.go subset]
│                               removeAll/paginate/paginateAggregation + helpers)
├── document.go                 Document[T], Meta, MetadataDocument                [port mongogit/document.go]
├── query.go                    apply, applyToPipeline, excludeMetadata            [port mongogit/query.go]
├── collection.go               Collection (git-aware public API)                  [port mongogit/collection.go]
├── consumer_test.go            (port mongox/consumer_test.go — optional, see T5)
├── document_test.go            [port mongogit/document_test.go]
├── query_test.go               [port mongogit/query_test.go]
├── collection_test.go          [port mongogit/collection_test.go — needs Mongo]
├── version/                    [port <reearthx>/asset/domain/version/*]
│   ├── version.go ref.go value.go version_or_ref.go query.go values.go
│   └── *_test.go
└── internal/
    ├── clock/clock.go          Now(), Mock()                                      [new]
    └── mongotest/mongotest.go  Connect(t)                                         [port mongox/mongotest/test.go]
```

---

## Task 1: Repository scaffolding

**Files:**
- Create: `/Users/dexter/active/mongogit/go.mod`
- Create: `/Users/dexter/active/mongogit/LICENSE`
- Create: `/Users/dexter/active/mongogit/README.md`
- Create: `/Users/dexter/active/mongogit/.gitignore`

**Interfaces:**
- Produces: module `github.com/reearth/mongogit` initialized as a git repo.

- [ ] **Step 1: Create the repo and git init**

```bash
mkdir -p /Users/dexter/active/mongogit
cd /Users/dexter/active/mongogit
git init
```

- [ ] **Step 2: Write `go.mod`**

```
module github.com/reearth/mongogit

go 1.26.0

require (
	github.com/chrispappas/golang-generics-set v1.0.1
	github.com/google/uuid v1.6.0
	github.com/samber/lo v1.50.0
	go.mongodb.org/mongo-driver v1.17.3
)
```

- [ ] **Step 3: Write `LICENSE`** — the standard MIT License text, with copyright line:

```
MIT License

Copyright (c) Re:Earth contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
... (full standard MIT body) ...
```

(Use the verbatim OSI MIT text; only the copyright line is customized.)

- [ ] **Step 4: Write `.gitignore`**

```
*.test
*.out
/coverage.txt
.DS_Store
```

- [ ] **Step 5: Write `README.md`** (initial; finalized in Task 17)

```markdown
# mongogit

Git-style versioned documents for MongoDB, in Go.

`mongogit` stores documents with version history — each write creates a new
immutable version with parents and named refs (like `latest`), so you can read
any version or ref of a document and walk its history.

## Install

```
go get github.com/reearth/mongogit
```

## Status

Extracted from [reearthx](https://github.com/reearth/reearthx). MIT licensed.
```

- [ ] **Step 6: Commit**

```bash
cd /Users/dexter/active/mongogit
git add -A
git commit -m "chore: initialize module"
```

---

## Task 2: Internal clock

**Files:**
- Create: `/Users/dexter/active/mongogit/internal/clock/clock.go`
- Test: `/Users/dexter/active/mongogit/internal/clock/clock_test.go`

**Interfaces:**
- Produces: `clock.Now() time.Time`; `clock.Mock(t time.Time) func()` (returns a reset function). Mirrors reearthx `util.Now`/`util.MockNow` semantics.

- [ ] **Step 1: Write the failing test**

```go
package clock

import (
	"testing"
	"time"
)

func TestNowMock(t *testing.T) {
	fixed := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	reset := Mock(fixed)
	if got := Now(); !got.Equal(fixed) {
		t.Fatalf("want %v, got %v", fixed, got)
	}
	reset()
	if Now().Equal(fixed) {
		t.Fatalf("expected real clock after reset")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/dexter/active/mongogit && go test ./internal/clock/`
Expected: FAIL — `undefined: Now`, `undefined: Mock`.

- [ ] **Step 3: Write `clock.go`**

```go
package clock

import (
	"sync"
	"time"
)

var (
	mu sync.Mutex
	fn func() time.Time
)

func Now() time.Time {
	mu.Lock()
	defer mu.Unlock()
	if fn == nil {
		return time.Now()
	}
	return fn()
}

func Mock(t time.Time) func() {
	mu.Lock()
	defer mu.Unlock()
	fn = func() time.Time { return t }
	return func() {
		mu.Lock()
		defer mu.Unlock()
		fn = nil
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/dexter/active/mongogit && go test ./internal/clock/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/clock
git commit -m "feat: add internal clock"
```

---

## Task 3: version package — clean files

**Files:**
- Create: `version/version.go`, `version/ref.go`, `version/value.go`, `version/version_or_ref.go`, `version/query.go`
- Create (tests): the matching `*_test.go` files for the above

**Interfaces:**
- Produces (package `version`): `Version`, `New()`, `Zero`, `Ref`, `Latest`, `Public`, `Versions`, `Refs`, `NewVersions`, `NewRefs`, `Value[T]`, `NewValue`, `MustBeValue`, `ValueFrom`, `VersionOrRef`, `MatchVersionOrRef`, `Query`, `All()`, `Eq()`, `QueryMatch`. (Used by every later task that touches `version.*`.)

- [ ] **Step 1: Port the five clean source files verbatim**

These five files import **no** reearthx packages — copy byte-for-byte, package stays `version`:
- `<reearthx>/asset/domain/version/version.go` → `version/version.go`
- `<reearthx>/asset/domain/version/ref.go` → `version/ref.go`
- `<reearthx>/asset/domain/version/value.go` → `version/value.go`
- `<reearthx>/asset/domain/version/version_or_ref.go` → `version/version_or_ref.go`
- `<reearthx>/asset/domain/version/query.go` → `version/query.go`

- [ ] **Step 2: Port the matching test files**

Copy `version_test.go`, `ref_test.go`, `value_test.go`, `version_or_ref_test.go`, `query_test.go` from `<reearthx>/asset/domain/version/` verbatim. If any imports `…/util` or `…/rerror`, leave those aside — they belong to `values_test.go` (Task 4); the five here are self-contained.

- [ ] **Step 3: Verify build/tests for what exists so far**

Run: `cd /Users/dexter/active/mongogit && go test ./version/ 2>&1 | head -40`
Expected: compile errors ONLY about `values.go`/`values_test.go` symbols not yet present (e.g. `Values`, `ErrArchived`). The five ported files and their tests must not themselves error. (If `values_test.go` was accidentally copied, remove it until Task 4.)

- [ ] **Step 4: Commit**

```bash
git add version
git commit -m "feat(version): port version object model (clean files)"
```

---

## Task 4: version package — values.go (shed reearthx)

**Files:**
- Create: `version/values.go`
- Create: `version/values_test.go`

**Interfaces:**
- Consumes: everything from Task 3.
- Produces: `Values[T]`, `NewValues`, `MustBeValues`, `UnwrapValues`, `var ErrArchived error`.

- [ ] **Step 1: Port `values.go` with substitutions**

Copy `<reearthx>/asset/domain/version/values.go` → `version/values.go`, then:

1. Replace the import block. Old:
```go
import (
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
)
```
New:
```go
import (
	"errors"

	"github.com/chrispappas/golang-generics-set/set"
	"github.com/reearth/mongogit/internal/clock"
	"github.com/samber/lo"
)
```
2. Replace the `ErrArchived` line:
```go
var ErrArchived = rerror.NewE(i18n.T("archived"))
```
with:
```go
var ErrArchived = errors.New("archived")
```
3. In `Add`, replace `t := util.Now()` with `t := clock.Now()`.
4. In `cloneValues`, replace:
```go
return util.Map(values, func(v *Value[V]) *Value[V] { return v.Clone() })
```
with:
```go
return lo.Map(values, func(v *Value[V], _ int) *Value[V] { return v.Clone() })
```

- [ ] **Step 2: Port `values_test.go` with substitutions**

Copy `<reearthx>/asset/domain/version/values_test.go` → `version/values_test.go`, then apply the symbol table: `util.MockNow(` → `clock.Mock(`, `util.Now()` → `clock.Now()`, and add `"github.com/reearth/mongogit/internal/clock"` to imports (remove the `…/util` import). If the test references `rerror`/`i18n`, replace assertions on `version.ErrArchived` to compare with the sentinel via `errors.Is` (it already is a plain error).

- [ ] **Step 3: Run version tests**

Run: `cd /Users/dexter/active/mongogit && go test ./version/...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add version
git commit -m "feat(version): port values with library-owned errors and clock"
```

---

## Task 5: errors + consumers + bson helpers

**Files:**
- Create: `errors.go`, `consumer.go`, `bson.go`
- Test: `bson_test.go`, `consumer_test.go`

**Interfaces:**
- Produces (package `mongogit`):
  - `var ErrNotFound, ErrAlreadyExists, ErrInvalidParams, ErrInternal error`; `func ErrInternalBy(err error) error`.
  - `Consumer` interface (`Consume(bson.Raw) error`); `FuncConsumer`, `SimpleConsumer[T]`, `SliceConsumer[T]`, `SliceRawFuncConsumer[T]`, `NewSliceRawFuncConsumer`, `SliceFuncConsumer[T,K]`, `NewSliceFuncConsumer`, `BatchConsumer`.
  - unexported `and(filter any, key string, f any) any`, `appendE(f any, e ...bson.E) any`, `getE`, `appendI`.

- [ ] **Step 1: Write `errors.go`** (new)

```go
package mongogit

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidParams = errors.New("invalid params")
	ErrInternal      = errors.New("internal")
)

func ErrInternalBy(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrInternal, err)
}
```

- [ ] **Step 2: Port `consumer.go`**

Copy `<reearthx>/mongox/consumer.go` → `consumer.go` verbatim, then change `package mongox` → `package mongogit`. (No other reearthx symbols appear in this file.)

- [ ] **Step 3: Port `bson.go`** (subset of `mongox/util.go`)

Create `bson.go` containing **only** these functions from `<reearthx>/mongox/util.go`, renamed to unexported, package `mongogit`:
- `And` → `and`
- `GetE` → `getE`
- `AppendE` → `appendE`
- `AppendI` → `appendI`

Inside their bodies, update the internal calls: `GetE(` → `getE(`, `AppendI(` → `appendI(`, `AppendE(` → `appendE(`. Do **not** port `DToM`, `AddCondition`, `isEmptyCondition`, `IndentPrint` (unused). Imports: `"go.mongodb.org/mongo-driver/bson"` only.

- [ ] **Step 4: Write `bson_test.go`** (new — covers the rename)

```go
package mongogit

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestAnd(t *testing.T) {
	got := and(bson.M{"a": 1}, "b", bson.M{"$gt": 2})
	m, ok := got.(bson.M)
	if !ok || m["a"] != 1 {
		t.Fatalf("unexpected: %#v", got)
	}
	if _, ok := m["b"]; !ok {
		t.Fatalf("missing key b: %#v", got)
	}
}

func TestAppendE(t *testing.T) {
	got := appendE(bson.M{"a": 1}, bson.E{Key: "b", Value: 2})
	m := got.(bson.M)
	if m["a"] != 1 || m["b"] != 2 {
		t.Fatalf("unexpected: %#v", got)
	}
}
```

- [ ] **Step 5: Port `consumer_test.go`** (optional but recommended)

Copy `<reearthx>/mongox/consumer_test.go` → `consumer_test.go`, change `package mongox` → `package mongogit`. Remove any test that references symbols not ported.

- [ ] **Step 6: Build and test**

Run: `cd /Users/dexter/active/mongogit && go test ./...`
Expected: PASS (version + clock + new files compile and pass).

- [ ] **Step 7: Commit**

```bash
git add errors.go consumer.go bson.go bson_test.go consumer_test.go
git commit -m "feat: add errors, consumers, and bson helpers"
```

---

## Task 6: pagination types

**Files:**
- Create: `pagination.go`
- Test: `pagination_test.go`

**Interfaces:**
- Produces (package `mongogit`): `Cursor` (string) + `CursorFromRef`, `Ref`, `CopyRef`, `StringRef`; `CursorPagination`, `OffsetPagination`, `Pagination` (+ `Clone`, `Wrap`); `Sort{Key string; Reverted bool}`; `PageInfo` (+ `NewPageInfo`, `EmptyPageInfo`, `OrEmpty`, `Clone`).

- [ ] **Step 1: Write `pagination.go`** (merge of `usecasex` cursor.go + pagination.go + pageinfo.go, shedding `util.CloneRef`)

```go
package mongogit

type Cursor string

func CursorFromRef(c *string) *Cursor {
	if c == nil {
		return nil
	}
	d := Cursor(*c)
	return &d
}

func (c Cursor) Ref() *Cursor {
	return &c
}

func (c *Cursor) CopyRef() *Cursor {
	if c == nil {
		return nil
	}
	d := *c
	return &d
}

func (c *Cursor) StringRef() *string {
	if c == nil {
		return nil
	}
	s := string(*c)
	return &s
}

// CursorPagination is Relay-style cursor pagination.
type CursorPagination struct {
	Before *Cursor
	After  *Cursor
	First  *int64
	Last   *int64
}

func (p *CursorPagination) Clone() *CursorPagination {
	if p == nil {
		return nil
	}
	return &CursorPagination{
		Before: p.Before.CopyRef(),
		After:  p.After.CopyRef(),
		First:  cloneInt64(p.First),
		Last:   cloneInt64(p.Last),
	}
}

func (p CursorPagination) Wrap() *Pagination {
	return &Pagination{Cursor: &p}
}

type OffsetPagination struct {
	Offset int64
	Limit  int64
}

func (p OffsetPagination) Wrap() *Pagination {
	return &Pagination{Offset: &p}
}

type Pagination struct {
	Cursor *CursorPagination
	Offset *OffsetPagination
}

func (p *Pagination) Clone() *Pagination {
	if p == nil {
		return nil
	}
	var offset *OffsetPagination
	if p.Offset != nil {
		o := *p.Offset
		offset = &o
	}
	return &Pagination{
		Cursor: p.Cursor.Clone(),
		Offset: offset,
	}
}

type Sort struct {
	Key      string
	Reverted bool
}

type PageInfo struct {
	TotalCount      int64
	StartCursor     *Cursor
	EndCursor       *Cursor
	HasNextPage     bool
	HasPreviousPage bool
}

func NewPageInfo(totalCount int64, startCursor, endCursor *Cursor, hasNextPage, hasPreviousPage bool) *PageInfo {
	return &PageInfo{
		TotalCount:      totalCount,
		StartCursor:     startCursor.CopyRef(),
		EndCursor:       endCursor.CopyRef(),
		HasNextPage:     hasNextPage,
		HasPreviousPage: hasPreviousPage,
	}
}

func EmptyPageInfo() *PageInfo {
	return &PageInfo{}
}

func (p *PageInfo) OrEmpty() *PageInfo {
	if p == nil {
		return EmptyPageInfo()
	}
	return p
}

func (p *PageInfo) Clone() *PageInfo {
	if p == nil {
		return nil
	}
	return &PageInfo{
		TotalCount:      p.TotalCount,
		StartCursor:     p.StartCursor.CopyRef(),
		EndCursor:       p.EndCursor.CopyRef(),
		HasNextPage:     p.HasNextPage,
		HasPreviousPage: p.HasPreviousPage,
	}
}

func cloneInt64(v *int64) *int64 {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}
```

- [ ] **Step 2: Write `pagination_test.go`** (new)

```go
package mongogit

import "testing"

func TestNewPageInfo(t *testing.T) {
	c := Cursor("x")
	pi := NewPageInfo(5, c.Ref(), c.Ref(), true, false)
	if pi.TotalCount != 5 || !pi.HasNextPage || pi.HasPreviousPage {
		t.Fatalf("unexpected: %#v", pi)
	}
	if pi.StartCursor == nil || *pi.StartCursor != c {
		t.Fatalf("bad start cursor: %#v", pi.StartCursor)
	}
}

func TestPaginationClone(t *testing.T) {
	first := int64(3)
	p := CursorPagination{First: &first}.Wrap()
	cl := p.Clone()
	if cl.Cursor.First == p.Cursor.First {
		t.Fatalf("expected First to be deep-copied")
	}
	if *cl.Cursor.First != 3 {
		t.Fatalf("unexpected value: %v", *cl.Cursor.First)
	}
}
```

- [ ] **Step 3: Test**

Run: `cd /Users/dexter/active/mongogit && go test ./...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add pagination.go pagination_test.go
git commit -m "feat: add pagination types"
```

---

## Task 7: index types

**Files:**
- Create: `index.go`
- Test: `index_test.go`

**Interfaces:**
- Produces (package `mongogit`): `Index{Name string; Key bson.D; Unique bool; CaseInsensitive bool; ExpireAfterSeconds *int32; Filter bson.M}`; methods `Normalize() Index`, `Model() mongo.IndexModel`, `Equal(Index) bool`; `IndexList []Index` with `Names`, `Models`, `Normalize`, `AddNamePrefix`, `RemoveDefaultIndex`.

- [ ] **Step 1: Port `index.go`** (subset of `mongox/index_model.go`)

Copy `<reearthx>/mongox/index_model.go` → `index.go`, then:
1. `package mongox` → `package mongogit`.
2. Remove `util.DiffResult` coupling: delete the `IndexResult` type and its three methods (`AddedNames`, `UpdatedNames`, `DeletedNames`) — they depend on `util.DiffResult` and are only used by the index-diff manager, which mongogit does not need.
3. Remove the `"github.com/reearth/reearthx/util"` import.
4. Keep: `prefix`, `Index`, `IndexFromKey`, `IndexFromKeys`, `TTLIndexFromKey`, `CaseInsensitiveIndexFromKey`, `toKeyBSON`, `Normalize`, `Model`, `Equal`, `IndexList` + its methods `Names`, `NamesWithoutPrefix`, `Models`, `Normalize`, `AddNamePrefix`, `RemoveDefaultIndex`. Imports become: `reflect`, `strings`, `github.com/samber/lo`, the three `mongo-driver` packages.

- [ ] **Step 2: Write `index_test.go`** (new)

```go
package mongogit

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestIndexModel(t *testing.T) {
	i := Index{Name: "x", Key: bson.D{{Key: "a", Value: 1}}, Unique: true}
	m := i.Model()
	if m.Options == nil || m.Options.Name == nil || *m.Options.Name != "x" {
		t.Fatalf("bad model: %#v", m.Options)
	}
}

func TestIndexEqual(t *testing.T) {
	a := Index{Name: "x", Key: bson.D{{Key: "a", Value: 1}}}
	b := Index{Name: "x", Key: bson.D{{Key: "a", Value: 1}}}
	if !a.Equal(b) {
		t.Fatalf("expected equal")
	}
}
```

- [ ] **Step 3: Test**

Run: `cd /Users/dexter/active/mongogit && go test ./...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add index.go index_test.go
git commit -m "feat: add index types"
```

---

## Task 8: mongoCollection — the mongo plumbing

**Files:**
- Create: `mongo.go`

**Interfaces:**
- Consumes: `Consumer`, `and`, `Sort`, `Pagination`, `Cursor`, `PageInfo`, `NewPageInfo`, `ErrNotFound`, `ErrInternalBy`.
- Produces (package `mongogit`, unexported): `type mongoCollection struct{ collection *mongo.Collection }`; `newMongoCollection(*mongo.Collection) *mongoCollection`; methods `Client() *mongo.Collection`, `Find(ctx, filter any, consumer Consumer, opts ...*options.FindOptions) error`, `FindOne(ctx, filter any, consumer Consumer, opts ...*options.FindOneOptions) error`, `Aggregate(ctx, pipeline []any, consumer Consumer, opts ...*options.AggregateOptions) error`, `Count(ctx, filter any) (int64, error)`, `CountAggregation(ctx, pipeline []any) (int64, error)`, `RemoveAll(ctx, filter any) error`, `Paginate(ctx, filter any, s *Sort, p *Pagination, consumer Consumer, opts ...*options.FindOptions) (*PageInfo, error)`, `PaginateAggregation(ctx, pipeline []any, s *Sort, p *Pagination, consumer Consumer, opts ...*options.AggregateOptions) (*PageInfo, error)`.

- [ ] **Step 1: Port the collection core** (subset of `mongox/collection.go`)

Into `mongo.go`, package `mongogit`, copy from `<reearthx>/mongox/collection.go`:
- `const idKey = "id"`
- the `findOptions` / `aggregateOptions` package vars
- `type Collection` → rename type to `mongoCollection` (field `collection *mongo.Collection`)
- `NewCollection` → rename to `newMongoCollection`
- methods `Client`, `Find`, `FindOne`, `Count`, `Aggregate`, `CountAggregation`, `RemoveAll` — copy verbatim onto `*mongoCollection`.
- **Do not** port `AggregateOne`, `RemoveOne`, `CreateOne`, `SaveOne`, `SetOne`, `SaveAll`, `UpdateMany`, `UpdateManyMany`, `Update` (unused by mongogit).
- `getCursor` — copy (used by pagination).
- `wrapError` — replace the whole function with:
```go
func wrapError(err error) error {
	return ErrInternalBy(err)
}
```
  and update every call site `wrapError(ctx, err)` → `wrapError(err)`.

Apply the symbol table throughout: `rerror.ErrNotFound` → `ErrNotFound`, `rerror.ErrAlreadyExists` → `ErrAlreadyExists` (only if a kept method uses it — `CountAggregation`/`Find`/`FindOne`/`Aggregate` use `ErrNotFound`), `usecasex.Cursor` → `Cursor`. Drop the `log` and `rerror` and `usecasex` imports. Resulting imports: `context`, `errors`, `fmt`, `io`, `go.mongodb.org/mongo-driver/bson`, `.../mongo`, `.../mongo/options`.

- [ ] **Step 2: Port the pagination methods** (subset of `mongox/pagination.go`) into the same `mongo.go`

Copy these onto `*mongoCollection` / as package funcs: `Paginate`, `paginate`, `PaginateAggregation`, `aggregateFilter`, `aggregateOptionsFromPagination`, `pageFilter`, `getCursorDocument`, `sortFilter`, `limit`, `sortDirection`, `consume`, `reverse`, `pageInfo`. **Do not** port `PaginateProject` (unused; it is the only user of `AddCondition`).

Apply substitutions:
- `*usecasex.Sort` → `*Sort`, `*usecasex.Pagination` → `*Pagination`, `usecasex.Pagination` → `Pagination`, `*usecasex.Cursor` → `*Cursor`, `usecasex.Cursor` → `Cursor`.
- `usecasex.NewPageInfo(` → `NewPageInfo(`.
- `And(` stays `and(` (the helper from `bson.go`). In `Paginate`, the line `filter = And(rawFilter, "", pFilter)` → `filter = and(rawFilter, "", pFilter)`.
- `rerror.ErrInternalByWithContext(ctx, ...)` → `ErrInternalBy(...)` (drop the `ctx` arg and the wrapping `fmt.Errorf` may stay as the inner error, e.g. `ErrInternalBy(fmt.Errorf("failed to find: %w", err))`).
- `c.collection.CountDocuments` and `c.collection.Aggregate`/`Find`/`FindOne` calls are unchanged (raw driver).

- [ ] **Step 3: Build**

Run: `cd /Users/dexter/active/mongogit && go build ./...`
Expected: success (no unresolved symbols). If `go vet` flags unused, remove the unused leftover.

- [ ] **Step 4: Commit**

```bash
git add mongo.go
git commit -m "feat: add mongo collection plumbing with pagination"
```

---

## Task 9: document.go

**Files:**
- Create: `document.go`

**Interfaces:**
- Consumes: `version.*`, `clock.Now`, `appendE`.
- Produces (package `mongogit`): `Document[T]` (+ `NewDocument`, `Value`, `MarshalBSON`, `UnmarshalBSON`); `Meta` (+ `Timestamp`); `ToValue[T]`; `MetadataDocument`; consts `versionKey`, `parentsKey`, `refsKey`, `metaKey`.

- [ ] **Step 1: Port `document.go`**

Copy `<reearthx>/asset/infrastructure/mongo/mongogit/document.go` → `document.go`, then:
1. Replace import block. Old:
```go
import (
	"time"

	"github.com/reearth/reearthx/asset/domain/version"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
```
New:
```go
import (
	"time"

	"github.com/reearth/mongogit/internal/clock"
	"github.com/reearth/mongogit/version"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
```
2. `util.Now()` → `clock.Now()` (two call sites: `NewDocument`, and within `Meta` construction).
3. In `Meta.apply`, `mongox.AppendE(` → `appendE(` (two calls).

- [ ] **Step 2: Build**

Run: `cd /Users/dexter/active/mongogit && go build ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add document.go
git commit -m "feat: add versioned document"
```

---

## Task 10: query.go

**Files:**
- Create: `query.go`

**Interfaces:**
- Consumes: `version.*`, `and`.
- Produces (package `mongogit`, unexported): `apply(q version.Query, f any) any`, `applyToPipeline(q version.Query, pipeline []any) []any`, `excludeMetadata(f any) any`.

- [ ] **Step 1: Port `query.go`**

Copy `<reearthx>/asset/infrastructure/mongo/mongogit/query.go` → `query.go`, then:
1. Replace import block. Old:
```go
import (
	"github.com/reearth/reearthx/asset/domain/version"
	"github.com/reearth/reearthx/mongox"
	"go.mongodb.org/mongo-driver/bson"
)
```
New:
```go
import (
	"github.com/reearth/mongogit/version"
	"go.mongodb.org/mongo-driver/bson"
)
```
2. `mongox.And(` → `and(` (in `apply` and `excludeMetadata`).

- [ ] **Step 2: Build**

Run: `cd /Users/dexter/active/mongogit && go build ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add query.go
git commit -m "feat: add version-aware query application"
```

---

## Task 11: collection.go (public API)

**Files:**
- Create: `collection.go`

**Interfaces:**
- Consumes: `mongoCollection`, `Document`, `Meta`, `MetadataDocument`, `apply`, `applyToPipeline`, `and`, `SliceConsumer`, `Consumer`, `Index`, `Sort`, `Pagination`, `PageInfo`, `version.*`, `clock.Now`, `ErrNotFound`, `ErrInvalidParams`, `ErrInternalBy`.
- Produces (package `mongogit`): `type Collection`; `NewCollection(*mongo.Collection) *Collection`; `Client() *mongo.Collection`; and the git-aware methods `FindOne`, `Find`, `Aggregate`, `Paginate`, `Count`, `PaginateAggregation`, `CountAggregation`, `SaveOne`, `SaveMany`, `SaveAll`, `DeleteRef`, `UpdateRef`, `IsArchived`, `ArchiveOne`, `Timestamp`, `RemoveOne`, `Empty`, `Indexes`, and unexported `meta`/`metas`.

- [ ] **Step 1: Port `collection.go`**

Copy `<reearthx>/asset/infrastructure/mongo/mongogit/collection.go` → `collection.go`, then:

1. Replace the import block. Old:
```go
import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/reearth/reearthx/asset/domain/version"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)
```
New:
```go
import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/reearth/mongogit/internal/clock"
	"github.com/reearth/mongogit/version"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
```
(Note: add `"go.mongodb.org/mongo-driver/mongo"` for the new `NewCollection` signature.)

2. Replace the type + constructor + `Client`:
```go
type Collection struct {
	client *mongox.Collection
}

func NewCollection(client *mongox.Collection) *Collection {
	return &Collection{client: client}
}

func (c *Collection) Client() *mongox.Collection {
	return c.client
}
```
with:
```go
type Collection struct {
	client *mongoCollection
}

func NewCollection(c *mongo.Collection) *Collection {
	return &Collection{client: newMongoCollection(c)}
}

func (c *Collection) Client() *mongo.Collection {
	return c.client.Client()
}
```

3. Apply the symbol table to the rest of the body:
   - `mongox.Consumer` → `Consumer` (params of `FindOne`, `Find`, `Aggregate`, `Paginate`, `PaginateAggregation`).
   - `*usecasex.Sort` → `*Sort`, `*usecasex.Pagination` → `*Pagination`, `*usecasex.PageInfo` → `*PageInfo`.
   - `mongox.SliceConsumer[` → `SliceConsumer[` (in `IsArchived`, `Timestamp`, `meta`, `metas`).
   - `mongox.And(` → `and(` (in `IsArchived`, `ArchiveOne`).
   - `mongox.Index` → `Index` (return type of `Indexes`).
   - `rerror.ErrNotFound` → `ErrNotFound`; `rerror.ErrInvalidParams` → `ErrInvalidParams`; `rerror.ErrInternalBy(` → `ErrInternalBy(`.
   - `util.Now()` → `clock.Now()`.
   - `version.*` unchanged (now resolves to `github.com/reearth/mongogit/version`).
   - `c.client.Client().InsertOne/InsertMany/UpdateMany/UpdateOne/DeleteOne/ReplaceOne/Drop` are unchanged — `c.client.Client()` returns the raw `*mongo.Collection`.
   - The `c.client.FindOne/Find/Aggregate/Paginate/Count/PaginateAggregation/CountAggregation/RemoveAll` calls are unchanged — `mongoCollection` exposes the same method names.

- [ ] **Step 2: Build + vet**

Run: `cd /Users/dexter/active/mongogit && go build ./... && go vet ./...`
Expected: success. Resolve any unused-import error by trimming.

- [ ] **Step 3: Commit**

```bash
git add collection.go
git commit -m "feat: add versioned mongo collection"
```

---

## Task 12: mongotest helper

**Files:**
- Create: `internal/mongotest/mongotest.go`

**Interfaces:**
- Produces: `var Env, Database string`; `func Connect(t *testing.T) func(*testing.T) *mongo.Database`. Skips the test when the env var named by `Env` is unset/empty.

- [ ] **Step 1: Port `mongotest`**

Copy `<reearthx>/mongox/mongotest/test.go` → `internal/mongotest/mongotest.go` verbatim (package stays `mongotest`). It has no reearthx imports. No changes needed.

- [ ] **Step 2: Build**

Run: `cd /Users/dexter/active/mongogit && go build ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/mongotest
git commit -m "test: add mongotest connect helper"
```

---

## Task 13: port document & query tests

**Files:**
- Create: `document_test.go`, `query_test.go`

**Interfaces:**
- Consumes: package `mongogit` symbols + `version.*`.

- [ ] **Step 1: Port `document_test.go`**

Copy `<reearthx>/asset/infrastructure/mongo/mongogit/document_test.go` → `document_test.go`. Change `package mongogit` stays. Update imports: drop `…/asset/domain/version` → add `github.com/reearth/mongogit/version`; replace any `rerror.ErrX` → `ErrX`, `mongox.X` → `X`, `usecasex.X` → `X`, `util.Now`→`clock.Now` / `util.MockNow`→`clock.Mock` (add `internal/clock` import if used).

- [ ] **Step 2: Port `query_test.go`**

Copy `<reearthx>/asset/infrastructure/mongo/mongogit/query_test.go` → `query_test.go`, applying the same substitutions.

- [ ] **Step 3: Run (these are pure unit tests — no Mongo)**

Run: `cd /Users/dexter/active/mongogit && go test -run 'Document|Query' ./...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add document_test.go query_test.go
git commit -m "test: port document and query tests"
```

---

## Task 14: port collection tests (Mongo-backed)

**Files:**
- Create: `collection_test.go`

**Interfaces:**
- Consumes: `internal/mongotest`, package `mongogit` symbols, `version.*`.

- [ ] **Step 1: Port `collection_test.go`**

Copy `<reearthx>/asset/infrastructure/mongo/mongogit/collection_test.go` → `collection_test.go`, then:
1. Imports: drop `…/asset/domain/version`, `…/mongox`, `…/mongox/mongotest`, `…/rerror`, `…/usecasex`; add `github.com/reearth/mongogit/version` and `github.com/reearth/mongogit/internal/mongotest`. Keep `github.com/samber/lo`.
2. `mongotest.Env = "REEARTH_DB"` → `mongotest.Env = "MONGO_URI"`.
3. Collection construction: replace the `mongox` client/collection setup. The current pattern obtains a `*mongo.Database` via `mongotest.Connect(t)(t)` and wraps it with `mongox.NewClientWithDatabase(db).WithCollection("name")`. Replace with the raw driver collection:
   - `mongox.NewClientWithDatabase(db).WithCollection("test")` → `db.Collection("test")`
   - `NewCollection(<that>)` now receives a `*mongo.Collection` directly.
4. Apply the symbol table to the body: `mongox.SliceConsumer[` → `SliceConsumer[`; `rerror.ErrNotFound` → `ErrNotFound`; `usecasex.Cursor` → `Cursor`, `usecasex.CursorPagination` → `CursorPagination`, `usecasex.NewPageInfo(` → `NewPageInfo(`; `version.*` unchanged (new import path). `lo.Must`, `lo.Must0`, `lo.ToPtr` unchanged.

- [ ] **Step 2: Run without Mongo (must skip, not fail)**

Run: `cd /Users/dexter/active/mongogit && go test ./...`
Expected: collection tests SKIP ("no db uri was provided"); everything else PASS. Build must be clean.

- [ ] **Step 3: Run with Mongo**

```bash
docker run -d --rm -p 27017:27017 --name mongogit-test mongo:6
cd /Users/dexter/active/mongogit
MONGO_URI="mongodb://localhost:27017" go test ./...
docker stop mongogit-test
```
Expected: all tests PASS (including collection tests).

- [ ] **Step 4: Commit**

```bash
git add collection_test.go
git commit -m "test: port mongo-backed collection tests"
```

---

## Task 15: linter config

**Files:**
- Create: `.golangci.yml`

- [ ] **Step 1: Write `.golangci.yml`**

```yaml
run:
  timeout: 5m

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - ineffassign
    - misspell
    - staticcheck
    - unconvert
    - unused

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - unused
```

- [ ] **Step 2: Run the linter (if installed)**

Run: `cd /Users/dexter/active/mongogit && golangci-lint run ./... || true`
Expected: clean, or only trivial findings to fix inline.

- [ ] **Step 3: Commit**

```bash
git add .golangci.yml
git commit -m "chore: add golangci-lint config"
```

---

## Task 16: CI workflow

**Files:**
- Create: `.github/workflows/ci.yml`

**Interfaces:**
- Produces: a workflow with `lint` and `test` jobs; test job runs a MongoDB service container and sets `MONGO_URI` for the test step; all tunables from `vars`/`secrets` with literal fallbacks; no AI/tooling provenance anywhere.

- [ ] **Step 1: Write `.github/workflows/ci.yml`**

```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ vars.GO_VERSION || '1.26.x' }}
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: ${{ vars.GOLANGCI_LINT_VERSION || 'latest' }}

  test:
    runs-on: ubuntu-latest
    services:
      mongo:
        image: ${{ vars.MONGO_IMAGE || 'mongo:6' }}
        ports:
          - 27017:27017
    env:
      MONGO_URI: mongodb://localhost:27017
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ vars.GO_VERSION || '1.26.x' }}
      - name: test
        run: go test -race -coverprofile=coverage.txt ./...
      - name: upload coverage
        if: ${{ secrets.CODECOV_TOKEN != '' }}
        uses: codecov/codecov-action@v4
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          files: coverage.txt
```

(`MONGO_URI` points at the service container — a fixed local address, not a secret. The image tag, Go version, and lint version come from `vars`; the Codecov token from `secrets`, and the upload step is skipped when it is absent.)

- [ ] **Step 2: Validate YAML locally**

Run: `cd /Users/dexter/active/mongogit && python3 -c "import yaml,sys; yaml.safe_load(open('.github/workflows/ci.yml'))" && echo OK`
Expected: `OK`.

- [ ] **Step 3: Confirm no provenance strings in `.github`**

Run: `cd /Users/dexter/active/mongogit && grep -riE "claude|superpower|copilot|generated by|ai-" .github || echo CLEAN`
Expected: `CLEAN`.

- [ ] **Step 4: Commit**

```bash
git add .github
git commit -m "ci: add lint and test workflow"
```

---

## Task 17: finalize — comments, README, tidy, full verification

**Files:**
- Modify: all source comments; `README.md`; `go.mod`/`go.sum`

- [ ] **Step 1: Clean comments with ai-slop-cleaner**

Invoke the `oh-my-claudecode:ai-slop-cleaner` skill over every `.go` file added in this plan. Remove narration/restatement comments; keep only comments that explain non-obvious intent (e.g. the relay pagination note in `mongo.go`). Apply its findings.

- [ ] **Step 2: Expand `README.md`**

Add a usage example (construct `mongogit.NewCollection(db.Collection("items"))`, `SaveOne`, `FindOne` with `version.Latest`), a "Running tests" section documenting `MONGO_URI`, and a license line. Keep it factual; no provenance/tooling mentions.

- [ ] **Step 3: Tidy and verify the whole module**

```bash
cd /Users/dexter/active/mongogit
go mod tidy
gofmt -l .          # expect no output
go vet ./...
go build ./...
grep -r "reearth/reearthx" . --include=*.go && echo "LEAK" || echo "NO REEARTHX IMPORTS"
```
Expected: `gofmt -l` prints nothing; vet/build clean; final grep prints `NO REEARTHX IMPORTS`.

- [ ] **Step 4: Full test run with Mongo**

```bash
docker run -d --rm -p 27017:27017 --name mongogit-test mongo:6
cd /Users/dexter/active/mongogit
MONGO_URI="mongodb://localhost:27017" go test -race ./...
docker stop mongogit-test
```
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "chore: finalize comments, README, and module metadata"
```

- [ ] **Step 6: Create the remote and push** (requires `reearth/mongogit` to exist on GitHub)

```bash
cd /Users/dexter/active/mongogit
git branch -M main
git remote add origin git@github.com:reearth/mongogit.git
git push -u origin main
```

---

## Self-Review

**Spec coverage** (spec §→task):
- §3 standalone/leaf → Tasks 5–12 (shed couplings); verified in Task 17 Step 3.
- §5 scope (version + mongogit, no mongox/memorygit) → Tasks 3–4 (version), 8–11 (mongogit); mongox/memorygit not ported.
- §6 shedding table → Task 5 (errors/consumers/bson), 6 (pagination), 7 (index), 2 (clock), 8 (mongo plumbing).
- §7 layout → File Structure section + all tasks.
- §8 license → Task 1 Step 3.
- §9 CI (vars/secrets, mongo service, no provenance) → Task 16.
- §10 testing (skip without DB) → Tasks 12, 14.
- §13 env rename `REEARTH_DB`→`MONGO_URI` → Task 14 Step 1.2, Task 16.

**Placeholder scan:** New-code files (clock, errors, pagination, CI, golangci) carry full content. Ports specify exact source path + transformations. No "TBD"/"handle errors"/"similar to" left.

**Type consistency:** `mongoCollection` exposes the same method names mongogit calls (`FindOne`/`Find`/`Aggregate`/`Paginate`/`PaginateAggregation`/`Count`/`CountAggregation`/`RemoveAll`/`Client`), so `collection.go` calls resolve unchanged. `Consumer`/`SliceConsumer`/`Sort`/`Pagination`/`PageInfo`/`Cursor`/`Index`/`ErrNotFound`/`ErrInvalidParams`/`ErrInternalBy`/`clock.Now` are defined before their consumers (Tasks 2,5,6,7 precede 8–11).

## Out of scope (separate plan)

The reearthx removal — deleting `asset/infrastructure/mongo/mongogit`, adding the `github.com/reearth/mongogit` dependency, repointing `asset/` version imports + `memorygit`, and writing the consumer-side adapters — is its own spec/plan (spec §11).
