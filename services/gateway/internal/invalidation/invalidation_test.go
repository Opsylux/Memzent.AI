// services/gateway/internal/invalidation/invalidation_test.go
package invalidation

import (
	"context"
	"testing"
	"time"
)

// fakeStore is an in-memory Store for testing the Invalidator without Valkey.
type fakeStore struct {
	kv   map[string]string
	ints map[string]int64
	sets map[string]map[string]struct{}
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		kv:   map[string]string{},
		ints: map[string]int64{},
		sets: map[string]map[string]struct{}{},
	}
}

func (f *fakeStore) Incr(_ context.Context, key string) (int64, error) {
	f.ints[key]++
	return f.ints[key], nil
}
func (f *fakeStore) GetRaw(_ context.Context, key string) (string, error) {
	if v, ok := f.kv[key]; ok {
		return v, nil
	}
	if n, ok := f.ints[key]; ok {
		return itoa(n), nil
	}
	return "", nil
}
func (f *fakeStore) SetRaw(_ context.Context, key, value string, _ time.Duration) error {
	f.kv[key] = value
	return nil
}
func (f *fakeStore) SAdd(_ context.Context, key string, _ time.Duration, members ...string) error {
	if f.sets[key] == nil {
		f.sets[key] = map[string]struct{}{}
	}
	for _, m := range members {
		f.sets[key][m] = struct{}{}
	}
	return nil
}
func (f *fakeStore) SPopAll(_ context.Context, key string) ([]string, error) {
	out := make([]string, 0, len(f.sets[key]))
	for m := range f.sets[key] {
		out = append(out, m)
	}
	delete(f.sets, key)
	return out, nil
}
func (f *fakeStore) DelKeys(_ context.Context, keys ...string) (int64, error) {
	var n int64
	for _, k := range keys {
		if _, ok := f.kv[k]; ok {
			delete(f.kv, k)
			n++
		}
	}
	return n, nil
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

func TestVersion_DefaultAndBump(t *testing.T) {
	inv := New(newFakeStore(), nil, time.Minute)
	ctx := context.Background()

	if v := inv.Version(ctx, "org1"); v != "0" {
		t.Errorf("initial version = %q, want 0", v)
	}
	v, err := inv.Bump(ctx, "org1")
	if err != nil {
		t.Fatal(err)
	}
	if v != "1" {
		t.Errorf("bumped version = %q, want 1", v)
	}
	// The in-memory cache should reflect the bump immediately.
	if got := inv.Version(ctx, "org1"); got != "1" {
		t.Errorf("version after bump = %q, want 1", got)
	}
}

func TestVersion_NilSafe(t *testing.T) {
	var inv *Invalidator
	if v := inv.Version(context.Background(), "org1"); v != "" {
		t.Errorf("nil invalidator Version = %q, want empty", v)
	}
}

func TestInvalidateTool_TargetedBust(t *testing.T) {
	store := newFakeStore()
	inv := New(store, nil, time.Minute)
	ctx := context.Background()

	keys := []string{"org:o1:m:gpt:p:hello", "org:o1:m:gpt:c:hash"}
	for _, k := range keys {
		_ = store.SetRaw(ctx, k, "cached-value", time.Minute)
	}
	inv.TagKeys(ctx, "o1", []string{"github-repo"}, keys)

	deleted, err := inv.InvalidateTool(ctx, "o1", "github-repo")
	if err != nil {
		t.Fatal(err)
	}
	if deleted != int64(len(keys)) {
		t.Errorf("deleted = %d, want %d", deleted, len(keys))
	}
	for _, k := range keys {
		if v, _ := store.GetRaw(ctx, k); v != "" {
			t.Errorf("key %q should have been deleted", k)
		}
	}
	// Second invalidation is a no-op (index consumed).
	if deleted, _ := inv.InvalidateTool(ctx, "o1", "github-repo"); deleted != 0 {
		t.Errorf("second invalidation deleted = %d, want 0", deleted)
	}
}

func TestHandleEvent_ToolDataBustsKeys(t *testing.T) {
	store := newFakeStore()
	inv := New(store, nil, time.Minute)
	ctx := context.Background()

	_ = store.SetRaw(ctx, "k1", "v", time.Minute)
	inv.TagKeys(ctx, "o1", []string{"tool-a"}, []string{"k1"})

	res, err := inv.HandleEvent(ctx, InvalidationEvent{OrgID: "o1", ChangeType: ChangeToolData, ToolIDs: []string{"tool-a"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.KeysDeleted != 1 {
		t.Errorf("keys_deleted = %d, want 1", res.KeysDeleted)
	}
}

func TestHandleEvent_PolicyBumpsVersion(t *testing.T) {
	inv := New(newFakeStore(), nil, time.Minute)
	ctx := context.Background()

	res, err := inv.HandleEvent(ctx, InvalidationEvent{OrgID: "o1", ChangeType: ChangePolicy})
	if err != nil {
		t.Fatal(err)
	}
	if !res.VersionBumped || res.NewVersion != "1" {
		t.Errorf("expected version bump to 1, got %+v", res)
	}
}

func TestHandleEvent_UnknownType(t *testing.T) {
	inv := New(newFakeStore(), nil, time.Minute)
	if _, err := inv.HandleEvent(context.Background(), InvalidationEvent{OrgID: "o1", ChangeType: "bogus"}); err == nil {
		t.Error("expected error for unknown change_type")
	}
}
