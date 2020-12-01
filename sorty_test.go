/*	Copyright (c) 2019, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package sorty

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
	"unsafe"

	"github.com/jfcg/opt"
	"github.com/jfcg/sixb"
	"github.com/shawnsmithdev/zermelo/zfloat32"
	"github.com/shawnsmithdev/zermelo/zuint32"
	"github.com/twotwotwo/sorts/sortutil"
	"github.com/yourbasic/radix"
)

const N = 1 << 26

var tst *testing.T
var name string

// fill sort test for uint32
func fstU4(sd int64, ar []uint32, srt func([]uint32)) time.Duration {
	rn := rand.New(rand.NewSource(sd))
	for i := len(ar) - 1; i >= 0; i-- {
		ar[i] = rn.Uint32()
	}

	now := time.Now()
	srt(ar)
	dur := time.Since(now)

	if 0 != IsSortedU4(ar) {
		tst.Fatal(name, "not sorted")
	}
	return dur
}

// fill sort test for uint64
func fstU8(sd int64, ar []uint64, srt func([]uint64)) time.Duration {
	rn := rand.New(rand.NewSource(sd))
	for i := len(ar) - 1; i >= 0; i-- {
		ar[i] = rn.Uint64()
	}

	now := time.Now()
	srt(ar)
	dur := time.Since(now)

	if 0 != IsSortedU8(ar) {
		tst.Fatal(name, "not sorted")
	}
	return dur
}

// fill sort test for int32
func fstI4(sd int64, ar []int32, srt func([]int32)) time.Duration {
	rn := rand.New(rand.NewSource(sd))
	for i := len(ar) - 1; i >= 0; i-- {
		ar[i] = int32(rn.Uint32())
	}

	now := time.Now()
	srt(ar)
	dur := time.Since(now)

	if 0 != IsSortedI4(ar) {
		tst.Fatal(name, "not sorted")
	}
	return dur
}

// fill sort test for int64
func fstI8(sd int64, ar []int64, srt func([]int64)) time.Duration {
	rn := rand.New(rand.NewSource(sd))
	for i := len(ar) - 1; i >= 0; i-- {
		ar[i] = int64(rn.Uint64())
	}

	now := time.Now()
	srt(ar)
	dur := time.Since(now)

	if 0 != IsSortedI8(ar) {
		tst.Fatal(name, "not sorted")
	}
	return dur
}

// fill sort test for float32
func fstF4(sd int64, ar []float32, srt func([]float32)) time.Duration {
	rn := rand.New(rand.NewSource(sd))
	for i := len(ar) - 1; i >= 0; i-- {
		ar[i] = float32(rn.NormFloat64())
	}

	now := time.Now()
	srt(ar)
	dur := time.Since(now)

	if 0 != IsSortedF4(ar) {
		tst.Fatal(name, "not sorted")
	}
	return dur
}

// fill sort test for float64
func fstF8(sd int64, ar []float64, srt func([]float64)) time.Duration {
	rn := rand.New(rand.NewSource(sd))
	for i := len(ar) - 1; i >= 0; i-- {
		ar[i] = rn.NormFloat64()
	}

	now := time.Now()
	srt(ar)
	dur := time.Since(now)

	if 0 != IsSortedF8(ar) {
		tst.Fatal(name, "not sorted")
	}
	return dur
}

// implant strings into ar
func implant(ar []uint32, fill bool) ([]string, []uint32) {
	// string size is 4*t bytes
	t := int(unsafe.Sizeof("") >> 2)

	// ar will hold n strings (headers followed by 4-byte bodies)
	n := len(ar) / (t + 1)

	t *= n // total string headers space
	ss := sixb.I4tSs(ar[:t:t])

	if fill {
		for i, k := n-1, len(ar)-1; i >= 0; i, k = i-1, k-1 {
			ss[i].Data = unsafe.Pointer(&ar[k])
			ss[i].Len = 4
		}
	}
	return *(*[]string)(unsafe.Pointer(&ss)), ar[t:]
}

// fill sort test for string
func fstS(sd int64, ar []uint32, srt func([]string)) time.Duration {
	as, ar := implant(ar, true)

	rn := rand.New(rand.NewSource(sd))
	for i := len(ar) - 1; i >= 0; i-- {
		ar[i] = rn.Uint32()
	}

	now := time.Now()
	srt(as)
	dur := time.Since(now)

	if 0 != IsSortedS(as) {
		tst.Fatal(name, "not sorted")
	}
	return dur
}

func compareU4(ar, ap []uint32) {
	l := len(ap)
	if l <= 0 {
		return
	}
	if len(ar) != l {
		tst.Fatal(name, "length mismatch:", len(ar), l)
	}

	for i := l - 1; i >= 0; i-- {
		if ar[i] != ap[i] {
			tst.Fatal(name, "values mismatch:", i, ar[i], ap[i])
		}
	}
}

func compareS(ar, ap []string) {
	l := len(ap)
	if len(ar) != l {
		tst.Fatal(name, "length mismatch:", len(ar), l)
	}

	for i := l - 1; i >= 0; i-- {
		if ar[i] != ap[i] {
			tst.Fatal(name, "values mismatch:", i, ar[i], ap[i])
		}
	}
}

// median of four durations
func medur(a, b, c, d time.Duration) time.Duration {
	if d < b {
		d, b = b, d
	}
	if c < a {
		c, a = a, c
	}
	if d < c {
		c = d
	}
	if b < a {
		b = a
	}
	return (b + c) >> 1
}

func printSec(d time.Duration) float64 {
	sec := d.Seconds()
	if testing.Short() {
		fmt.Printf("%10s %5.2fs\n", name, sec)
	}
	return sec
}

// median fst & compare for uint32
func mfcU4(tn string, srt func([]uint32), ar, ap []uint32) float64 {
	name = tn
	d1 := fstU4(1, ar, srt) // median of four different sorts
	d2 := fstU4(2, ar, srt)
	d3 := fstU4(3, ar, srt)
	d1 = medur(fstU4(4, ar, srt), d1, d2, d3)

	compareU4(ar, ap)
	return printSec(d1)
}

// slice conversions
func F4toU4(p *[]float32) []uint32 {
	return *(*[]uint32)(unsafe.Pointer(p))
}

func F4toI4(p *[]float32) []int32 {
	return *(*[]int32)(unsafe.Pointer(p))
}

func U8toF8(p *[]uint64) []float64 {
	return *(*[]float64)(unsafe.Pointer(p))
}

func U8toI8(p *[]uint64) []int64 {
	return *(*[]int64)(unsafe.Pointer(p))
}

// median fst & compare for float32
func mfcF4(tn string, srt func([]float32), ar, ap []float32) float64 {
	name = tn
	d1 := fstF4(5, ar, srt) // median of four different sorts
	d2 := fstF4(6, ar, srt)
	d3 := fstF4(7, ar, srt)
	d1 = medur(fstF4(8, ar, srt), d1, d2, d3)

	compareU4(F4toU4(&ar), F4toU4(&ap))
	return printSec(d1)
}

// median fst & compare for string
func mfcS(tn string, srt func([]string), ar, ap []uint32) float64 {
	name = tn
	d1 := fstS(9, ar, srt) // median of four different sorts
	d2 := fstS(10, ar, srt)
	d3 := fstS(11, ar, srt)
	d1 = medur(fstS(12, ar, srt), d1, d2, d3)

	if len(ap) > 0 {
		as, ar := implant(ar, false)
		aq, ap := implant(ap, false)
		compareS(as, aq)
		compareU4(ar, ap)
	}
	return printSec(d1)
}

var srnm = []byte("sorty-0")

// return sum of SortU4() times for 1..4 goroutines
// compare with ap and among themselves
func sumtU4(ar, ap []uint32) float64 {
	s := .0
	for Mxg = 1; Mxg < 5; Mxg++ {
		srnm[6] = byte(Mxg + '0')
		s += mfcU4(string(srnm), SortU4, ar, ap)
		ap, ar = ar, ap[:cap(ap)]
	}
	return s
}

// return sum of SortF4() times for 1..4 goroutines
// compare with ap and among themselves
func sumtF4(ar, ap []float32) float64 {
	s := .0
	for Mxg = 1; Mxg < 5; Mxg++ {
		srnm[6] = byte(Mxg + '0')
		s += mfcF4(string(srnm), SortF4, ar, ap)
		ap, ar = ar, ap[:cap(ap)]
	}
	return s
}

// return sum of SortS() times for 1..4 goroutines
// compare with ap and among themselves
func sumtS(ar, ap []uint32) float64 {
	s := .0
	for Mxg = 1; Mxg < 5; Mxg++ {
		srnm[6] = byte(Mxg + '0')
		s += mfcS(string(srnm), SortS, ar, ap)
		ap, ar = ar, ap[:cap(ap)]
	}
	return s
}

// sort uint32 array with Sort()
func sort3i(aq []uint32) {
	lsw := func(i, k, r, s int) bool {
		if aq[i] < aq[k] {
			if r != s {
				aq[r], aq[s] = aq[s], aq[r]
			}
			return true
		}
		return false
	}
	Sort(len(aq), lsw)
}

var lswnm = []byte("sortyLsw-0")

// return sum of sort3i() times for 1..4 goroutines
// compare with ap and among themselves
func sumtLswU4(ar, ap []uint32) float64 {
	s := .0
	for Mxg = 1; Mxg < 5; Mxg++ {
		lswnm[9] = byte(Mxg + '0')
		s += mfcU4(string(lswnm), sort3i, ar, ap)
		ap, ar = ar, ap[:cap(ap)]
	}
	return s
}

// sort float32 array with Sort()
func sort3f(aq []float32) {
	lsw := func(i, k, r, s int) bool {
		if aq[i] < aq[k] {
			if r != s {
				aq[r], aq[s] = aq[s], aq[r]
			}
			return true
		}
		return false
	}
	Sort(len(aq), lsw)
}

// return sum of sort3f() times for 1..4 goroutines
// compare with ap and among themselves
func sumtLswF4(ar, ap []float32) float64 {
	s := .0
	for Mxg = 1; Mxg < 5; Mxg++ {
		lswnm[9] = byte(Mxg + '0')
		s += mfcF4(string(lswnm), sort3f, ar, ap)
		ap, ar = ar, ap[:cap(ap)]
	}
	return s
}

// sort string array with Sort()
func sort3s(aq []string) {
	lsw := func(i, k, r, s int) bool {
		if aq[i] < aq[k] {
			if r != s {
				aq[r], aq[s] = aq[s], aq[r]
			}
			return true
		}
		return false
	}
	Sort(len(aq), lsw)
}

// return sum of sort3s() times for 1..4 goroutines
// compare with ap and among themselves
func sumtLswS(ar, ap []uint32) float64 {
	s := .0
	for Mxg = 1; Mxg < 5; Mxg++ {
		lswnm[9] = byte(Mxg + '0')
		s += mfcS(string(lswnm), sort3s, ar, ap)
		ap, ar = ar, ap[:cap(ap)]
	}
	return s
}

// sort and signal
func sasU8(sd int64, al []uint64, ch chan bool) {
	fstU8(sd, al, SortU8)
	ch <- false
}

func sasF8(sd int64, al []float64, ch chan bool) {
	fstF8(sd, al, SortF8)
	ch <- false
}

func sasI4(sd int64, al []int32, ch chan bool) {
	fstI4(sd, al, SortI4)
	ch <- false
}

func sasI8(sd int64, al []int64, ch chan bool) {
	fstI8(sd, al, SortI8)
	ch <- false
}

// main test routine, needs -short flag
func TestShort(t *testing.T) {
	if !testing.Short() {
		t.SkipNow()
	}
	tst = t

	// a & b buffers will hold all arrays to sort
	af := make([]float32, N)
	bf := make([]float32, N)

	// different type views of the same buffers
	au, bu := F4toU4(&af), F4toU4(&bf)         // uint32
	ai, _ := F4toI4(&af), F4toI4(&bf)          // int32
	au2, bu2 := sixb.I4tI8(au), sixb.I4tI8(bu) // uint64
	af2, _ := U8toF8(&au2), U8toF8(&bu2)       // float64
	_, bi2 := U8toI8(&au2), U8toI8(&bu2)       // int64

	// test & time sorting uint32 arrays for different libraries
	// compare their results
	fmt.Println("Sorting uint32")
	mfcU4("sort.Slice", func(al []uint32) {
		sort.Slice(al, func(i, k int) bool { return al[i] < al[k] })
	}, bu, nil)
	mfcU4("sortutil", sortutil.Uint32s, au, bu)
	mfcU4("zermelo", zuint32.Sort, au, bu)
	sumtU4(au, bu) // sorty
	sumtLswU4(au, bu)

	if 0 != IsSorted(len(au), func(i, k, r, s int) bool { return au[i] < au[k] }) {
		t.Fatal("IsSorted() does not work")
	}

	// test & time sorting float32 arrays for different libraries
	// compare their results
	fmt.Println("\nSorting float32")
	mfcF4("sort.Slice", func(al []float32) {
		sort.Slice(al, func(i, k int) bool { return al[i] < al[k] })
	}, bf, nil)
	mfcF4("sortutil", sortutil.Float32s, af, bf)
	mfcF4("zermelo", zfloat32.Sort, af, bf)
	sumtF4(af, bf) // sorty
	sumtLswF4(af, bf)

	if 0 != IsSorted(len(af), func(i, k, r, s int) bool { return af[i] < af[k] }) {
		t.Fatal("IsSorted() does not work")
	}

	// test & time sorting string arrays for different libraries
	// compare their results
	fmt.Println("\nSorting string")
	mfcS("sort.Slice", func(al []string) {
		sort.Slice(al, func(i, k int) bool { return al[i] < al[k] })
	}, bu, nil)
	mfcS("sortutil", sortutil.Strings, au, bu)
	mfcS("radix", radix.Sort, au, bu)
	sumtS(au, bu) // sorty
	sumtLswS(au, bu)

	// Is Sort*() multi-goroutine safe?
	fmt.Println("\nConcurrent calls to Sort*()")
	name = "multi"
	K, L, ch := N/2, N/4, make(chan bool)
	Mxg = 2

	// two concurrent calls to SortU8() & SortF8() each
	// up to 8 goroutines total
	go sasU8(21, bu2[:L:L], ch)
	go sasF8(22, af2[:L:L], ch)
	go sasU8(21, bu2[L:], ch)
	fstF8(22, af2[L:], SortF8)

	for i := 3; i > 0; i-- {
		<-ch // wait others
	}
	compareU4(bu[:K:K], bu[K:]) // same buffers
	compareU4(au[:K:K], au[K:])

	// two concurrent calls to SortI4() & SortI8() each
	// up to 8 goroutines total
	go sasI4(23, ai[:K:K], ch)
	go sasI8(24, bi2[:L:L], ch)
	go sasI4(23, ai[K:], ch)
	fstI8(24, bi2[L:], SortI8)

	for i := 3; i > 0; i-- {
		<-ch // wait others
	}
	compareU4(bu[:K:K], bu[K:]) // same buffers
	compareU4(au[:K:K], au[K:])

	// Sort()ing short arrays
	for l := -3; l < 2; l++ {
		Sort(l, iarlsw)
		if iar[0] != 9 || iar[1] != 8 {
			t.Fatal("Sort()ing short arrays does not work")
		}
	}
	for l := 2; l < 4; l++ {
		Sort(l, iarlsw)
		for k := 2; k >= 0; k-- {
			if iar[k] != iar[12+k-l] {
				t.Fatal("Sort()ing short arrays does not work")
			}
		}
	}

	// SortI() calls SortI4() (on 32-bit) or SortI8() (on 64-bit).
	SortI(iar)
	if 0 != IsSortedI(iar) {
		t.Fatal("SortI() does not work")
	}

	// test Search()
	n := len(iar)
	k := Search(n, func(i int) bool { return iar[i] >= 5 })
	l := Search(n, func(i int) bool { return iar[i] >= 10 })
	if k <= 0 || k >= n || iar[k] != 5 || iar[k-1] != 4 || l != n {
		t.Fatal("Search() does not work")
	}
}

var iar = []int{
	9, 8, 7, 6, 5, 4, 3, 2, 1, 7, 8, 9, 7, 6, 5, 4, 3, 2, 1, 0, 1, 2, 0, -1, 1, 2, 0,
	-9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 1, 2, 0, -1,
	9, 8, 7, 6, 5, 4, 3, 2, 1, 0, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 0, -1,
	-9, 8, -7, 6, -5, 4, -3, 2, -1, 0, 9, -8, 7, -6, 5, -4, 3, -2, 1, 0, 1, 2, 0, -1,
	9, 8, 7, 6, 5, 4, 3, 2, 1, 0, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 0, -1,
	-9, -8, -7, -6, -5, -4, -3, -2, -1, 0, -9, -8, -7, -6, -5, -4, -3, -2, -1, 0, -9}

func iarlsw(i, k, r, s int) bool {
	if iar[i] < iar[k] {
		if r != s {
			iar[r], iar[s] = iar[s], iar[r]
		}
		return true
	}
	return false
}

func printOpt(x, y int, v float64) {
	fmt.Printf("%3d %3d %5.2fs\n", x, y, v)
}

// Optimize max array lengths for insertion sort/recursion (Mli,Mlr)
// takes a long time, run without -short flag
func TestOpt(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	tst = t

	as := make([]float32, N)
	aq := make([]float32, N)
	ar := F4toU4(&as)
	ap := make([]uint32, N)

	nm := [...]string{"SortU4/F4", "Lsw-U4/F4", "SortS", "Lsw-S"}
	fn := [...]func() float64{
		// optimize for native arithmetic types
		func() float64 { return sumtU4(ar, ap[:0]) + sumtF4(as, aq[:0]) },

		// optimize for function-based sort
		// carry over ap,aq for further comparison
		func() float64 { return sumtLswU4(ar, ap) + sumtLswF4(as, aq) },

		// optimize for native string
		func() float64 { return sumtS(ar, ap[:0]) },

		// optimize for function-based sort (string key)
		// carry over ap for further comparison
		func() float64 { return sumtLswS(ar, ap) }}

	s1, s2 := "Mli", 96

	for i := 0; i < len(fn); i++ {
		fmt.Printf("\n%s\n%s Mlr:\n", nm[i], s1)

		_, _, _, n := opt.FindMinTri(2, s2, 480, 16, 120,
			func(x, y int) float64 {
				Mli, Hmli, Mlr = x, x, y
				return fn[i]()
			}, printOpt)
		fmt.Println(n, "calls")

		s1, s2 = "Hmli", 48
	}
}
