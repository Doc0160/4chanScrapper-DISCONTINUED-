package main
import(
	"os"
	"encoding/json"
	"net/http"
	"time"
	"fmt"
	"io"
	"strconv"
)
type FourChanExplorer struct{
	Config Config
	// NOTE(doc): resusing same client take less time
	Client http.Client
	Pages Pages
	Thread Thread
	LastUpdate int
	Loop bool
	Queue chan DownloadTask
}
func (fce *FourChanExplorer)Init()error{
	fce.LastUpdate=0
	fce.Loop=false
	return fce.LoadConfig()
}
// NOTE(doc): check struct Config to know how to structure the json file
func (fce *FourChanExplorer)LoadConfig()error{
	fc, err := os.Open("config.json")
	if err != nil {
		if os.IsNotExist(err){
			return err
		}else{
			return err
		}
	}
	err = json.NewDecoder(fc).Decode(&fce.Config)
	fc.Close()
	if err!=nil{
		return err
	}
	// return config
	return nil
}

func (fce *FourChanExplorer)StartLoop(){
	fce.Loop = true
	var real_last_update int64 = 0
	var new_last_update int = 0
	for fce.Loop{
		// SIDENOTE(doc): LIBERTAD DEL PROCO
		if !(time.Now().Unix()-real_last_update>fce.Config.MinTimeBetweenUpdates){
			time.Sleep(time.Duration((fce.Config.MinTimeBetweenUpdates-(time.Now().Unix()-real_last_update)))*time.Second)
		}
		real_last_update=time.Now().Unix()
		r, err := fce.Client.Get("https://a.4cdn.org/b/threads.json")
		if err!=nil{
			// TODO(doc): err, can't connect to webpage
			fmt.Println(1,err)
		}else{
			fce.Pages = Pages{}
			err = json.NewDecoder(r.Body).Decode(&fce.Pages)
			r.Body.Close()
			if err!=nil{
				// TODO(doc): err, can't decode webpage / not a valid json document
				fmt.Println(2,err)
			}else{
				for _,page:=range fce.Pages{
					for _, thread := range page.Threads{
						if thread.LastModified>fce.LastUpdate{
							r, err := fce.Client.Get("https://a.4cdn.org/b/thread/"+strconv.Itoa(thread.No)+".json")
							if err!=nil{
								// TODO(doc): err, can't connect webpage
								fmt.Println(3,err)
							}else{
								// NOTE(doc): reuse var, bc no explicit free in go
								fce.Thread = Thread{}
								err = json.NewDecoder(r.Body).Decode(&fce.Thread)
								r.Body.Close()
								switch(err){
								case io.EOF: // NOTE(doc): does nothing in go, but does'nt end in default
									// TODO(doc): WHY??? JSON---EOF ??
								case nil:
									if fce.Thread.Posts!=nil{
										folder := what_is_the_right_thread(fce.Config.Keywords, fce.Thread.Posts[0])
										if folder!=""{
											for _, post := range fce.Thread.Posts{
												// NOTE(doc): check if image exist in the post
												if post.Tim!=0{
													filename:=strconv.Itoa(post.Tim) + " - " + post.GetOriginalFilenameWithExt()
													// NOTE(doc): checking if already downloaded
													if !file_exist(filename){
														fce.Queue<-DownloadTask{
															folder+"/"+filename,
															post.GetImgURL(),
														}
													}
												}
											}
										}
									}else{
										// TODO(doc): retry
									}
								default:
									// TODO(doc): err, can't decode
									fmt.Println(4,err)
								}
							}
							if thread.LastModified>new_last_update{
								new_last_update=thread.LastModified
							}
						}
					}
				}
				fce.LastUpdate=new_last_update
			}
		}
	}
}