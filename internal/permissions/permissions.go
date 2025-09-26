package permissions

import (
	"os"
	"sync"

	"github.com/bytedance/sonic"
)

// Manager menampung dan mengelola data izin dari file JSON.
type Manager struct {
	AllowedGroups map[string]bool `json:"allowedGroups"`
	AllowedUsers  map[string]bool `json:"allowedUsers"`

	mu       sync.RWMutex
	filePath string
}

// NewManager membuat instance baru dari permission manager.
func NewManager(path string) (*Manager, error) {
	m := &Manager{
		filePath:      path,
		AllowedGroups: make(map[string]bool),
		AllowedUsers:  make(map[string]bool),
	}

	file, err := os.ReadFile(path)
	// Jika file tidak ada, tidak apa-apa. File akan dibuat saat pertama kali menyimpan.
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}

	// Jika file ada, muat datanya.
	if err := sonic.Unmarshal(file, m); err != nil {
		return nil, err
	}

	return m, nil
}

// Save menyimpan data izin saat ini ke file JSON.
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := sonic.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0644)
}

// IsGroupAllowed memeriksa apakah ID grup ada di dalam daftar izin.
func (m *Manager) IsGroupAllowed(groupID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.AllowedGroups[groupID]
}

// IsUserAllowed memeriksa apakah ID pengguna ada di dalam daftar izin.
func (m *Manager) IsUserAllowed(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.AllowedUsers[userID]
}

// AddAllowedUser menambahkan pengguna ke daftar izin dan menyimpannya.
func (m *Manager) AddAllowedUser(userID string) error {
	m.mu.Lock()
	m.AllowedUsers[userID] = true
	m.mu.Unlock()
	return m.Save()
}

// RemoveAllowedUser menghapus pengguna dari daftar izin dan menyimpannya.
func (m *Manager) RemoveAllowedUser(userID string) error {
	m.mu.Lock()
	delete(m.AllowedUsers, userID)
	m.mu.Unlock()
	return m.Save()
}

// AddAllowedGroup menambahkan grup ke daftar izin dan menyimpannya.
func (m *Manager) AddAllowedGroup(groupID string) error {
	m.mu.Lock()
	m.AllowedGroups[groupID] = true
	m.mu.Unlock()
	return m.Save()
}

// RemoveAllowedGroup menghapus grup dari daftar izin dan menyimpannya.
func (m *Manager) RemoveAllowedGroup(groupID string) error {
	m.mu.Lock()
	delete(m.AllowedGroups, groupID)
	m.mu.Unlock()
	return m.Save()
}
