package lark

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nuln/agent-core"
	"github.com/stretchr/testify/assert"
)

type mockKVStore struct {
	data map[string][]byte
}

func (m *mockKVStore) Get(key []byte) ([]byte, error)   { return m.data[string(key)], nil }
func (m *mockKVStore) Put(key, value []byte) error      { m.data[string(key)] = value; return nil }
func (m *mockKVStore) Delete(key []byte) error          { delete(m.data, string(key)); return nil }
func (m *mockKVStore) List() (map[string][]byte, error) { return m.data, nil }

type mockKVStoreProvider struct {
	store *mockKVStore
}

func (m *mockKVStoreProvider) GetStore(name string) (agent.KVStore, error) {
	return m.store, nil
}

func TestDialogRecorder_Table(t *testing.T) {
	tests := []struct {
		name     string
		msg      *agent.Message
		validate func(t *testing.T, log larkMessageLog)
	}{
		{
			name: "standard text message",
			msg: &agent.Message{
				MessageID:  "m1",
				UserID:     "u1",
				UserName:   "John",
				Content:    "hello",
				CreateTime: time.Now(),
				ReplyCtx:   replyContext{chatID: "c1"},
			},
			validate: func(t *testing.T, log larkMessageLog) {
				assert.Equal(t, "m1", log.MessageID)
				assert.Equal(t, "u1", log.UserOpenID)
				assert.Equal(t, "c1", log.ChatID)
				assert.Equal(t, "hello", log.Content)
			},
		},
		{
			name: "rich text content",
			msg: &agent.Message{
				MessageID: "m2",
				Content:   "line1\nline2 🚀",
				ReplyCtx:  replyContext{chatID: "c2"},
			},
			validate: func(t *testing.T, log larkMessageLog) {
				assert.Equal(t, "line1\nline2 🚀", log.Content)
				assert.Equal(t, "c2", log.ChatID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockKVStore{data: make(map[string][]byte)}
			provider := &mockKVStoreProvider{store: store}

			lark := &LarkAccess{accessName: "lark"}
			lark.SetStorage(provider)

			traceID := "t-" + tt.name
			err := lark.RecordMessage(traceID, tt.msg)
			assert.NoError(t, err)

			data, ok := store.data[traceID]
			assert.True(t, ok)

			var log larkMessageLog
			json.Unmarshal(data, &log)
			tt.validate(t, log)
		})
	}
}
