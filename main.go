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
func fileContentsComparison(a, b io.Reader) (bool, error) {
	bufferSize := os.Getpagesize()
	ba := make([]byte, bufferSize)
	bb := make([]byte, bufferSize)
	for {
		la, erra := a.Read(ba)
		lb, errb := b.Read(bb)
		if la > 0 || lb > 0 {
			if !bytes.Equal(ba, bb) {
				return false, nil
			}
		}
		switch {
		case erra == io.EOF && errb == io.EOF:
			return true, nil
		case erra != nil && errb != nil:
			return false, errors.New(erra.Error() + " " + errb.Error())
		case erra != nil:
			return false, erra
		case errb != nil:
			return false, errb
		}
	}
}
func isSameFile(f1 os.FileInfo, f2 os.FileInfo, Folder string) (bool, error) {
	if f1.Size()!=f2.Size(){
		return false, nil
	}
	a, erra := os.Open("./"+Folder+"/"+f1.Name())
	b, errb := os.Open("./"+Folder+"/"+f2.Name())
	if erra != nil {
		return false, erra
	}
	if errb != nil {
		return false, errb
	}
	r, err := fileContentsComparison(a, b)
	a.Close()
	b.Close()
	return r, err
}
func checkDuplicates(t string, Folder string){
	f1,_ := os.Stat(t)
	files, _ := ioutil.ReadDir("./"+Folder)
	for _, f2 := range files{
		if f1.Name()!=f2.Name(){
			r, err := isSameFile(f1, f2, Folder)
			if !r {
				if err != nil {
					log(f1.Name() + " " + f2.Name() + " " + err.Error())
				}
			} else {
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
}
func download(filename string, url string, Folder string)error{
	Client := &http.Client{}
	resp, errhttp := Client.Get(url)
	out, errf := os.Create(filename)
	if errhttp==nil && errf==nil{
		_, err := io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
		if err==nil{
			checkDuplicates(filename, Folder)
			return nil
		}else{
			os.Remove(filename)
			return err
		}
	}else{
		switch {
		case errhttp != nil && errf != nil:
			return errors.New(errhttp.Error() + " " + errf.Error())
		case errhttp != nil:
			return errhttp
		case errf != nil:
			return errf
		}
		return errors.New("ukn")
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
func checkConfig(config *Config, file string) (bool, error) {
	t, err := os.Stat(file)
	if err != nil {
		return false, err
	}
	if t.ModTime().After(config.LastModified) {
		fc, err := os.Open(file)
		if err != nil {
			return false, err
		}
		err = json.NewDecoder(fc).Decode(config)
		fc.Close()
		if err!=nil{
			return false, err
		}
		http.DefaultClient = &http.Client{
			Timeout : config.Timeout * time.Second,
		}
		config.ParsedKeywords = nil
		config.ParsedKeywords = make(map[string]string)
		for k,v := range config.Keywords{
			for _,v2 := range v{
				config.ParsedKeywords[v2] = k
			}
		}
		if config.MinTimeBetweenUpdates == 0 {
			config.MinTimeBetweenUpdates = 10
		}
		if config.Timeout == 0 {
			config.Timeout = 10
		}
		config.LastModified = t.ModTime()
		return true, nil
	}
	return false, nil
}
func checkKeywords(config *Config, decoded_thread Thread)string{
	for v2, k := range config.ParsedKeywords {
		if strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Sub), strings.ToUpper(v2)) || strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Com), strings.ToUpper(v2)){
			return k
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
								Filename: strconv.Itoa(post.Tim) + "_" + post.Filename + post.Ext,
								Folder: folder,
								Url: "https://i.4cdn.org/b/" + strconv.Itoa(post.Tim) + post.Ext, 
								Size: post.Fsize,
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
	go download_loop(dl)
	//
	var real_last_update int64 = 0
	var last_update int = 0
	var new_last_update int = 0
	var decoded_pages Pages
	threads_req, _ := http.NewRequest("GET", "https://a.4cdn.org/b/threads.json", nil)
	var wg sync.WaitGroup
	//
	for {
		// load config
		ok, err := checkConfig(&config, "config.json")
		if ok {
			log("New config")
			println("\tTimeout: " + strconv.Itoa(int(config.Timeout)) + "s")
			println("\tMinimum time between updates: " + strconv.Itoa(int(config.MinTimeBetweenUpdates)) + "s")
			fmt.Println(config)
		} else if err != nil {
			log(err.Error())
		}
		real_last_update=time.Now().Unix()
		r, err := http.DefaultClient.Do(threads_req)
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
	ParsedKeywords map[string]string
	//
	MinTimeBetweenUpdates int64 `json:"min_time_between_updates"`
	Timeout time.Duration `json:"timeout"`
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
		MD5 string `json:"md5"`
	} `json:"posts"`
}