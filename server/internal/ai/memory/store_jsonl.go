package memory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mayfly-go/pkg/logx"
	"mayfly-go/pkg/utils/filex"
)

const (
	numLockShards = 64
	maxLineSize   = 1 * 1024 * 1024 // 1 MB
)

// JSONLStore 基于 JSONL 文件的记忆存储实现
type JSONLStore struct {
	dir   string
	locks [numLockShards]sync.Mutex
}

var _ Store = (*JSONLStore)(nil)

// NewJSONLStore 创建 JSONL 存储实例
func NewJSONLStore(dir string) (*JSONLStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create memory directory: %w", err)
	}
	return &JSONLStore{dir: dir}, nil
}

// userLock 获取用户级别的锁
func (s *JSONLStore) userLock(userID string) *sync.Mutex {
	h := fnv.New32a()
	h.Write([]byte(userID))
	return &s.locks[h.Sum32()%numLockShards]
}

// filePath 获取用户记忆文件路径
func (s *JSONLStore) filePath(userID string) string {
	safeID := strings.ReplaceAll(userID, "/", "_")
	return filepath.Join(s.dir, safeID+"_memories.jsonl")
}

// readMemories 读取用户的所有记忆
func (s *JSONLStore) readMemories(userID string) ([]*MemoryItem, error) {
	path := s.filePath(userID)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return []*MemoryItem{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open memory file: %w", err)
	}
	defer f.Close()

	var items []*MemoryItem
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var item MemoryItem
		if err := json.Unmarshal(line, &item); err != nil {
			logx.Warnf("skip corrupt memory line: %v", err)
			continue
		}
		items = append(items, &item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan memory file: %w", err)
	}

	return items, nil
}

// writeMemories 原子写入用户的所有记忆
func (s *JSONLStore) writeMemories(userID string, items []*MemoryItem) error {
	var lines []string
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("marshal memory item: %w", err)
		}
		lines = append(lines, string(data))
	}

	content := strings.Join(lines, "\n")
	if len(content) > 0 {
		content += "\n"
	}

	return filex.WriteFileAtomic(s.filePath(userID), []byte(content), 0o644)
}

// GetByUser 根据用户ID和标签获取记忆
func (s *JSONLStore) GetByUser(ctx context.Context, userID string, tags []string) ([]*MemoryItem, error) {
	l := s.userLock(userID)
	l.Lock()
	defer l.Unlock()

	items, err := s.readMemories(userID)
	if err != nil {
		return nil, err
	}

	// 如果指定了标签，进行过滤
	if len(tags) > 0 {
		items = s.filterByTags(items, tags)
	}

	return items, nil
}

// Save 保存记忆（追加模式）
func (s *JSONLStore) Save(ctx context.Context, items []*MemoryItem) error {
	if len(items) == 0 {
		return nil
	}

	userID := items[0].UserID
	l := s.userLock(userID)
	l.Lock()
	defer l.Unlock()

	// 读取现有记忆
	existing, err := s.readMemories(userID)
	if err != nil {
		return err
	}

	// 为新记忆生成 ID 和时间戳
	now := time.Now()
	for i := range items {
		if items[i].ID == "" {
			items[i].ID = s.generateID(items[i])
		}
		items[i].CreatedAt = now
		items[i].UpdatedAt = now
	}

	// 追加新记忆
	all := append(existing, items...)
	return s.writeMemories(userID, all)
}

// Delete 删除指定的记忆
func (s *JSONLStore) Delete(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	l := s.userLock(userID)
	l.Lock()
	defer l.Unlock()

	items, err := s.readMemories(userID)
	if err != nil {
		return err
	}

	// 构建要删除的 ID 集合
	deleteIDs := make(map[string]bool)
	for _, id := range ids {
		deleteIDs[id] = true
	}

	// 过滤掉要删除的记忆
	var remaining []*MemoryItem
	for _, item := range items {
		if !deleteIDs[item.ID] {
			remaining = append(remaining, item)
		}
	}

	return s.writeMemories(userID, remaining)
}

// Search 语义搜索记忆（简化实现：基于关键词匹配）
// TODO: 未来可集成向量数据库实现真正的语义搜索
func (s *JSONLStore) Search(ctx context.Context, userID string, query string, limit int) ([]*MemoryItem, error) {
	l := s.userLock(userID)
	l.Lock()
	defer l.Unlock()

	items, err := s.readMemories(userID)
	if err != nil {
		return nil, err
	}

	// 简单实现：按创建时间倒序返回
	// TODO: 实现真正的语义相似度排序
	if limit > 0 && len(items) > limit {
		items = items[len(items)-limit:]
	}

	return items, nil
}

// filterByTags 根据标签过滤记忆
func (s *JSONLStore) filterByTags(items []*MemoryItem, tags []string) []*MemoryItem {
	if len(tags) == 0 {
		return items
	}

	var result []*MemoryItem
	for _, item := range items {
		if s.hasAnyTag(item.Tags, tags) {
			result = append(result, item)
		}
	}
	return result
}

// hasAnyTag 检查记忆是否包含任意一个目标标签
func (s *JSONLStore) hasAnyTag(itemTags []string, targetTags []string) bool {
	tagSet := make(map[string]bool)
	for _, tag := range itemTags {
		tagSet[tag] = true
	}
	for _, tag := range targetTags {
		if tagSet[tag] {
			return true
		}
	}
	return false
}

// generateID 生成记忆 ID（使用时间戳 + 随机数保证唯一性）
func (s *JSONLStore) generateID(item *MemoryItem) string {
	return fmt.Sprintf("%s_%d_%d", item.UserID, time.Now().UnixNano(), rand.Intn(10000))
}
