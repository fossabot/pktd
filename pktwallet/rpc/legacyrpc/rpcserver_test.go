// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package legacyrpc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestThrottle(t *testing.T) {
	const threshold = 1
	busy := make(chan struct{})

	srv := httptest.NewServer(throttledFn(threshold,
		func(w http.ResponseWriter, r *http.Request) {
			<-busy
		}),
	)

	failed := int32(0)
	codes := make(chan int, 2)
	for i := 0; i < cap(codes); i++ {
		go func() {
			res, err := http.Get(srv.URL)
			if err != nil {
				// t.Fatal(err)
				fmt.Println(err)
				atomic.StoreInt32(&failed, int32(1))
			}
			codes <- res.StatusCode
		}()
	}

	got := make(map[int]int, cap(codes))
	for i := 0; i < cap(codes); i++ {
		got[<-codes]++

		if i == 0 {
			close(busy)
		}
	}

	want := map[int]int{200: 1, 429: 1}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("status codes: want: %v, got: %v", want, got)
	}
	if failed != 0 {
		t.Fail()
	}
}
