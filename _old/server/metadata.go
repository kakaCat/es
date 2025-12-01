package main

import (
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "sync"
    "time"
)

type IndexMetadata struct {
    Name        string            `json:"name"`
    Namespace   string            `json:"namespace"`
    Tenant      string            `json:"tenant"`
    Dimension   int               `json:"dimension"`
    Metric      string            `json:"metric"`
    IVFParams   map[string]int    `json:"ivf_params"`
    SegmentSize int               `json:"segment_size"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type MetadataStore struct {
    path  string
    mu    sync.RWMutex
    items map[string]IndexMetadata
}

func NewMetadataStore(p string) *MetadataStore {
    return &MetadataStore{path: p, items: map[string]IndexMetadata{}}
}

func (s *MetadataStore) ensureDir() error {
    dir := filepath.Dir(s.path)
    return os.MkdirAll(dir, 0o755)
}

func (s *MetadataStore) Load() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if err := s.ensureDir(); err != nil { return err }
    b, err := os.ReadFile(s.path)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            s.items = map[string]IndexMetadata{}
            return nil
        }
        return err
    }
    var m map[string]IndexMetadata
    if err := json.Unmarshal(b, &m); err != nil { return err }
    s.items = m
    return nil
}

func (s *MetadataStore) Save() error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if err := s.ensureDir(); err != nil { return err }
    b, err := json.MarshalIndent(s.items, "", "  ")
    if err != nil { return err }
    return os.WriteFile(s.path, b, 0o644)
}

func (s *MetadataStore) Upsert(item IndexMetadata) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    now := time.Now()
    existing, ok := s.items[item.Name]
    if ok {
        item.CreatedAt = existing.CreatedAt
    } else {
        item.CreatedAt = now
    }
    item.UpdatedAt = now
    s.items[item.Name] = item
    return s.Save()
}

func (s *MetadataStore) GetAll() []IndexMetadata {
    s.mu.RLock()
    defer s.mu.RUnlock()
    res := make([]IndexMetadata, 0, len(s.items))
    for _, v := range s.items { res = append(res, v) }
    return res
}

func (s *MetadataStore) Get(name string) (IndexMetadata, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    v, ok := s.items[name]
    return v, ok
}

func (s *MetadataStore) Delete(name string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.items, name)
    return s.Save()
}
