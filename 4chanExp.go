package main
import(
	// "os"
	"encoding/json"
	"net/http"
	"time"
	"fmt"
	"io"
	"strconv"
	// "net/url"
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
func (fce *FourChanExplorer)StartLoop(){
	fce.LastUpdate=0
	fce.Loop = true
	var real_last_update int64 = 0
	var new_last_update int = 0
	for fce.Loop{
		real_last_update=time.Now().Unix()
		// SIDENOTE(doc): LIBERTAD DEL PROCO
		if !(time.Now().Unix()-real_last_update>fce.Config.MinTimeBetweenUpdates){
			time.Sleep(time.Duration(fce.Config.MinTimeBetweenUpdates)*time.Second)
		}
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
							var fail bool
							for !fail{
								fail=false
								r, err := fce.Client.Get("https://a.4cdn.org/b/thread/"+strconv.Itoa(thread.No)+".json")
								if err!=nil{
									// TODO(doc): err, can't connect webpage
									fail=true
									fmt.Println(3,err)
								}else{
									// NOTE(doc): reuse var, bc no explicit free in go
									fce.Thread = Thread{}
									err = json.NewDecoder(r.Body).Decode(&fce.Thread)
									r.Body.Close()
									switch(err){
									case io.EOF: // NOTE(doc): does nothing in go, but does'nt end in default
										// TODO(doc): WHY??? JSON---EOF ??
										fail=true
									case nil:
										if fce.Thread.Posts!=nil{
											var folder string = ""
											for k,v := range fce.Config.Keywords{
												for _,v2 := range v{
													if is_it_the_right_thread/*yet?*/(v2, fce.Thread.Posts[0]){
														//return k //YEEES!!!
														folder=k
														break
													}
													// NO è_é
												}
											}
											// return "" // 'K bye 
											//folder := what_is_the_right_thread(fce.Config.Keywords, fce.Thread.Posts[0])
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
											fail=true
										}
									default:
										fail=true
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
				}
				fce.LastUpdate=new_last_update
			}
		}
	}
}
// https://github.com/4chan/4chan-API
type Pages []struct {
	Page int `json:"page"`
	Threads []struct {
		No int `json:"no"`
		LastModified int `json:"last_modified"`
	} `json:"threads"`
}
type Thread struct{
	Posts []Post `json:"posts"`
}
type Post struct{
	No int `json:"no"`
	Resto int `json:"resto"`
	Time int `json:"time"`
	Name string `json:"name"`
	Sub string `json:"sub"`
	Com string `json:"com"`
	Filename string `json:"filename"`
	Tim int `json:"tim"`
	Ext string `json:"ext"`
	Fsize int `json:"fsize"`
	Country int `json:"country"`
	Country_name int `json:"country_name"`
}
func (p *Post)GetOriginalFilenameWithExt()string{
	//p.Filename, err = url.QueryUnescape(p.Filename)
	p.Filename = misc.CorrectName(p.Filename)
	// NOTE(doc): go can't create files with a name>255 on windows -bug report
	p.Filename = misc.Shortener(p.Filename,200)
	return misc.Replace(p.Filename+p.Ext, ' ', '_')
}
func (p *Post)GetImgURL()string{
	return "https://i.4cdn.org/b/" + p.GetFilenameWithExt()
}
func (p *Post)GetFilenameWithExt()string{
	return strconv.Itoa(p.Tim)+p.Ext
}