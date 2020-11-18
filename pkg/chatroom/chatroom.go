package chatroom

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"tcpspeeddating/pkg/textcolour"
	"time"
)

type username string
type user struct {
	name username
	in   chan string
	out  chan string
}

var (
	waitingRoom = make(map[username]user)
	seen        = make(map[username]map[username]struct{})
	activeRooms = make(map[username]chan username)
	mu          = new(sync.Mutex)
)

func StartChat() {
	waitingMessage := "waiting for more people..."

	for {
		time.Sleep(5 * time.Second)
		mu.Lock()
		for _, waiting := range waitingRoom {
			waiting.out <- waitingMessage
		loop:
			for {
				select {
				case <-waiting.in:
				default:
					break loop
				}
			}
		}
		mu.Unlock()
	}
}

func Available(name string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, inWaiting := waitingRoom[username(name)]
	_, inActive := activeRooms[username(name)]
	return !inWaiting && !inActive
}

func Add(name string, in, out chan string) func() {
	username := username(name)
	user := user{username, in, out}
	addToWaitingRoom(user)

	return func() {
		log.Printf("removing user %q\n", name)
		mu.Lock()

		roomKiller, ok := activeRooms[username]
		if ok {
			mu.Unlock()
			roomKiller <- username
			<-roomKiller
			mu.Lock()
		}

		delete(waitingRoom, username)
		delete(seen, username)
		for _, others := range seen {
			delete(others, username)
		}
		mu.Unlock()
	}
}

func addToWaitingRoom(user user) {
	mu.Lock()
	defer func() {
		mu.Unlock()
		pairer()
	}()

	waitingRoom[user.name] = user

	_, ok := seen[user.name]
	if !ok {
		log.Printf("adding new user %q to waiting room\n", user.name)
		seen[user.name] = make(map[username]struct{})
	} else {
		log.Printf("returning user %q to waiting room\n", user.name)
	}

	user.out <- textcolour.Green("you're in the waiting room")
}

func pairer() {
	mu.Lock()
	defer mu.Unlock()

	for username, user := range waitingRoom {
		userSeen := seen[user.name]
		for othername, other := range waitingRoom {
			if username == othername {
				continue
			}

			_, hasSeen := userSeen[other.name]
			if !hasSeen {
				done := chat(user, other)
				activeRooms[user.name] = done
				activeRooms[other.name] = done
				seen[user.name][other.name] = struct{}{}
				seen[other.name][user.name] = struct{}{}
				delete(waitingRoom, user.name)
				delete(waitingRoom, other.name)
				return
			}
		}
	}
}

func chat(a, b user) chan username {
	log.Printf("making chat with %q and %q\n", a.name, b.name)
	done := make(chan username)

	go func() {
		reQueueA := false
		reQueueB := false

		aLiked := false
		bLiked := false

		defer func() {
			mu.Lock()
			delete(activeRooms, a.name)
			delete(activeRooms, b.name)
			mu.Unlock()
			if reQueueA {
				addToWaitingRoom(a)
			}
			if reQueueB {
				addToWaitingRoom(b)
			}
		}()

		msg := textcolour.Green("type .like if you like someone and .next to go back to waiting room for your next match")
		a.out <- msg
		b.out <- msg

		msg = textcolour.Red("you're now in a room")
		a.out <- msg
		b.out <- msg

		for {
			select {
			case msg := <-a.in:
				continueLoop := processMessage(msg, a, b, &aLiked, &bLiked)
				if !continueLoop {
					reQueueA = true
					reQueueB = true
					return
				}
			case msg := <-b.in:
				continueLoop := processMessage(msg, b, a, &bLiked, &aLiked)
				if !continueLoop {
					reQueueA = true
					reQueueB = true
					return
				}
			case exitingUser := <-done:
				log.Printf("%q of chat (%q, %q) has disconnected, returning other to waiting room\n", exitingUser, a.name, b.name)
				if exitingUser == a.name {
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

func processMessage(msg string, a, b user, aLiked, bLiked *bool) bool {
	msg = strings.Trim(msg, " ")
	switch msg {
	case ".next":
		if *bLiked && !*aLiked {
			b.out <- textcolour.Red("rejected")
		}
		return false
	case ".like":
		if *aLiked {
			if *bLiked {
				a.out <- textcolour.Magenta("you and %s already like each other!", b.name)
			} else {
				a.out <- textcolour.Magenta("you have already liked the other person!")
			}
		} else {
			*aLiked = true
			if *bLiked {
				a.out <- textcolour.Magenta("%s likes you too!", b.name)
				b.out <- textcolour.Magenta("%s likes you too!", a.name)
			} else {
				a.out <- textcolour.Magenta("you like the other person")
			}
		}
	case ".heart":
		msg = "❤️"
		fallthrough
	default:
		if *aLiked && *bLiked {
			b.out <- fmt.Sprintf("%s: %s", a.name, textcolour.Blue(msg))
		} else {
			b.out <- "anon: " + textcolour.Blue(msg)
		}
	}
	return true
}
