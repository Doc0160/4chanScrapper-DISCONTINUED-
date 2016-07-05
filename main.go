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
	"runtime/debug"
	"errors"
	"sync"
)
func log(t string){
	print("[")
	print(time.Now().UnixNano())
	print("] ")
	print(t)
	print(" ")
	println()
}
func isSameFile(f1 os.FileInfo, f2 os.FileInfo, Folder string)bool{
	if f1.Size()!=f2.Size(){
		return false
	}
	sf, err := os.Open("./"+Folder+"/"+f1.Name())
	defer sf.Close()
	if err != nil {
		log(err.Error())
		return false
	}
	df, err := os.Open("./"+Folder+"/"+f2.Name())
	defer df.Close()
	if err != nil {
		log(err.Error())
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
}
func checkDuplicates(t string, Folder string){
	f1,_ := os.Stat(t)
	files, _ := ioutil.ReadDir("./"+Folder)
	for _, f2 := range files{
		// big deep check byte per byte if necessary to determine if two files are the same
		if f1.Name()!=f2.Name() && isSameFile(f1, f2, Folder) {
			if f1.ModTime().Before(f2.ModTime()) {
				log(f2.Name() + " duplicate removed")
 				os.Remove("./"+Folder+"/"+f1.Name())
 			}else{
				log(f1.Name() + " duplicate removed")
 				os.Remove("./"+Folder+"/"+f2.Name())
 			}
		}
	}
}
func download(filename string, url string, Folder string)error{
	resp, err := http.Get(url)
	if err==nil{
		defer resp.Body.Close()
		out, err := os.Create(filename)
		if err==nil{
			defer out.Close()
			_, err := io.Copy(out, resp.Body)
			if err==nil{
				checkDuplicates(filename, Folder)
				return nil
			}else{
				defer os.Remove(filename)
				return err
			}
		}else{
			defer os.Remove(filename)
			return err
		}
	}else{
		return err
	}
}
func download_loop(dl chan dl_struct){
	debug.SetGCPercent(500)
	runtime.LockOSThread()
	var err error
	var file dl_struct
	for {
		file = <-dl
		os.MkdirAll("./"+file.Folder+"/", 0777)
		t := "./"+file.Folder+"/"+file.Filename
		if strings.Contains(file.Filename,"."){
			if _, err = os.Stat(t); os.IsNotExist(err) {
				log(file.Filename + " downloading")
				err := download(t, file.Url, file.Folder)
				if err==nil{
					log(file.Filename + " downloaded " + format_o(file.Size))
				}else{
					log(file.Filename + " " + err.Error())
				}
			}
		}
	}
	runtime.UnlockOSThread()
}
func checkConfig(config *Config, file string)error{
	t, err := os.Stat(file)
	if err != nil {
		log(err.Error())
		return err
	}
	if t.ModTime().After(config.LastModified) {
		fc, err := os.Open(file)
		if err != nil {
			log(err.Error())
			return err
		}
		err = json.NewDecoder(fc).Decode(config)
		fc.Close()
		if err!=nil{
			log(err.Error())
			return err
		}
		config.LastModified = t.ModTime()
		return nil
	}
	return errors.New("")
}
func checkKeywords(config *Config, decoded_thread Thread)string{
	for k,v := range config.Keywords{
		for _,v2 := range v{
			if strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Sub), strings.ToUpper(v2)) || strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Com), strings.ToUpper(v2)){
				return k
			}
		}
	}
	return ""
}
func doThread(wg *sync.WaitGroup, threadURL string, dl chan dl_struct, config Config){
	defer wg.Done()
	r, err := http.Get(threadURL)
	if err==io.EOF{
	}else if err!=nil{
		log(err.Error())
	}else{
		decoded_thread := Thread{}
		err = json.NewDecoder(r.Body).Decode(&decoded_thread)
		if err==io.EOF {
		}else if err!=nil {
			log(err.Error())
		}else{
			r.Body.Close()
			if decoded_thread.Posts!=nil {
				folder := checkKeywords(&config, decoded_thread);
				if folder!="" {
					for _, post := range decoded_thread.Posts {
						if post.Tim!=0 {
							dl<-dl_struct{
								strconv.Itoa(post.Tim) + post.Ext,
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
}
func main(){
	debug.SetGCPercent(500)
	var config Config
	// download every picture sent in dl chan
	dl:=make(chan dl_struct,100)
	go download_loop(dl)
	//
	var real_last_update int64 = 0
	var last_update int = 0
	var new_last_update int = 0
	client := &http.Client{
		Timeout : 10 * time.Second,
	}
	var decoded_pages Pages
	// var decoded_thread Thread
	threads_req, _ := http.NewRequest("GET", "https://a.4cdn.org/b/threads.json", nil)
	var wg sync.WaitGroup
	//
	for {
		// load config
		if err := checkConfig(&config, "config.json"); err == nil{
			fmt.Println(config)
		}
		real_last_update=time.Now().Unix()
		r, err := client.Do(threads_req)
		if err==io.EOF{
		}else if err!=nil{
			log(err.Error())
		}else{
			err = json.NewDecoder(r.Body).Decode(&decoded_pages)
			r.Body.Close()
			if err==io.EOF{
			}else if err!=nil{
				log(err.Error())
			}else{
				for _,page := range decoded_pages{
					for _, thread := range page.Threads {
						if thread.LastModified>last_update {
							wg.Add(1)
							doThread(&wg, "https://a.4cdn.org/b/thread/"+strconv.Itoa(thread.No)+".json", dl, config)
							if thread.LastModified>new_last_update{
								new_last_update=thread.LastModified
							}
						}
					}
				}
				wg.Wait()
				last_update=new_last_update
			}
		}
		log("Check took " + strconv.Itoa(int(time.Now().Unix() - real_last_update)) + "s")
		pause(config.MinTimeBetweenUpdates - (time.Now().Unix() - real_last_update))
	}
}
type dl_struct struct{
	Filename string
	Folder string
	Url string
	Size int
}
type Config struct {
	LastModified time.Time
	MinTimeBetweenUpdates int64 `json:"min_time_between_updates"`
	Keywords map[string][]string `json:"keywords"`
}
func pause(durationS int64) {
	time.Sleep(time.Duration(durationS) * time.Second)
}
func format_o(i int)string {
	switch{
		case i>1024*1024*1024*1024:
		return strconv.Itoa(i/(1024*1024*1024*1024))+"To"
		case i>1024*1024*1024:
		return strconv.Itoa(i/(1024*1024*1024))+"Go"
		case i>1024*1024:
		return strconv.Itoa(i/(1024*1024))+"Mo"
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
		Time int `json:"time"`
		Name string `json:"name"`
		Sub string `json:"sub"`
		Com string `json:"com"`
		Filename string `json:"filename"`
		Tim int `json:"tim"`
		Ext string `json:"ext"`
		Fsize int `json:"fsize"`
	} `json:"posts"`
}