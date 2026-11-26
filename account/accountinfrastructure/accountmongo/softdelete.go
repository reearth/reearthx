package accountmongo

import "go.mongodb.org/mongo-driver/bson"

// reearth-accounts soft-deletes users and workspaces by setting deletedat. Documents written by reearthx
// (e.g. OSS single-DB setups) never have the field, so they always match.
var notDeleted = bson.M{"deletedat": bson.M{"$exists": false}}

// excludeDeleted narrows filter to documents that are not soft-deleted.
func excludeDeleted(filter any) bson.M {
	return bson.M{"$and": bson.A{filter, notDeleted}}
}
