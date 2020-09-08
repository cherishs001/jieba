package jieba

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cherishs001/jieba"
)

var (
	jiebaInstancesMu = &sync.RWMutex{}
	jiebaInstances   = map[string]*JiebaInstance{}
)

const DictDirEnvName = "JIEBA_DICT_DIR"

// JiebaInstance is a thread-safe *jieba.Jieba for a given dict directory.
type JiebaInstance struct {
	dictDir  string
	mu       sync.RWMutex
	val      *jieba.Jieba
	loadTime time.Time // load time
	loadDur  time.Duration
}

// NewJiebaInstance creates a new JiebaInstance or returns an exists JiebaInstance for a given dict directory.
func NewJiebaInstance(dictDir string) *JiebaInstance {
	processDictDir(&dictDir)

	// Big lock here, but ok.
	jiebaInstancesMu.Lock()
	defer jiebaInstancesMu.Unlock()

	inst, ok := jiebaInstances[dictDir]
	if ok {
		return inst
	}

	inst = &JiebaInstance{
		dictDir: dictDir,
	}
	inst.load()

	jiebaInstances[dictDir] = inst

	return inst
}

// FindJiebaInstance returns an exists JiebaInstance for a given dict directory or nil if not found.
func FindJiebaInstance(dictDir string) *JiebaInstance {
	processDictDir(&dictDir)

	jiebaInstancesMu.RLock()
	defer jiebaInstancesMu.RUnlock()
	return jiebaInstances[dictDir]
}

// FindAllJiebaInstances returns all exists JiebaInstances.
func FindAllJiebaInstances() []*JiebaInstance {
	jiebaInstancesMu.RLock()
	defer jiebaInstancesMu.RUnlock()
	ret := make([]*JiebaInstance, 0, len(jiebaInstances))
	for _, v := range jiebaInstances {
		ret = append(ret, v)
	}
	return ret
}

func processDictDir(dictDir *string) {
	// Try env if dictDir is empty.
	if *dictDir == "" {
		*dictDir = os.Getenv(DictDirEnvName)
	}
}

// DictDir returns the dict directory.
func (inst *JiebaInstance) DictDir() string {
	return inst.dictDir
}

// LoadTime returns the load time of data.
func (inst *JiebaInstance) LoadTime() (t time.Time, dur time.Duration) {
	inst.mu.RLock()
	t = inst.loadTime
	dur = inst.loadDur
	inst.mu.RUnlock()
	return
}

// Get returns *jieba.Jieba and a defer function which MUST be called after using.
func (inst *JiebaInstance) Get() (*jieba.Jieba, func()) {
	inst.mu.RLock()
	return inst.val, func() { inst.mu.RUnlock() }
}

// Reload reloads data.
func (inst *JiebaInstance) Reload() {
	newInst := &JiebaInstance{
		dictDir: inst.dictDir,
	}
	newInst.load()

	inst.mu.Lock()
	oldVal := inst.val
	inst.val = newInst.val
	inst.loadTime = newInst.loadTime
	inst.loadDur = newInst.loadDur
	inst.mu.Unlock()

	// XXX: Remeber to free the old data to avoid memory leak
	oldVal.Free()
}

func (inst *JiebaInstance) load() {
	start := time.Now()
	defer func() {
		inst.loadTime = start
		inst.loadDur = time.Since(start)
	}()

	if inst.dictDir == "" {
		inst.val = jieba.NewJieba()
		return
	}

	dictPath := filepath.Join(inst.dictDir, "jieba.dict.utf8")
	hmmPath := filepath.Join(inst.dictDir, "hmm_model.utf8")
	userDictPath := filepath.Join(inst.dictDir, "user.dict.utf8")
	idfPath := filepath.Join(inst.dictDir, "idf.utf8")
	stopWordsPath := filepath.Join(inst.dictDir, "stop_words.utf8")
	inst.val = jieba.NewJieba(dictPath, hmmPath, userDictPath, idfPath, stopWordsPath)
	return
}
