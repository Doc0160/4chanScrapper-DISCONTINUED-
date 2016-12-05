/* ========================================================================
   $File: $
   $Date: $
   $Revision: $
   $Creator: Tristan Magniez $
   ======================================================================== */

package main

import (
    "sync"
    "sync/atomic"
    "time"
)

type TicketMutex struct{
    sync.Mutex
    now_serving uint32
    next_ticket uint32
}

func (m*TicketMutex)Release(ticket uint32) {
    if ticket < 10 {
        println(ticket)
    }
    // TODO
    atomic.AddUint32(&m.now_serving, 1)
    /*for !atomic.CompareAndSwapUint32(&m.now_serving, ticket, ticket+1) {
        time.Sleep(100)
    }*/
}

func (m*TicketMutex)Acquire() uint32 {
    ticket := 1-atomic.AddUint32(&m.next_ticket, 1)
    return ticket
}

func (m*TicketMutex)AllDone() bool {
    return atomic.LoadUint32(&m.next_ticket) == atomic.LoadUint32(&m.now_serving)
}

func (m*TicketMutex)Current() uint32 {
    //println(atomic.LoadUint32(&m.next_ticket), atomic.LoadUint32(&m.now_serving))
    return atomic.LoadUint32(&m.next_ticket)-atomic.LoadUint32(&m.now_serving)
}

func (m*TicketMutex)Wait(til uint32){
    for m.Current() > til {
        time.Sleep(100)
    }
}
