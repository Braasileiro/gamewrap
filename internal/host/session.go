package host

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"gamewrap/internal/config"
)

type Session struct {
	log         *log.Logger
	settings    config.Settings
	commands    commandRunner
	lock        *os.File
	vrrOutput   string
	vrrOriginal config.VRRMode
	vrrChanged  bool
}

func NewSession(logger *log.Logger, settings config.Settings) *Session {
	return &Session{log: logger, settings: settings, commands: systemCommandRunner{}}
}

// Prepare acquires the host lock before changing shared VRR state.
func (s *Session) Prepare() {
	if s.settings.VRR == "" {
		return
	}

	lock, acquired, err := acquireSessionLock()
	if err != nil {
		s.log.Printf("host settings disabled: failed to acquire session lock: %v", err)
		return
	}

	if !acquired {
		s.log.Print("host settings disabled: another managed game is running")
		return
	}

	s.lock = lock

	s.enableVRR()
	if !s.vrrChanged {
		s.releaseLock()
		s.lock = nil
	}
}

// Restore reverts only the host settings changed by this session.
func (s *Session) Restore() {
	if s.lock == nil {
		return
	}

	defer s.releaseLock()

	if s.vrrChanged {
		if err := setVRR(s.commandRunner(), s.vrrOutput, s.vrrOriginal); err != nil {
			s.log.Printf("VRR: failed to restore %s: %v", s.vrrOriginal, err)
		} else {
			s.log.Printf("VRR: restored %s on %s", s.vrrOriginal, s.vrrOutput)
		}
	}
}

func (s *Session) commandRunner() commandRunner {
	if s.commands == nil {
		return systemCommandRunner{}
	}

	return s.commands
}

func acquireSessionLock() (*os.File, bool, error) {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.TempDir()
	}

	path := filepath.Join(dir, fmt.Sprintf("gamewrap-%d.lock", os.Getuid()))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, false, err
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, false, nil
		}

		return nil, false, err
	}

	return file, true, nil
}

func (s *Session) releaseLock() {
	if err := syscall.Flock(int(s.lock.Fd()), syscall.LOCK_UN); err != nil {
		s.log.Printf("failed to release session lock: %v", err)
	}

	if err := s.lock.Close(); err != nil {
		s.log.Printf("failed to close session lock: %v", err)
	}
}
