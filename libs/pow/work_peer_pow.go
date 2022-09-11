package pow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/appditto/pippin_nano_wallet/libs/pow/net"
)

// HTTP post work_generate to every work peer at the same time in parallel
// When the first one returns, accept the result and cancel all other goroutines
// Send work_cancel to all work peers
// Return the work generated by the peer

func WorkGenerateAPIRequest(ctx context.Context, url string, hash string, difficulty string, out chan *string) {
	fmt.Printf("Making work_generate request to %s\n", url)
	resp, err := net.MakeWorkGenerateRequest(ctx, url, hash, difficulty)
	if err == nil && resp.Work != "" {
		WriteChannelSafe(out, resp.Work)
	}
}

func WorkCancelAPIRequest(url string, hash string) {
	net.MakeWorkCancelRequest(context.Background(), url, hash)
}

// Invoke work_generate to all peers simultaneously
func (p *PippinPow) WorkGeneratePeers(hash string, difficulty string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultChan := make(chan *string, len(p.WorkPeers))
	defer close(resultChan)

	for _, peer := range p.WorkPeers {
		go WorkGenerateAPIRequest(ctx, peer, hash, difficulty, resultChan)
	}

	select {
	case result := <-resultChan:
		// Send work cancel
		for _, peer := range p.WorkPeers {
			go WorkCancelAPIRequest(peer, hash)
		}
		return *result, nil
	// 30
	case <-time.After(10 * time.Second):
		// Send work cancel
		for _, peer := range p.WorkPeers {
			go WorkCancelAPIRequest(peer, hash)
		}
		return "", errors.New("Unable to generate work")
	}
}

// Recovers from writing to close channel
func WriteChannelSafe(out chan *string, msg string) (err error) {
	defer func() {
		// recover from panic caused by writing to a closed channel
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			return
		}
	}()

	out <- &msg // write on possibly closed channel

	return err
}
