package lark

import (
	"encoding/json"
	"log/slog"

	"github.com/nuln/agent-core"
)

// larkMessageLog is Lark's own data structure for recording incoming messages.
// It captures Lark-specific fields that Core never needs to know about.
type larkMessageLog struct {
	TraceID    string   `json:"trace_id"`
	EventID    string   `json:"event_id,omitempty"`
	MessageID  string   `json:"msg_id"`
	ChatID     string   `json:"chat_id"`
	ChatType   string   `json:"chat_type"`
	UserOpenID string   `json:"user_open_id"`
	UserName   string   `json:"user_name"`
	MsgType    string   `json:"msg_type,omitempty"`
	Content    string   `json:"content,omitempty"`
	ImageCount int      `json:"image_count,omitempty"`
	FileCount  int      `json:"file_count,omitempty"`
	HasAudio   bool     `json:"has_audio,omitempty"`
	CreateTime int64    `json:"create_time,omitempty"`
	ImagePaths []string `json:"image_paths,omitempty"`
	FilePaths  []string `json:"file_paths,omitempty"`
}

// SetStorage implements agent.StorageAware. Engine injects a scoped
// KVStoreProvider (namespace: plugins/feishu or plugins/lark) during registration.
func (p *LarkAccess) SetStorage(store agent.KVStoreProvider) {
	p.storage = store
}

// RecordMessage implements agent.DialogRecorder. It stores the incoming message
// in Lark's own format under plugins/<name>/messages/{trace_id}.
func (p *LarkAccess) RecordMessage(traceID string, msg *agent.Message) error {
	if p.storage == nil {
		return nil
	}
	store, err := p.storage.GetStore("messages")
	if err != nil {
		return err
	}

	// Extract chatID and chatType from replyCtx if available
	chatID := ""
	if rc, ok := msg.ReplyCtx.(replyContext); ok {
		chatID = rc.chatID
	}

	createTime := int64(0)
	if !msg.CreateTime.IsZero() {
		createTime = msg.CreateTime.UnixMilli()
	}

	log := larkMessageLog{
		TraceID:    traceID,
		MessageID:  msg.MessageID,
		ChatID:     chatID,
		UserOpenID: msg.UserID,
		UserName:   msg.UserName,
		Content:    msg.Content,
		ImageCount: len(msg.Images),
		FileCount:  len(msg.Files),
		HasAudio:   msg.Audio != nil,
		CreateTime: createTime,
	}

	data, err := json.Marshal(log)
	if err != nil {
		return err
	}
	if err := store.Put([]byte(traceID), data); err != nil {
		slog.Warn(p.tag()+": failed to record message", "error", err, "trace_id", traceID)
		return err
	}
	return nil
}
