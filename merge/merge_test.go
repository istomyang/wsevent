package merge

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkMerge(b *testing.B) {
	merge := NewMerge(context.Background(), time.Millisecond*500)
	merge.Run()
	b.Cleanup(func() {
		merge.Close()
	})

	var total = 0
	var running = 0

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		var k = Key(fmt.Sprint("task-", rand.Intn(20)))
		total++
		if i == 100 {
			merge.UpdateOnline(time.Millisecond * 60)
			b.Log("merge.UpdateOnline(time.Millisecond * 60)")
		}
		if i == 500 {
			merge.Clear()
			b.Log("merge.Clear()")
		}
		if i == 1000 {
			merge.UpdateOnline(time.Millisecond * 200)
			b.Log("merge.UpdateOnline(time.Millisecond * 200)")
		}
		b.StartTimer()
		if merge.Allowed(k) {
			running++
		}
	}

	b.Logf("total: %d", total)
	b.Logf("running: %d", running)
}
