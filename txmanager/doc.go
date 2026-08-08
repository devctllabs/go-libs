// Package txmanager defines the transaction boundary shared by service packages
// and database adapters.
//
// Services depend on Manager or Managers instead of repeating backend-specific
// transaction interfaces. Adapters create a Coordinator over their native
// transaction type and expose its reader and writer views. A transaction is
// carried in context and remains private to the adapter that knows how to route
// queries to it.
package txmanager
