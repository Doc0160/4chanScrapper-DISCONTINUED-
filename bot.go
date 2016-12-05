/* ========================================================================
   $File: $
   $Date: $
   $Revision: $
   $Creator: Tristan Magniez $
   ======================================================================== */

package main

import (
    "sync"
    "log"
    "strconv"
    "time"
    "strings"
    "os"
)
var _ = time.NewTicker

var bot = NewBot()
type Bot struct {
    sync.Mutex
    tasks chan Task
    last_updates map[int]int
    tickets TicketMutex
}
func NewBot() Bot {
    b := Bot{}
    b.tasks = make(chan Task, 1000)
    b.last_updates = make(map[int]int)
    return b
}

type Task interface{}

type Picture struct {
    Tries uint8
	Filename string
	Folder string
	Url string
	Size int
}

func (b*Bot)t(){
    ticker := time.NewTicker(time.Second * 10)
    ticker2 := time.NewTicker(time.Second * 1)
    for {
        select {
        case <-ticker2.C:
            if b.tickets.Current() > 10 {
                //log.Println(len(b.tasks), b.tickets.Current())
            }
            /*if len(b.tasks) > 0 {
                log.Println(len(b.tasks))
            }*/
            
        case <-ticker.C:
            /*go func() {
                ticket := b.tickets.Acquire()
                defer b.tickets.Release(ticket)
                FullDupeCheck(config)
            }()*/
            
        case task := <- b.tasks:
            if !b.tickets.AllDone() {
                //log.Println(len(b.tasks), b.tickets.Current())
            }
            switch task.(type) {
            case Picture:
                go func(){
                    ticket := b.tickets.Acquire()
                    var err error
                    file := task.(Picture)
                    t := "./"+file.Folder+"/"+file.Filename
                    if strings.Contains(file.Filename,"."){
                        if _, err = os.Stat(t); os.IsNotExist(err) {
                            log.Println("Downloading", file.Filename)
                            err := download(t, file.Url, file.Folder)
                            if err==nil{
                                log.Println("Downloaded", file.Filename + " " +
                                    format_o(file.Size))
                            }else{
                                file.Tries++
                                if file.Tries < config.DownloadRetries {
                                    b.tasks<-file
                                    println("Retry", file.Filename)
                                }
                                log.Println("ERROR : " + file.Filename + " " + err.Error())
                            } 
                        }
                    }
                    b.tickets.Release(ticket)
                }()
                
            case PageThread:
                go func(){
                    ticket := b.tickets.Acquire()
                    decoded_thread := Thread{}
                    pthread := task.(PageThread)
                    err := decoded_thread.Load(strconv.Itoa(pthread.No))
                    if err == nil {
                        if decoded_thread.Posts != nil {
                            folder := checkKeywords(&config, decoded_thread);
                            if folder != "" {
                                for _, post := range decoded_thread.Posts {
                                    if post.Time >
                                        last_updates[pthread.No] &&
                                        post.Tim != 0 {
                                        b.tasks<-Picture{
                                            Filename: CleanName(
                                                strconv.Itoa(post.Tim) + "_" +
                                                    post.Filename) + post.Ext,
                                            Folder: folder,
                                            Url: post.GetFileUrl(), 
                                            Size: post.Fsize,
                                        }
                                    }
                                }
                            }
                        }
                        b.Lock()
                        last_updates[pthread.No] = pthread.LastModified
                        b.Unlock()
                    } else {
                        log.Println("ERROR : " + err.Error())
                    }
                    b.tickets.Release(ticket)
                }()
            default:
                log.Println(task)
            }
        }
    }
}
