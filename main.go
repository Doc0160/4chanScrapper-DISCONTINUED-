package main
import(
	"fmt"
	"net/http"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"strings"
	"time"
	"os"
	"io"
	"bufio"
	"bytes"
	"runtime"
)
func checkDuplicates(f1 os.FileInfo, Folder string){
	// f1,_ := os.Stat(t)
	files, _ := ioutil.ReadDir("./"+Folder)
	for _, f2 := range files{
		// big deep check byte per byte if necessary to determine if two files are the same
		if f1.Name()!=f2.Name() && func()bool{
			if f1.Size()!=f2.Size(){
				return false
			}
			sf, err := os.Open("./"+Folder+"/"+f1.Name())
			defer sf.Close()
			if err != nil {
				println(err.Error)
				return false
			}
			df, err := os.Open("./"+Folder+"/"+f2.Name())
			defer df.Close()
			if err != nil {
				println(err.Error)
				return false
			}
			sscan := bufio.NewScanner(sf)
			dscan := bufio.NewScanner(df)
			for sscan.Scan() {
				dscan.Scan()
				if !bytes.Equal(sscan.Bytes(), dscan.Bytes()) {
					return false
				}
			}
			return true
		}(){
			if f1.ModTime().Before(f2.ModTime()){
				fmt.Println(f2.Name(), "duplicate removed")
				os.Remove("./"+Folder+"/"+f1.Name())
			}else{
				fmt.Println(f1.Name(), "duplicate removed")
				os.Remove("./"+Folder+"/"+f2.Name())
			}
		}
	}
}
func download(dl chan dl_struct){
	runtime.LockOSThread()
	for {
		file := <-dl
		if strings.Contains(file.Filename,"."){
			os.MkdirAll("./"+file.Folder+"/", 0777)
			t := "./"+file.Folder+"/"+file.Filename
			if _, err := os.Stat(t); os.IsNotExist(err){
				fmt.Println(file.Filename, "downloading")
				out, err := os.Create(t)
				if err==nil{
					resp, err := http.Get(file.Url)
					if err!=nil{
						println(err.Error)
						out.Close()
						os.Remove(t)
					}else{
						_, err := io.Copy(out, resp.Body)	
						out.Close()
						resp.Body.Close()
						if err==nil{
							fmt.Println(file.Filename, "downloaded")
							// check duplicates and keep the most recent one
							f1,_ := os.Stat(t)
							checkDuplicates(f1, file.Folder)
						}else{
							os.Remove(t)
						}
					}
				}else{
					println(err.Error)
				}
			}
		}
	}
	runtime.UnlockOSThread()
}
func checkConfig(c *Config){
	
}
func main(){
	var config Config
	// load config
	fc, err := os.Open("config.json")
	if err != nil {
		println(err.Error)
	}
	err = json.NewDecoder(fc).Decode(&config)
	fc.Close()
	if err!=nil{
		println(err.Error)
	}
	// download every picture sent in dl chan
	dl:=make(chan dl_struct,100)
	go download(dl)
	//
	var real_last_update int64 = 0
	var last_update int = 0
	var new_last_update int = 0
	client := &http.Client{}
	var decoded_pages Pages
	var decoded_thread Thread
	//
	for true{
		pause(config.MinTimeBetweenUpdates - (time.Now().Unix() - real_last_update))
		// time.Sleep(time.Duration(config.MinTimeBetweenUpdates - (time.Now().Unix() - real_last_update)) * time.Second)
		// for !(time.Now().Unix()-real_last_update>config.MinTimeBetweenUpdates){}
		real_last_update=time.Now().Unix()
		r, err := client.Get("https://a.4cdn.org/b/threads.json")
		if err==io.EOF{
		}else if err!=nil{
			println(err.Error)
		}else{
			err = json.NewDecoder(r.Body).Decode(&decoded_pages)
			r.Body.Close()
			if err==io.EOF{
			}else if err!=nil{
				println(err.Error)
			}else{
				for _,page:=range decoded_pages{
					for _, thread := range page.Threads{
						if thread.LastModified>last_update{
							r, err := client.Get("https://a.4cdn.org/b/thread/"+strconv.Itoa(thread.No)+".json")
							if err==io.EOF{
							}else if err!=nil{
								println(err.Error)
							}else{
								decoded_thread = Thread{}
								err = json.NewDecoder(r.Body).Decode(&decoded_thread)
								if err==io.EOF{
								}else if err!=nil{
									println(err.Error)
								}else{
									r.Body.Close()
									if decoded_thread.Posts!=nil{
										folder := func()string{
											for k,v := range config.Keywords{
												for _,v2 := range v{
													if strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Sub), strings.ToUpper(v2)) || strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Com), strings.ToUpper(v2)){
														return k
													}
												}
											}
											return ""
										}()
										if folder!=""{
											for _, post := range decoded_thread.Posts{
												if post.Tim!=0{
													dl<-dl_struct{
														strconv.Itoa(post.Tim) + "_" + post.Filename + post.Ext,
														folder,
														"https://i.4cdn.org/b/" + strconv.Itoa(post.Tim) + post.Ext, 
														post.Fsize,
													}
												}
											}
										}
									}
								}
							}
							if thread.LastModified>new_last_update{
								new_last_update=thread.LastModified
							}
						}
					}
				}
				last_update=new_last_update
			}
		}
	}
}
type dl_struct struct{
	Filename string
	Folder string
	Url string
	Size int
}
type Config struct{
	MinTimeBetweenUpdates int64 `json:"min_time_between_updates"`
	Keywords map[string][]string `json:"keywords"`
}
func pause(durationS int64){
	time.Sleep(time.Duration(durationS) * time.Second)
	// for !(durationS>config.MinTimeBetweenUpdates){}
}
func format_o(i int)string{
	switch{
		case i>1073741824:
		return strconv.Itoa(i/1073741824)+"Go"
		case i>1048576:
		return strconv.Itoa(i/1048576)+"Mo"
		case i>1024:
		return strconv.Itoa(i/1024)+"ko"
	}
	return strconv.Itoa(i)+"o"
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
	Posts []struct{
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
	} `json:"posts"`
}