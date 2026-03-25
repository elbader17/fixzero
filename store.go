package fixzero

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNotFound      = errors.New("message not found")
	ErrInvalidSeqNum = errors.New("invalid sequence number")
)

type MessageStore interface {
	SaveMessage(seqNum int, msg []byte) error
	GetMessage(seqNum int) ([]byte, error)
	GetRange(start, end int) ([][]byte, error)
	SetNextSenderSeqNum(num int)
	SetNextTargetSeqNum(num int)
	GetNextSenderSeqNum() int
	GetNextTargetSeqNum() int
	IncrNextSenderSeqNum()
	IncrNextTargetSeqNum()
	Refresh() error
	Close() error
}

var memoryBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 256)
		return &buf
	},
}

func getMemoryBuf() *[]byte {
	return memoryBufPool.Get().(*[]byte)
}

func putMemoryBuf(b *[]byte) {
	*b = (*b)[:0]
	memoryBufPool.Put(b)
}

type MemoryStore struct {
	mu            sync.RWMutex
	messages      map[int][]byte
	nextSenderSeq int
	nextTargetSeq int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		messages: make(map[int][]byte),
	}
}

func (s *MemoryStore) SaveMessage(seqNum int, msg []byte) error {
	if seqNum <= 0 {
		return ErrInvalidSeqNum
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	buf := getMemoryBuf()
	*buf = append(*buf, msg...)
	s.messages[seqNum] = *buf
	return nil
}

func (s *MemoryStore) GetMessage(seqNum int) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg, ok := s.messages[seqNum]
	if !ok {
		return nil, ErrNotFound
	}
	return msg, nil
}

func (s *MemoryStore) GetRange(start, end int) ([][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if start <= 0 || end < start {
		return nil, ErrInvalidSeqNum
	}

	result := make([][]byte, 0, end-start+1)
	for seq := start; seq <= end; seq++ {
		if msg, ok := s.messages[seq]; ok {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (s *MemoryStore) SetNextSenderSeqNum(num int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSenderSeq = num
}

func (s *MemoryStore) SetNextTargetSeqNum(num int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextTargetSeq = num
}

func (s *MemoryStore) GetNextSenderSeqNum() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nextSenderSeq
}

func (s *MemoryStore) GetNextTargetSeqNum() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nextTargetSeq
}

func (s *MemoryStore) IncrNextSenderSeqNum() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSenderSeq++
}

func (s *MemoryStore) IncrNextTargetSeqNum() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextTargetSeq++
}

func (s *MemoryStore) Refresh() error {
	return nil
}

func (s *MemoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = make(map[int][]byte)
	s.nextSenderSeq = 0
	s.nextTargetSeq = 0
	return nil
}

type sequences struct {
	NextSenderSeq int `json:"next_sender_seq"`
	NextTargetSeq int `json:"next_target_seq"`
}

type FileStore struct {
	mu            sync.RWMutex
	basePath      string
	nextSenderSeq int
	nextTargetSeq int
}

func NewFileStore(basePath string) (*FileStore, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	store := &FileStore{
		basePath:      basePath,
		nextSenderSeq: 1,
		nextTargetSeq: 1,
	}

	if err := store.loadSequences(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return store, nil
}

func (s *FileStore) msgPath(seqNum int) string {
	return filepath.Join(s.basePath, filepath.Base(s.basePath)+"_"+intToString(seqNum)+".msg")
}

func (s *FileStore) loadSequences() error {
	seqFile := filepath.Join(s.basePath, "sequence.json")
	data, err := os.ReadFile(seqFile)
	if err != nil {
		return err
	}

	var seq sequences
	if err := json.Unmarshal(data, &seq); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSenderSeq = seq.NextSenderSeq
	s.nextTargetSeq = seq.NextTargetSeq
	return nil
}

func (s *FileStore) saveSequences() error {
	s.mu.RLock()
	seq := sequences{
		NextSenderSeq: s.nextSenderSeq,
		NextTargetSeq: s.nextTargetSeq,
	}
	s.mu.RUnlock()

	data, err := json.Marshal(seq)
	if err != nil {
		return err
	}

	seqFile := filepath.Join(s.basePath, "sequence.json")
	return os.WriteFile(seqFile, data, 0644)
}

func (s *FileStore) SaveMessage(seqNum int, msg []byte) error {
	if seqNum <= 0 {
		return ErrInvalidSeqNum
	}

	path := s.msgPath(seqNum)
	return os.WriteFile(path, msg, 0644)
}

func (s *FileStore) GetMessage(seqNum int) ([]byte, error) {
	if seqNum <= 0 {
		return nil, ErrInvalidSeqNum
	}

	path := s.msgPath(seqNum)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

func (s *FileStore) GetRange(start, end int) ([][]byte, error) {
	if start <= 0 || end < start {
		return nil, ErrInvalidSeqNum
	}

	result := make([][]byte, 0, end-start+1)
	for seq := start; seq <= end; seq++ {
		if msg, err := s.GetMessage(seq); err == nil {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (s *FileStore) SetNextSenderSeqNum(num int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSenderSeq = num
}

func (s *FileStore) SetNextTargetSeqNum(num int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextTargetSeq = num
}

func (s *FileStore) GetNextSenderSeqNum() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nextSenderSeq
}

func (s *FileStore) GetNextTargetSeqNum() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nextTargetSeq
}

func (s *FileStore) IncrNextSenderSeqNum() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSenderSeq++
}

func (s *FileStore) IncrNextTargetSeqNum() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextTargetSeq++
}

func (s *FileStore) Refresh() error {
	return s.loadSequences()
}

func (s *FileStore) Close() error {
	return s.saveSequences()
}

type SQLiteStore struct {
	db            *sql.DB
	basePath      string
	mu            sync.RWMutex
	nextSenderSeq int
	nextTargetSeq int
}

func NewSQLiteStore(basePath string) (*SQLiteStore, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(basePath, "messages.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{
		db:       db,
		basePath: basePath,
	}

	if err := store.initTables(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) initTables() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			seq_num INTEGER PRIMARY KEY,
			data BLOB NOT NULL,
			created_at INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS sequences (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			next_sender_seq INTEGER NOT NULL DEFAULT 1,
			next_target_seq INTEGER NOT NULL DEFAULT 1
		);
		INSERT OR IGNORE INTO sequences (id, next_sender_seq, next_target_seq) VALUES (1, 1, 1);
	`)
	return err
}

func (s *SQLiteStore) loadSequences() (int, int, error) {
	var senderSeq, targetSeq int
	err := s.db.QueryRow("SELECT next_sender_seq, next_target_seq FROM sequences WHERE id = 1").Scan(&senderSeq, &targetSeq)
	if err != nil {
		return 0, 0, err
	}
	return senderSeq, targetSeq, nil
}

func (s *SQLiteStore) SaveMessage(seqNum int, msg []byte) error {
	if seqNum <= 0 {
		return ErrInvalidSeqNum
	}

	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO messages (seq_num, data, created_at) VALUES (?, ?, ?)",
		seqNum, msg, time.Now().Unix(),
	)
	return err
}

func (s *SQLiteStore) GetMessage(seqNum int) ([]byte, error) {
	if seqNum <= 0 {
		return nil, ErrInvalidSeqNum
	}

	var data []byte
	err := s.db.QueryRow("SELECT data FROM messages WHERE seq_num = ?", seqNum).Scan(&data)
	if err != nil {
		return nil, ErrNotFound
	}
	return data, nil
}

func (s *SQLiteStore) GetRange(start, end int) ([][]byte, error) {
	if start <= 0 || end < start {
		return nil, ErrInvalidSeqNum
	}

	rows, err := s.db.Query("SELECT data FROM messages WHERE seq_num >= ? AND seq_num <= ? ORDER BY seq_num", start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result [][]byte
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		result = append(result, data)
	}

	return result, rows.Err()
}

func (s *SQLiteStore) SetNextSenderSeqNum(num int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Exec("UPDATE sequences SET next_sender_seq = ? WHERE id = 1", num)
}

func (s *SQLiteStore) SetNextTargetSeqNum(num int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Exec("UPDATE sequences SET next_target_seq = ? WHERE id = 1", num)
}

func (s *SQLiteStore) GetNextSenderSeqNum() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var seq int
	s.db.QueryRow("SELECT next_sender_seq FROM sequences WHERE id = 1").Scan(&seq)
	return seq
}

func (s *SQLiteStore) GetNextTargetSeqNum() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var seq int
	s.db.QueryRow("SELECT next_target_seq FROM sequences WHERE id = 1").Scan(&seq)
	return seq
}

func (s *SQLiteStore) IncrNextSenderSeqNum() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Exec("UPDATE sequences SET next_sender_seq = next_sender_seq + 1 WHERE id = 1")
}

func (s *SQLiteStore) IncrNextTargetSeqNum() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.Exec("UPDATE sequences SET next_target_seq = next_target_seq + 1 WHERE id = 1")
}

func (s *SQLiteStore) Refresh() error {
	senderSeq, targetSeq, err := s.loadSequences()
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.nextSenderSeq = senderSeq
	s.nextTargetSeq = targetSeq
	s.mu.Unlock()
	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func intToString(i int) string {
	return string(rune(i))
}
