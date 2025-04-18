package ds

import(
	"context"
)


// To prevent other libs colliding with us in the context.Value keyspace, use this private key
type contextKey int
const datastoreProviderKey contextKey = 0


// SetProvider embeds a provider inside a Context, for later retrieval
func SetProvider(ctx context.Context, p DatastoreProvider) context.Context {
	return context.WithValue(ctx, datastoreProviderKey, p)
}

// GetProvider retrieves the provider from the context; panics if not found
func GetProviderOrPanic(ctx context.Context) DatastoreProvider {
	p, ok := ctx.Value(datastoreProviderKey).(DatastoreProvider)
	if !ok { panic("GetDSProvider called on a context that had no DSPROvider") }
	return p
}
