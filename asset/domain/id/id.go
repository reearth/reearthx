package id

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/idx"
)

type (
	Workspace struct{}
	User      struct{}
	Asset     struct{}
	Event     struct{}
)

func (Workspace) Type() string { return "workspace" }
func (User) Type() string      { return "user" }
func (Asset) Type() string     { return "asset" }
func (Event) Type() string     { return "event" }

type (
	WorkspaceID = idx.ID[Workspace]
	UserID      = idx.ID[User]
	AssetID     = idx.ID[Asset]
	EventID     = idx.ID[Event]
)

var (
	NewWorkspaceID = idx.New[Workspace]
	NewUserID      = idx.New[User]
	NewAssetID     = idx.New[Asset]
	NewEventID     = idx.New[Event]
)

var (
	MustWorkspaceID = idx.Must[Workspace]
	MustUserID      = idx.Must[User]
	MustAssetID     = idx.Must[Asset]
	MustEventID     = idx.Must[Event]
)

var (
	WorkspaceIDFrom = idx.From[Workspace]
	UserIDFrom      = idx.From[User]
	AssetIDFrom     = idx.From[Asset]
	EventIDFrom     = idx.From[Event]
)

var (
	WorkspaceIDFromRef = idx.FromRef[Workspace]
	UserIDFromRef      = idx.FromRef[User]
	AssetIDFromRef     = idx.FromRef[Asset]
	EventIDFromRef     = idx.FromRef[Event]
)

type (
	WorkspaceIDList = idx.List[accountdomain.Workspace]
	UserIDList      = idx.List[accountdomain.User]
	AssetIDList     = idx.List[Asset]
)

var (
	WorkspaceIDListFrom = idx.ListFrom[accountdomain.Workspace]
	UserIDListFrom      = idx.ListFrom[accountdomain.User]
	AssetIDListFrom     = idx.ListFrom[Asset]
)

type (
	WorkspaceIDSet = idx.Set[Workspace]
	UserIDSet      = idx.Set[User]
	AssetIDSet     = idx.Set[Asset]
)

var (
	NewWorkspaceIDSet = idx.NewSet[Workspace]
	NewUserIDSet      = idx.NewSet[User]
	NewAssetIDSet     = idx.NewSet[Asset]
)

type Project struct{}

func (Project) Type() string { return "project" }

type (
	ProjectID     = idx.ID[Project]
	ProjectIDList = idx.List[Project]
)

var (
	MustProjectID     = idx.Must[Project]
	NewProjectID      = idx.New[Project]
	ProjectIDFrom     = idx.From[Project]
	ProjectIDFromRef  = idx.FromRef[Project]
	ProjectIDListFrom = idx.ListFrom[Project]
)

type Model struct{}

func (Model) Type() string { return "model" }

type (
	ModelID     = idx.ID[Model]
	ModelIDList = idx.List[Model]
)

var (
	MustModelID     = idx.Must[Model]
	NewModelID      = idx.New[Model]
	ModelIDFrom     = idx.From[Model]
	ModelIDFromRef  = idx.FromRef[Model]
	ModelIDListFrom = idx.ListFrom[Model]
)

type Field struct{}

func (Field) Type() string { return "field" }

type (
	FieldID     = idx.ID[Field]
	FieldIDList = idx.List[Field]
)

var (
	MustFieldID     = idx.Must[Field]
	NewFieldID      = idx.New[Field]
	FieldIDFrom     = idx.From[Field]
	FieldIDFromRef  = idx.FromRef[Field]
	FieldIDListFrom = idx.ListFrom[Field]
)

type Tag struct{}

func (Tag) Type() string { return "tag" }

type (
	TagID     = idx.ID[Tag]
	TagIDList = idx.List[Tag]
)

var (
	MustTagID     = idx.Must[Tag]
	NewTagID      = idx.New[Tag]
	TagIDFrom     = idx.From[Tag]
	TagIDFromRef  = idx.FromRef[Tag]
	TagIDListFrom = idx.ListFrom[Tag]
)

type Schema struct{}

func (Schema) Type() string { return "schema" }

type (
	SchemaID     = idx.ID[Schema]
	SchemaIDList = idx.List[Schema]
)

var (
	MustSchemaID     = idx.Must[Schema]
	NewSchemaID      = idx.New[Schema]
	SchemaIDFrom     = idx.From[Schema]
	SchemaIDFromRef  = idx.FromRef[Schema]
	SchemaIDListFrom = idx.ListFrom[Schema]
)

type Group struct{}

func (Group) Type() string { return "group" }

type (
	GroupID     = idx.ID[Group]
	GroupIDList = idx.List[Group]
)

var (
	MustGroupID     = idx.Must[Group]
	NewGroupID      = idx.New[Group]
	GroupIDFrom     = idx.From[Group]
	GroupIDFromRef  = idx.FromRef[Group]
	GroupIDListFrom = idx.ListFrom[Group]
)

type ItemGroup struct{}

func (ItemGroup) Type() string { return "item_group" }

type (
	ItemGroupID     = idx.ID[ItemGroup]
	ItemGroupIDList = idx.List[ItemGroup]
)

var (
	MustItemGroupID     = idx.Must[ItemGroup]
	NewItemGroupID      = idx.New[ItemGroup]
	ItemGroupIDFrom     = idx.From[ItemGroup]
	ItemGroupIDFromRef  = idx.FromRef[ItemGroup]
	ItemGroupIDListFrom = idx.ListFrom[ItemGroup]
)

type Thread struct{}

func (Thread) Type() string { return "thread" }

type (
	ThreadID     = idx.ID[Thread]
	ThreadIDList = idx.List[Thread]
)

var (
	NewThreadID     = idx.New[Thread]
	MustThreadID    = idx.Must[Thread]
	ThreadIDFrom    = idx.From[Thread]
	ThreadIDFromRef = idx.FromRef[Thread]
)

type Comment struct{}

func (Comment) Type() string { return "comment" }

type (
	CommentID     = idx.ID[Comment]
	CommentIDList = idx.List[Comment]
)

var (
	NewCommentID     = idx.New[Comment]
	MustCommentID    = idx.Must[Comment]
	CommentIDFrom    = idx.From[Comment]
	CommentIDFromRef = idx.FromRef[Comment]
)

type Item struct{}

func (Item) Type() string { return "item" }

type (
	ItemID     = idx.ID[Item]
	ItemIDList = idx.List[Item]
)

var (
	MustItemID     = idx.Must[Item]
	NewItemID      = idx.New[Item]
	ItemIDFrom     = idx.From[Item]
	ItemIDFromRef  = idx.FromRef[Item]
	ItemIDListFrom = idx.ListFrom[Item]
)

type Integration struct{}

func (Integration) Type() string { return "integration" }

type (
	IntegrationID     = idx.ID[Integration]
	IntegrationIDList = idx.List[Integration]
)

var (
	MustIntegrationID     = idx.Must[Integration]
	NewIntegrationID      = idx.New[Integration]
	IntegrationIDFrom     = idx.From[Integration]
	IntegrationIDFromRef  = idx.FromRef[Integration]
	IntegrationIDListFrom = idx.ListFrom[Integration]
)

type Webhook struct{}

func (Webhook) Type() string { return "webhook" }

type (
	WebhookID     = idx.ID[Webhook]
	WebhookIDList = idx.List[Webhook]
)

var (
	MustWebhookID     = idx.Must[Webhook]
	NewWebhookID      = idx.New[Webhook]
	WebhookIDFrom     = idx.From[Webhook]
	WebhookIDFromRef  = idx.FromRef[Webhook]
	WebhookIDListFrom = idx.ListFrom[Webhook]
)

type Task struct{}

func (Task) Type() string { return "task" }

type TaskID = idx.ID[Task]

var (
	NewTaskID     = idx.New[Task]
	MustTaskID    = idx.Must[Task]
	TaskIDFrom    = idx.From[Task]
	TaskIDFromRef = idx.FromRef[Task]
)

type TaskIDList = idx.List[Task]

var TaskIDListFrom = idx.ListFrom[Task]

type TaskIDSet = idx.Set[Task]

var NewTaskIDSet = idx.NewSet[Task]

type Request struct{}

func (Request) Type() string { return "request" }

type (
	RequestID     = idx.ID[Request]
	RequestIDList = idx.List[Request]
)

var (
	NewRequestID     = idx.New[Request]
	MustRequestID    = idx.Must[Request]
	RequestIDFrom    = idx.From[Request]
	RequestIDFromRef = idx.FromRef[Request]
)

type View struct{}

func (View) Type() string { return "view" }

type (
	ViewID     = idx.ID[View]
	ViewIDList = idx.List[View]
)

var (
	NewViewID     = idx.New[View]
	MustViewID    = idx.Must[View]
	ViewIDFrom    = idx.From[View]
	ViewIDFromRef = idx.FromRef[View]
)

type Resource struct{}

func (Resource) Type() string { return "resource" }

type (
	ResourceID     = idx.ID[Resource]
	ResourceIDList = idx.List[Resource]
)

var (
	NewResourceID     = idx.New[Resource]
	MustResourceID    = idx.Must[Resource]
	ResourceIDFrom    = idx.From[Resource]
	ResourceIDFromRef = idx.FromRef[Resource]
)
