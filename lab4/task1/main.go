// Task 1: Notifications + participation task (grades table into an io.Writer).
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

type Notifier interface {
	Notify(msg string) error
}

type Email struct{ Address string }

func (e Email) Notify(msg string) error {
	fmt.Printf("email    -> %s: %s\n", e.Address, msg)
	return nil
}

const smsMaxLen = 20

var ErrTooLong = errors.New("message too long")

type SMS struct{ Phone string }

// Notify counts characters (runes), not bytes: "привет" is 6 characters
// but 12 bytes.
func (s SMS) Notify(msg string) error {
	if n := utf8.RuneCountInString(msg); n > smsMaxLen {
		return fmt.Errorf("sms to %s: %d characters, max %d: %w", s.Phone, n, smsMaxLen, ErrTooLong)
	}
	fmt.Printf("sms      -> %s: %s\n", s.Phone, msg)
	return nil
}

// Telegram changes its own state in Notify, so it needs a pointer receiver.
// With a value receiver Sent would be incremented on a copy and stay 0.
// As a result only *Telegram is a Notifier, not Telegram.
type Telegram struct {
	Chat string
	Sent int
}

func (t *Telegram) Notify(msg string) error {
	t.Sent++
	fmt.Printf("telegram -> %s: %s\n", t.Chat, msg)
	return nil
}

// Compile-time checks.
var (
	_ Notifier = Email{}
	_ Notifier = SMS{}
	_ Notifier = (*Telegram)(nil)
)

func Broadcast(ns []Notifier, msg string) []error {
	var errs []error
	for _, n := range ns {
		if err := n.Notify(msg); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Participation task.

type Grade struct {
	Name  string
	Score int
}

func printGrades(w io.Writer, grades []Grade) {
	fmt.Fprintf(w, "%-10s %5s\n", "Name", "Score")
	fmt.Fprintln(w, strings.Repeat("-", 16))
	for _, g := range grades {
		fmt.Fprintf(w, "%-10s %5d\n", g.Name, g.Score)
	}
}

func main() {
	tg := &Telegram{Chat: "@se-2401"}
	ns := []Notifier{
		Email{Address: "aruzhan@uni.kz"},
		SMS{Phone: "+77001234567"},
		tg,
	}

	for _, msg := range []string{"Lab 4 is out", "Defence is moved to Friday, room 305"} {
		for _, err := range Broadcast(ns, msg) {
			fmt.Println("  error:", err)
			fmt.Println("  is ErrTooLong:", errors.Is(err, ErrTooLong))
		}
	}
	fmt.Println("telegram sent:", tg.Sent)

	grades := []Grade{
		{"Aruzhan", 95},
		{"Dias", 80},
		{"Timur", 72},
		{"Bota", 88},
	}

	fmt.Println()
	printGrades(os.Stdout, grades)

	var sb strings.Builder
	printGrades(&sb, grades)
	fmt.Printf("\nstrings.Builder holds %d bytes:\n%s", sb.Len(), sb.String())
}
