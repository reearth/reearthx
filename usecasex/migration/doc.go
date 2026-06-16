// Package migration runs versioned database migrations keyed by an integer
// version that is tracked in a ConfigRepo.
//
// Two runners share the same Migrations/ConfigRepo machinery:
//
//   - Client is the original runner for backends that expose
//     usecasex.Transaction (Begin/Commit/End) — e.g. Mongo.
//   - Runner is the successor for backends that expose usecasex.Transactor
//     (WithinTransaction) — e.g. Postgres via pgxx.Client. New backends should
//     use Runner; Client remains for Mongo until it is migrated over.
//
// Both apply pending Migrations in ascending version order, each inside its own
// transaction. The ConfigRepo persists the current version: Begin/End wrap the
// whole run (e.g. to acquire and release a lock), Current reports the applied
// version, and Save records each newly-applied version.
package migration
