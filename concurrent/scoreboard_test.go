package concurrent

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const defKey = "key"

func TestNewScoreboardSetAndGet(t *testing.T) {
	b := NewScoreboard()
	assert.Equal(t, b.View(), map[string]int64{})

	b.Set(defKey, 42)
	assert.Equal(t, b.View(), map[string]int64{defKey: 42})
	assert.Equal(t, 42, b.GetInt(defKey))

	b.Set(defKey, 0)
	assert.Equal(t, b.View(), map[string]int64{})
}

func TestScoreboardDrainInDeepSleep(t *testing.T) {
	b := NewScoreboard(1)
	b.Set(defKey, 1)
	go func() {
		time.Sleep(1 * time.Millisecond)
		res := b.Add(defKey, -1)
		assert.Equal(t, int64(0), res)
	}()

	res := b.Drain(defKey, 0, Indefinitely, 1*time.Hour)
	assert.Equal(t, int64(0), res)
}

func TestScoreboardAwaitCtxCancel(t *testing.T) {
	b := NewScoreboard()
	b.Set(defKey, 1)
	ctx, cancel := Forever(context.Background())
	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	defer cancel()
	res := b.AwaitCtx(ctx, defKey, I64Equal(0), 1*time.Hour)
	assert.Equal(t, int64(1), res)
}

func TestScoreboardAwaitCtxInDeepSleep(t *testing.T) {
	b := NewScoreboard(1)
	b.Set(defKey, 1)
	go func() {
		time.Sleep(1 * time.Millisecond)
		res := b.Dec(defKey)
		assert.Equal(t, int64(0), res)
	}()

	ctx, cancel := Forever(context.Background())
	defer cancel()
	res := b.AwaitCtx(ctx, defKey, I64Equal(0), 1*time.Hour)
	assert.Equal(t, int64(0), res)
}

func TestScoreboardDrainWithTimeout(t *testing.T) {
	b := NewScoreboard(4)
	b.Set(defKey, 1)
	res := b.Drain(defKey, 0, 1*time.Microsecond)
	assert.Equal(t, int64(1), res)
}

func TestScoreboardFillWithTimeout(t *testing.T) {
	b := NewScoreboard()
	res := b.Fill(defKey, 1, 1*time.Microsecond)
	assert.Equal(t, int64(0), res)
}

func TestScoreboardIncrement(t *testing.T) {
	b := NewScoreboard()
	res := b.Inc(defKey)
	assert.Equal(t, int64(1), res)
	assert.Equal(t, 1, b.GetInt(defKey))
}

func TestScoreboardThreadedIncrement(t *testing.T) {
	b := NewScoreboard()

	const routines = 10
	const perRoutine = 100

	wg := sync.WaitGroup{}
	wg.Add(routines)
	for r := 0; r < routines; r++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perRoutine; j++ {
				b.Inc(defKey)
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, routines*perRoutine, b.GetInt(defKey))
}

func TestScoreboardThreadedIncrementAndDecrement(t *testing.T) {
	b := NewScoreboard()

	const routines = 10
	const perRoutine = 100

	wg := sync.WaitGroup{}
	wg.Add(routines)
	for r := 0; r < routines; r++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perRoutine; j++ {
				b.Inc(defKey)
				runtime.Gosched()
				b.Dec(defKey)
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, 0, b.GetInt(defKey))
	assert.Empty(t, b.View())
}

func TestScoreboardSet(t *testing.T) {
	b := NewScoreboard(3)
	b.Set(defKey, 7)
	assert.Equal(t, 7, b.GetInt(defKey))
}

func TestScoreboardClear(t *testing.T) {
	b := NewScoreboard(3)
	b.Set(defKey, 7)
	b.Clear()
	assert.Empty(t, b.View())
}

func TestScoreboardStringer(t *testing.T) {
	b := NewScoreboard()
	b.Set(defKey, 1)
	assert.Equal(t, "Scoreboard[map[key:1]]", b.String())
}
