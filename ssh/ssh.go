package ssh

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
	"pkg.nimblebun.works/wordle-cli/game"
)

// StartSSH starts the SSH game.
func StartSSH(addr string, conf *game.GameConfig) error {

	s, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithPublicKeyAuth(pubKeyHandler),
		wish.WithMiddleware(
			bm.Middleware(teaHandler(conf)),
			lm.Middleware(),
		),
	)

	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s", addr)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Printf("failed to start ssh server, %s", err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func teaHandler(conf *game.GameConfig) func(ssh.Session) (tea.Model, []tea.ProgramOption) {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil, nil
		}

		pubKey := s.PublicKey()
		if pubKey != nil {
			h := fnv.New64a()
			h.Write(pubKey.Marshal())
			conf.User = h.Sum64()
		}

		fmt.Printf("New SSH connection from %d\n", conf.User)

		model := game.NewGame(conf)

		return model, []tea.ProgramOption{tea.WithAltScreen()}
	}
}

func pubKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}
