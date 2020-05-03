/*	Copyright (c) 2019, Serhat Şevki Dinçer.
	This Source Code Form is subject to the terms of the Mozilla Public
	License, v. 2.0. If a copy of the MPL was not distributed with this
	file, You can obtain one at http://mozilla.org/MPL/2.0/.
*/

package sorty

import "sync/atomic"

// IsSorted checks if underlying collection of length n is sorted as per less().
// An existing lesswap() can be used like:
//  IsSorted(n, func(i, k int) bool { return lesswap(i, k, 0, 0) })
func IsSorted(n int, less func(i, k int) bool) bool {
	for i := n - 1; i > 0; i-- {
		if less(i, i-1) {
			return false
		}
	}
	return true
}

// Lesswap function operates on an underlying collection to be sorted as:
//  if less(i, k) { // strict ordering like < or >
//  	if r != s {
//  		swap(r, s)
//  	}
//  	return true
//  }
//  return false
type Lesswap func(i, k, r, s int) bool

// insertion sort ar[lo..hi], assumes lo < hi
func insertion(lsw Lesswap, lo, hi int) {

	for l, h := mid(lo, hi-1)-1, hi; l >= lo; {
		lsw(h, l, h, l)
		h--
		l--
	}
	for h := lo; ; {
		for l := h; lsw(l+1, l, l+1, l); {
			l--
			if l < lo {
				break
			}
		}
		h++
		if h >= hi {
			break
		}
	}
}

// arrange median-of-5 as ar[l,l+1] <= ar[m] = pivot <= ar[h-1,h]
// This allows ar[l,l+1] and ar[h-1,h] to assist pivoting of the two sub-ranges in
// next pivot5() calls: 3 new values from a sub-range, 2 expectedly good values from
// parent range. Users of pivot5() must ensure l+5 < h
func pivot5(lsw Lesswap, l, h int) (int, int, int) {
	e, c := l, mid(l, h)
	lsw(h, l, h, l)
	l++
	h--
	b, d := h, l
	lsw(h, l, h, l)

	if lsw(h+1, h, 0, 0) {
		d, e = e, d
		b++
	}
	lsw(c, e, c, e)

	if lsw(b, c, b, c) {
		d = e
	}
	lsw(c, d, c, d)

	return l + 1, c, h - 1
}

// arrange median-of-9 as ar[a-1] <= ar[l,l+1] <= ar[a] <= ar[m] = pivot <= ar[b] <=
// ar[h-1,h] <= ar[b+1] where m,a,b = mid(l,h), mid(l,m), mid(m,h). After pivot9() ar
// will look like: 2nd 3rd .. 1st 4th .. 5th=pivot .. 6th 9th .. 7th 8th.
// This allows ar[l,l+1] and ar[h-1,h] to assist pivoting of the two sub-ranges in
// next pivot9() or pivot5() calls: 7 or 3 new values from a sub-range, 2 expectedly
// good values from parent range. Users of pivot9() must ensure l+11 < h
func pivot9(lsw Lesswap, l, h int) (int, int, int) {

	s := [9]int{0, l, l + 1, 0, mid(l, h), 0, h - 1, h, 0}
	s[3], s[5] = mid(l, s[4]), mid(s[4], h)
	s[0], s[8] = s[3]-1, s[5]+1

	for i := 2; i >= 0; i-- { // insertion sort via s
		lsw(s[i+6], s[i], s[i+6], s[i])
	}
	lsw(s[1], s[0], s[1], s[0])
	for i := 2; i < len(s); i++ {
		for k, r := i-1, s[i]; lsw(r, s[k], r, s[k]); {
			r = s[k]
			k--
			if k < 0 {
				break
			}
		}
	}
	return s[3] + 1, s[4], s[5] - 1 // l,pv,h suitable for partition()
}

// partition ar[l..h] into <= and >= pivot, assumes l < pv < h
func partition(lsw Lesswap, l, pv, h int) int {
	for {
		if lsw(h, pv, 0, 0) { // avoid unnecessary comparisons
			for {
				if lsw(pv, l, h, l) {
					break
				}
				l++
				if l >= pv { // until pv & avoid it
					l++
					goto next
				}
			}
		} else if lsw(pv, l, 0, 0) { // extend ranges in balance
			for {
				h--
				if pv >= h { // until pv & avoid it
					h--
					goto next
				}
				if lsw(h, pv, h, l) {
					break
				}
			}
		}
		l++
		h--
		if l >= pv {
			l++
			break
		}
		if pv >= h {
			h--
			goto next
		}
	}
	if pv >= h {
		h--
	}

next:
	for l < h {
		if lsw(h, pv, 0, 0) { // avoid unnecessary comparisons
			for {
				if lsw(pv, l, h, l) {
					break
				}
				l++
				if l >= h {
					return l + 1
				}
			}
		} else if lsw(pv, l, 0, 0) { // extend ranges in balance
			for {
				h--
				if l >= h {
					return l
				}
				if lsw(h, pv, h, l) {
					break
				}
			}
		}
		l++
		h--
	}
	if l == h && lsw(h, pv, 0, 0) { // classify mid element
		l++
	}
	return l
}

// rearrange ar[l..a] & ar[b..h] into <= and >= pivot, assumes l <= a < pv < b <= h
// gap (a..b) expands until one of the intervals is fully consumed
func dpartition(lsw Lesswap, l, a, pv, b, h int) (int, int) {
	for {
		if lsw(b, pv, 0, 0) { // avoid unnecessary comparisons
			for {
				if lsw(pv, a, b, a) {
					break
				}
				a--
				if a < l {
					return a, b
				}
			}
		} else if lsw(pv, a, 0, 0) { // extend ranges in balance
			for {
				b++
				if b > h {
					return a, b
				}
				if lsw(b, pv, b, a) {
					break
				}
			}
		}
		a--
		b++
		if a < l || b > h {
			return a, b
		}
	}
}

// Sort concurrently sorts underlying collection of length n via lsw().
// Once for each non-trivial type you want to sort in a certain way, you
// can implement a custom sorting routine (for a slice for example) as:
//  func SortObjAsc(c []Obj) {
//  	lsw := func(i, k, r, s int) bool {
//  		if c[i].Key < c[k].Key { // your custom strict comparator (like < or >)
//  			if r != s {
//  				c[r], c[s] = c[s], c[r]
//  			}
//  			return true
//  		}
//  		return false
//  	}
//  	sorty.Sort(len(c), lsw)
//  }
func Sort(n int, lsw Lesswap) {
	var (
		mli  = Mli >> 1
		ngr  = uint32(1)               // number of sorting goroutines including this
		done chan bool                 // end signal
		srt  func(*chan int, int, int) // recursive sort function
	)

	// new-goroutine sort function
	gsrt := func(lo, hi int) {
		var par chan int // dual partitioning channel
		srt(&par, lo, hi)

		if atomic.AddUint32(&ngr, ^uint32(0)) == 0 { // decrease goroutine counter
			done <- false // we are the last, all done
		}
	}

	// dual partitioning
	dualpar := func(par *chan int, lo, hi int) int {
		a, pv, b := pivot9(lsw, lo, hi)

		// range long enough for dual partitioning? available goroutine?
		// not atomic but good enough
		m, dual := 0, hi-lo >= 8*Mlr && ngr < Mxg
		if dual {
			if atomic.AddUint32(&ngr, 1) == 0 { // increase goroutine counter
				panic("Sort: dualpar: counter overflow")
			}
			if *par == nil {
				*par = make(chan int) // make it when 1st time we need it
			}

			go func(a, b int) {
				*par <- partition(lsw, a, pv, b)

				if atomic.AddUint32(&ngr, ^uint32(0)) == 0 { // decrease goroutine counter
					panic("Sort: dualpar: counter underflow")
				}
			}(a, b)
		} else {
			m = partition(lsw, a, pv, b) // <= and >= boundary
		}

		a -= 3
		b += 3
		lo += 2
		hi -= 2
		a, b = dpartition(lsw, lo, a, pv, b, hi)
		if dual {
			m = <-*par
		}

		// only one gap is possible
		for ; lo <= a; a-- { // gap left in low range?
			if lsw(pv, a, m-1, a) {
				m--
				if m == pv { // swapped pivot when closing gap?
					pv = a // Thanks to my wife Tansu who discovered this
				}
			}
		}
		for ; b <= hi; b++ { // gap left in high range?
			if lsw(b, pv, b, m) {
				if m == pv { // swapped pivot when closing gap?
					pv = b // It took days of agony to discover these two if's :D
				}
				m++
			}
		}
		return m
	}

	srt = func(par *chan int, lo, hi int) { // assumes hi-lo >= mli
	start:
		var l int
		if hi-lo < 4*mli {
			a, pv, b := pivot5(lsw, lo, hi)
			l = partition(lsw, a, pv, b)
		} else {
			l = dualpar(par, lo, hi)
		}
		h := l - 1

		if h-lo < hi-l {
			h, hi = hi, h // [lo,h] is the longer range
			l, lo = lo, l
		}

		// branches below are optimally laid out for fewer jumps
		// at least one short range?
		if hi-l < mli {
			insertion(lsw, l, hi)

			if h-lo < mli { // two short ranges?
				insertion(lsw, lo, h)
				return
			}
			hi = h
			goto start
		}

		// range not long enough for new goroutine? max goroutines?
		// not atomic but good enough
		if hi-l < Mlr || ngr >= Mxg {
			srt(par, l, hi) // start a recursive sort on the shorter range
			hi = h
			goto start
		}

		if atomic.AddUint32(&ngr, 1) == 0 { // increase goroutine counter
			panic("Sort: counter overflow")
		}
		go gsrt(lo, h) // start a new-goroutine sort on the longer range
		lo = l
		goto start
	}

	n-- // high indice
	if n > 2*Mlr {
		done = make(chan bool, 1)
		gsrt(0, n) // start master sort
		<-done
		return
	}
	if n >= mli {
		srt(nil, 0, n) // single goroutine
		return
	}
	if n > 0 {
		insertion(lsw, 0, n) // length 2+
	}
}
