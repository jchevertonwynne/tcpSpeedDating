package chatroom

import (
	"sync"
	"time"
)

type Chans struct {
	In  chan []byte
	Out chan []byte
}

var (
	waitingRoom = make(map[Chans]struct{})
	seen        = make(map[Chans]map[Chans]struct{})
	active      = make(map[Chans]chan Chans)
	mu          = new(sync.Mutex)
)

func StartChat() {
	waitingMessage := []byte("waiting for more people...")

	for {
		time.Sleep(5 * time.Second)
		mu.Lock()
		for w := range waitingRoom {
		loop:
			for {
				select {
				case msg := <-w.In:
					for wOther := range waitingRoom {
						if w != wOther {
							wOther.Out <- msg
						}
					}
				default:
					break loop
				}
			}

			w.Out <- waitingMessage
		}
		mu.Unlock()
	}
}

func pairer() {
	mu.Lock()
	defer mu.Unlock()
	for user := range waitingRoom {
		userSeen := seen[user]
		for other := range waitingRoom {
			if other == user {
				continue
			}
			_, hasSeen := userSeen[other]
			if !hasSeen {
				done := chat(user, other)
				active[user] = done
				active[other] = done
				seen[user][other] = struct{}{}
				seen[other][user] = struct{}{}
				delete(waitingRoom, user)
				delete(waitingRoom, other)
				return
			}
		}
	}
}

func AddToPool(chans Chans) {
	mu.Lock()

	waitingRoom[chans] = struct{}{}
	_, s := seen[chans]
	if !s {
		seen[chans] = make(map[Chans]struct{})
	}
	chans.Out <- []byte("\u001b[31myou're in the pool\u001b[0m")

	mu.Unlock()

	pairer()
}

func Remove(chans Chans) {
	mu.Lock()
	defer mu.Unlock()

	done, ok := active[chans]
	if ok {
		mu.Unlock()
		done <- chans
		<-done
		mu.Lock()
	}

	delete(waitingRoom, chans)
	delete(seen, chans)
	delete(active, chans)
	for _, others := range seen {
		delete(others, chans)
	}
}

func chat(a, b Chans) chan Chans {
	done := make(chan Chans)

	go func() {
		defer func() {
			mu.Lock()
			defer mu.Unlock()
			delete(active, a)
			delete(active, b)
		}()

		msg := red("you're now in a room")
		a.Out <- msg
		b.Out <- msg
	loop:
		for {
			select {
			case m := <-a.In:
				if string(m) == ".next" {
					b.Out <- red("rejected")
					AddToPool(a)
					AddToPool(b)
					break loop
				}
				b.Out <- magenta(string(m))
			case m := <-b.In:
				if string(m) == ".next" {
					a.Out <- red("rejected")
					AddToPool(a)
					AddToPool(b)
					break loop
				}
				a.Out <- magenta(string(m))
			case c := <-done:
				if c == a {
					AddToPool(b)
				} else {
					AddToPool(a)
				}
				close(done)
				return
			}
		}
	}()

	return done
}

func red(msg string) []byte {
	return formatColour(msg, []byte("\u001b[31m"))
}

func magenta(msg string) []byte {
	return formatColour(msg, []byte("\u001b[35m"))
}

func formatColour(msg string, colour []byte) []byte {
	return append(append(colour, msg...), []byte("\u001b[0m")...)
}
