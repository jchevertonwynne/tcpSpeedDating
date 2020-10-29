package chatroom

import (
	"fmt"
	"strings"
	"sync"
	"tcpspeeddating/pkg/textcolour"
	"time"
)

type Username string

type User struct {
	Name Username
	In   chan string
	Out  chan string
}

var (
	waitingRoom = make(map[Username]User)
	seen        = make(map[Username]map[Username]struct{})
	activeRooms = make(map[Username]chan Username)
	mu          = new(sync.Mutex)
)

func StartChat() {
	waitingMessage := "waiting for more people..."

	for {
		time.Sleep(5 * time.Second)
		mu.Lock()
		for _, waiting := range waitingRoom {
			waiting.Out <- waitingMessage
		loop:
			for {
				select {
				case <-waiting.In:
				default:
					break loop
				}
			}
		}
		mu.Unlock()
	}
}

func Available(name Username) bool {
	mu.Lock()
	defer mu.Unlock()
	_, inWaiting := waitingRoom[name]
	_, inActive := activeRooms[name]
	return !inWaiting && !inActive
}

func AddToWaitingRoom(user User) {
	mu.Lock()
	defer func() {
		mu.Unlock()
		pairer()
	}()
	waitingRoom[user.Name] = user
	_, ok := seen[user.Name]
	if !ok {
		fmt.Printf("adding new user %s to waiting room\n", user.Name)
		seen[user.Name] = make(map[Username]struct{})
	} else {
		fmt.Printf("returning user %s to waiting room\n", user.Name)
	}
	user.Out <- textcolour.Green("you're in the waiting room")
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
				activeRooms[user.Name] = done
				activeRooms[other.Name] = done
				seen[user.Name][other.Name] = struct{}{}
				seen[other.Name][user.Name] = struct{}{}
				delete(waitingRoom, user.Name)
				delete(waitingRoom, other.Name)
				return
			}
		}
	}
}

func Remove(user User) {
	fmt.Printf("removing user %s\n", user.Name)
	mu.Lock()
	defer mu.Unlock()

	roomKiller, ok := activeRooms[user.Name]
	if ok {
		mu.Unlock()
		roomKiller <- user.Name
		<-roomKiller
		mu.Lock()
	}

	delete(waitingRoom, user.Name)
	delete(seen, user.Name)
	for _, others := range seen {
		delete(others, user.Name)
	}
}

func chat(a, b User) chan Username {
	fmt.Printf("making chat with %s and %s\n", a.Name, b.Name)
	done := make(chan Username)

	msg := textcolour.Green("type .like if you like someone and .next to go back to waiting room for your next match")
	a.Out <- msg
	b.Out <- msg

	go func() {
		reQueueA := false
		reQueueB := false

		defer func() {
			mu.Lock()
			delete(activeRooms, a.Name)
			delete(activeRooms, b.Name)
			mu.Unlock()
			if reQueueA {
				AddToWaitingRoom(a)
			}
			if reQueueB {
				AddToWaitingRoom(b)
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
				fmt.Printf("%s of chat (%s, %s) has disconnected, returning other to waiting room\n", c, a.Name, b.Name)
				if c == a.Name {
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
	switch msg {
	case ".next":
		if *bLiked && !*aLiked {
			b.Out <- textcolour.Red("rejected")
		}
		return false
	case ".like":
		if *aLiked {
			if *bLiked {
				a.Out <- textcolour.Magenta(fmt.Sprintf("you and %s already like each other!", b.Name))
			} else {
				a.Out <- textcolour.Magenta("you have already liked the other person!")
			}
		} else {
			*aLiked = true
			if *bLiked {
				a.Out <- textcolour.Magenta(fmt.Sprintf("%s likes you too!", b.Name))
				b.Out <- textcolour.Magenta(fmt.Sprintf("%s likes you too!", a.Name))
			} else {
				a.Out <- textcolour.Magenta("you like the other person")
			}
		}
	case ".heart":
		msg = "❤️"
		fallthrough
	default:
		if *aLiked && *bLiked {
			b.Out <- string(a.Name) + ": " + textcolour.Blue(msg)
		} else {
			b.Out <- "anon: " + textcolour.Blue(msg)
		}
	}
	return true
}
