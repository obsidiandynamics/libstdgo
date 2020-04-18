package concurrent

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/obsidiandynamics/libstdgo/arity"
)

type shard struct {
	lock     sync.Mutex
	notify   chan int
	counters map[string]int64
}

func newShard() *shard {
	return &shard{
		counters: make(map[string]int64),
		notify:   make(chan int, 1),
	}
}

func (s *shard) add(key string, amount int64) int64 {
	defer s.notifyUpdate()
	s.lock.Lock()
	defer s.lock.Unlock()
	updated := s.counters[key] + amount
	if updated == 0 {
		delete(s.counters, key)
	} else {
		s.counters[key] = updated
	}
	return updated
}

func (s *shard) set(key string, amount int64) {
	defer s.notifyUpdate()
	s.lock.Lock()
	defer s.lock.Unlock()
	if amount == 0 {
		delete(s.counters, key)
	} else {
		s.counters[key] = amount
	}
}

func (s *shard) notifyUpdate() {
	select {
	case s.notify <- 0:
		Nop()
	default:
		Nop()
	}
}

func (s *shard) get(key string) int64 {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.counters[key]
}

func (s *shard) view(view map[string]int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for k, v := range s.counters {
		view[k] = v
	}
}

func (s *shard) clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.counters = make(map[string]int64)
}

func (s *shard) await(ctx context.Context, key string, cond I64Condition, interval ...time.Duration) int64 {
	checkInterval := optional(DefaultScoreboardCheckInterval, interval...)
	var sleepTicker *time.Ticker
	for {
		value := s.get(key)
		if cond(value) {
			return value
		}

		if sleepTicker == nil {
			sleepTicker = time.NewTicker(checkInterval)
			defer sleepTicker.Stop()
		}

		select {
		case <-ctx.Done():
			return value
		case <-s.notify:
			Nop()
		case <-sleepTicker.C:
			Nop()
		}
	}
}

// Scoreboard is a compactly represented map of atomic counters, where a counter takes up a map slot only if it is
// not equal to zero.
type Scoreboard interface {
	fmt.Stringer
	Add(key string, amount int64) int64
	Inc(key string) int64
	Dec(key string) int64
	Get(key string) int64
	GetInt(key string) int
	Set(key string, value int64)
	Clear()
	View() map[string]int64
	Fill(key string, atLeast int64, timeout time.Duration, interval ...time.Duration) int64
	Drain(key string, atMost int64, timeout time.Duration, interval ...time.Duration) int64
	Await(key string, cond I64Condition, timeout time.Duration, interval ...time.Duration) int64
	AwaitCtx(ctx context.Context, key string, cond I64Condition, interval ...time.Duration) int64
}

type scoreboard struct {
	shards []*shard
}

// DefaultConcurrency is the default level of concurrency applied in the scoreboard constructor.
const DefaultConcurrency = 16

// NewScoreboard creates a new scoreboard instance with an optionally specified concurrency level, controlling
// the number of internal shards. If unspecified, concurrency is set to DefaultConcurrency.
//
// Each shard is individually locked. A greater number of shards allows for a greater degree of uncontended access,
// provided the keys are well-distributed. Shards are created up-front, meaning that scoreboards with more
// concurrency take up more space.
func NewScoreboard(concurrency ...int) Scoreboard {
	conc := arity.SoleUntyped(DefaultConcurrency, concurrency).(int)
	b := &scoreboard{
		shards: make([]*shard, conc),
	}
	for i := 0; i < conc; i++ {
		b.shards[i] = newShard()
	}
	return b
}

// String obtains a string representation of the scoreboard.
func (b scoreboard) String() string {
	return fmt.Sprint("Scoreboard[", b.View(), "]")
}

// Adds a specified amount to the score for the given key, returning the updated value.
func (b *scoreboard) Add(key string, amount int64) int64 {
	return b.forKey(key).add(key, amount)
}

// Increments the score for the given key, returning the updated value.
func (b *scoreboard) Inc(key string) int64 {
	return b.Add(key, 1)
}

// Decrements the score for the given key, returning the updated value.
func (b *scoreboard) Dec(key string) int64 {
	return b.Add(key, -1)
}

// Gets the current score for the given key.
func (b *scoreboard) Get(key string) int64 {
	return b.forKey(key).get(key)
}

// GetInt obtains the current score as a signed int.
func (b *scoreboard) GetInt(key string) int {
	return int(b.Get(key))
}

// Sets a new score value.
func (b *scoreboard) Set(key string, value int64) {
	b.forKey(key).set(key, value)
}

// Clear purges the contents of this scoreboard.
func (b *scoreboard) Clear() {
	for _, shard := range b.shards {
		shard.clear()
	}
}

func (b *scoreboard) View() map[string]int64 {
	view := make(map[string]int64)
	for _, shard := range b.shards {
		shard.view(view)
	}
	return view
}

func (b *scoreboard) forKey(key string) *shard {
	index := hash(key) % uint32(len(b.shards))
	return b.shards[index]
}

func hash(str string) uint32 {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(str))
	return algorithm.Sum32()
}

// DefaultScoreboardCheckInterval is the default check interval used by Await/AwaitCtx/Drain/Fill.
const DefaultScoreboardCheckInterval = 10 * time.Millisecond

// Fill blocks until the score reaches a value that is at least a given minimum.
func (b *scoreboard) Fill(key string, atLeast int64, timeout time.Duration, interval ...time.Duration) int64 {
	return b.Await(key, I64GreaterThanOrEqual(atLeast), timeout, interval...)
}

// Drain blocks until the score drops to a value that does not exceed a given maximum.
func (b *scoreboard) Drain(key string, atMost int64, timeout time.Duration, interval ...time.Duration) int64 {
	return b.Await(key, I64LessThanOrEqual(atMost), timeout, interval...)
}

// Await blocks until a condition is met or expires, returning the last observed score. The optional
// interval argument places an upper bound on the check interval (defaults to DefaultScoreboardCheckInterval).
func (b *scoreboard) Await(key string, cond I64Condition, timeout time.Duration, interval ...time.Duration) int64 {
	ctx, cancel := Timeout(context.Background(), timeout)
	defer cancel()
	return b.AwaitCtx(ctx, key, cond, interval...)
}

// Await blocks until a condition is met or the context is cancelled, returning the last observed score.
// The optional interval argument places an upper bound on the check interval (defaults to DefaultScoreboardCheckInterval).
func (b *scoreboard) AwaitCtx(ctx context.Context, key string, cond I64Condition, interval ...time.Duration) int64 {
	return b.forKey(key).await(ctx, key, cond, interval...)
}
