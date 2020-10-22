package chatroom

import (
	"fmt"
	"strings"
	"sync"
	"tcpspeeddating/pkg/textcolour"
	"time"
)

type User struct {
	Name string
	In   chan string
	Out  chan string
}

var (
	waitingRoom = make(map[string]User)
	seen        = make(map[string]map[string]struct{})
	active      = make(map[string]chan User)
	mu          = new(sync.Mutex)
)

func StartChat() {
	waitingMessage := "waiting for more people..."

	for {
		time.Sleep(5 * time.Second)
		mu.Lock()
		for _, w := range waitingRoom {
			w.Out <- waitingMessage
		}
		mu.Unlock()
	}
}

func pairer() {
	mu.Lock()
	defer mu.Unlock()
	for username, user := range waitingRoom {
		userSeen := seen[user.Name]
		for othername, other := range waitingRoom {
			if username == othername {
				continue
			}
			_, hasSeen := userSeen[other.Name]
			if !hasSeen {
				done := chat(user, other)
				active[user.Name] = done
				active[other.Name] = done
				seen[user.Name][other.Name] = struct{}{}
				seen[other.Name][user.Name] = struct{}{}
				delete(waitingRoom, user.Name)
				delete(waitingRoom, other.Name)
				return
			}
		}
	}
}

func Available(name string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, inWaiting := waitingRoom[name]
	_, inActive := active[name]
	return !inWaiting && !inActive
}

func AddToPool(user User) {
	fmt.Printf("adding user %s\n", user.Name)
	mu.Lock()
	waitingRoom[user.Name] = user
	_, ok := seen[user.Name]
	if !ok {
		seen[user.Name] = make(map[string]struct{})
	}
	user.Out <- textcolour.Green("you're in the pool")
	mu.Unlock()
	pairer()
}

func Remove(user User) {
    fmt.Printf("removing user %s\n", user.Name)
	mu.Lock()
	defer mu.Unlock()

	done, ok := active[user.Name]
	if ok {
		mu.Unlock()
		done <- user
		<-done
		mu.Lock()
	}

	delete(waitingRoom, user.Name)
	delete(seen, user.Name)
	for _, others := range seen {
		delete(others, user.Name)
	}
}

func chat(a, b User) chan User {
	fmt.Printf("making chat with %s and %s\n", a.Name, b.Name)
	done := make(chan User)

	go func() {
	    reQueueA := false
	    reQueueB := false

		defer func() {
			mu.Lock()
			delete(active, a.Name)
			delete(active, b.Name)
			mu.Unlock()
			if reQueueA {
			    AddToPool(a)
            }
            if reQueueB {
                AddToPool(b)
            }
		}()

		aLiked := false
		bLiked := false

		msg := textcolour.Red("you're now in a room")
		a.Out <- msg
		b.Out <- msg

		for {
			select {
			case msg := <-a.In:
				continueLoop := processMessage(msg, a, b, &aLiked, &bLiked)
				if !continueLoop {
					reQueueA = true
					reQueueB = true
					return
				}
			case msg := <-b.In:
				continueLoop := processMessage(msg, b, a, &bLiked, &aLiked)
				if !continueLoop {
                    reQueueA = true
                    reQueueB = true
					return
				}
			case c := <-done:
				fmt.Printf("%s of chat (%s, %s) has disconnected, returning other to pool\n", c.Name, a.Name, b.Name)
				if c.Name == a.Name {
					reQueueB = true
				} else {
					reQueueA = true
				}
				close(done)
				return
			}
		}
	}()

	return done
}

func processMessage(msg string, a, b User, aLiked, bLiked *bool) bool {
	msg = strings.Trim(msg, " ")
	if msg == ".next" {
		b.Out <- textcolour.Red("rejected")
		return false
	} else if msg == ".like" {
		*aLiked = true
		if *bLiked {
			a.Out <- textcolour.Magenta(fmt.Sprintf("%s likes you too!", b.Name))
			b.Out <- textcolour.Magenta(fmt.Sprintf("%s likes you too!", a.Name))
		}
		return true
	}
	b.Out <- textcolour.Blue(msg)
	return true
}
