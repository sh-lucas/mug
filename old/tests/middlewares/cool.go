package middlewares

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func CoolMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		emojis := []string{"🚀", "🦄", "🌈", "🔥", "🎉", "💻", "✨"}
		// A slice of cool facts.
		facts := []string{
			"A group of flamingos is called a flamboyance.",
			"Honey never spoils.",
			"The Eiffel Tower can be 15 cm taller during the summer.",
			"A shrimp's heart is in its head.",
			"Go was created at Google in 2007.",
		}

		source := rand.NewSource(time.Now().UnixNano())
		rng := rand.New(source)

		randomEmoji := emojis[rng.Intn(len(emojis))]
		fmt.Printf("%s %s %s\n", randomEmoji, r.Method, r.URL.Path)

		randomFact := facts[rng.Intn(len(facts))]
		w.Header().Set("X-Cool-Fact", randomFact)
		next.ServeHTTP(w, r)
	})
}

func FactLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fact := w.Header().Get("X-Cool-Fact")
		if fact != "" {
			fmt.Printf("💡 Cool Fact Sent: \"%s\"\n", fact)
		}
		next.ServeHTTP(w, r)
	})
}
