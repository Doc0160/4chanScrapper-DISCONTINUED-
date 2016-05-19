package main
import(
	"fmt"
	"net/http"
	"strconv"
	"io/ioutil"
	// "encoding/json"
	"strings"
	// "time"
	"os"
	// "io"
	"bufio"
	"bytes"
	// "path"
	"net/url"
    "html/template"
	"github.com/gorilla/mux"
	"sort"
)
type PairList []os.FileInfo
func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].ModTime().After(p[j].ModTime()) }
type Config struct{
	MinTimeBetweenUpdates int64               `json:"min_time_between_updates"`
	Port                  int                 `json:"port"`
	Keywords              map[string][]string `json:"keywords"`
}
func is_it_the_right_thread(Keyword string, post Post)bool{
	return (strings.Contains(strings.ToUpper(post.Sub), strings.ToUpper(Keyword)) || strings.Contains(strings.ToUpper(post.Com), strings.ToUpper(Keyword)))
}
func what_is_the_right_thread(Keywords map[string][]string, post Post)string{
	for k,v := range Keywords{
		for _,v2 := range v{
			if is_it_the_right_thread(v2, post){
				return k
			}
		}
	}
	return ""
}
// NOTE(doc): return true if names(path included) are the same
//	false if size are different
//	else go for a deep check et stop when at least 1 byte is different
func compare_files(f1 os.FileInfo, f2 os.FileInfo, Folder string)bool{
	if f1.Name()==f2.Name(){
		return true
	}
	if f1.Size()!=f2.Size(){
		return false
	}
	sf, err := os.Open("./"+Folder+"/"+f1.Name())
	defer sf.Close()
	if err != nil {
		fmt.Println(err)
		return false
	}
	df, err := os.Open("./"+Folder+"/"+f2.Name())
	defer df.Close()
	if err != nil {
		fmt.Println(err)
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
// NOTE(doc): if duplicate found, delete the old copy
func check_for_duplicates(t string, Folder string){
	f1,_ := os.Stat(t)
	files, _ := ioutil.ReadDir("./"+Folder)
	for _, f2 := range files{
		if f1.Name()!=f2.Name() && compare_files(f1,f2,Folder){
			if f1.ModTime().Before(f2.ModTime()){
				fmt.Println("\""+shorter(f2.Name(), 30)+"\"", "duplicate removed")
				os.Remove("./"+Folder+"/"+f1.Name())
			}else{
				fmt.Println("\""+shorter(f1.Name(), 30)+"\"", "duplicate removed")
				os.Remove("./"+Folder+"/"+f2.Name())
			}
		}
	}
}
func file_exist(t string)bool{
	_, err := os.Stat(t)
	return !os.IsNotExist(err)
}
func main(){
	Download_queue := make(chan DownloadTask,100)
	var fce FourChanExplorer
	var dler Downloader
	dler.Queue = Download_queue
	fce.Queue = Download_queue
	err := fce.Init()
	if err!=nil {
		fmt.Println(err)
	}else{
		// download every picture sent in dl chan
		go dler.StartLoop()
		go fce.StartLoop()
		fmt.Println(fce.Config)
		//
		t, err := template.ParseFiles("basic.tpl");
		if err!=nil{
			fmt.Println(err)
		}
		r := mux.NewRouter()
		r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
			files, _ := ioutil.ReadDir("./")
			type Data struct{
				Title string
				Body string
				List []struct{
					Name string
					URL string
				}
			}
			var data Data
			data.Title ="index"
			data.Body ="List:"
			for _, f := range files{
				if f.IsDir() && f.Name()[0]!='.'{
					var l struct{
						Name string
						URL string
					}
					l.Name=f.Name()
					l.URL="./"+url.QueryEscape(f.Name())
					data.List=append(data.List, l)
				}
			}
			err := t.ExecuteTemplate(w, "basic_list", data);
			if err!=nil{
				fmt.Println(err)
			}
		})
		a := func(w http.ResponseWriter, r *http.Request, nb int){
			PP:=200
			vars := mux.Vars(r)
			type Data struct{
				Title string
				Body string
				Prev string
				Next string
				List []struct{
					Name string
					URL string
				}
			}
			var data Data
			data.Title=vars["folder"]
			data.Body="Page "+strconv.Itoa(nb)+"\n"+vars["folder"]
			data.Prev="./"+strconv.Itoa(nb-1)
			data.Next="./"+strconv.Itoa(nb+1)
			i:=0
			j:=0
			var files PairList
			files, _ = ioutil.ReadDir("./"+vars["folder"]+"/")
			sort.Sort(files)
			for _, f := range files{
				if !f.IsDir() && f.Name()[0]!='.' && j>=PP*(nb-1) && j<PP*nb {
					var l struct{
						Name string
						URL string
					}
					l.Name=f.Name()
					l.URL="/static/"+vars["folder"]+"/"+f.Name()
					data.List=append(data.List, l)
					i++
				}
				j++
			}
			data.Body+="("+strconv.Itoa(PP*(nb-1)+i)+"/"+strconv.Itoa(j)+"):"
			err = t.ExecuteTemplate(w, "files", data);
			if err!=nil{
				fmt.Println(err)
			}
		}
		r.HandleFunc("/{folder}/", func(w http.ResponseWriter, r *http.Request){
			a(w, r, 1)
		})
		r.HandleFunc("/{folder}/{nb}", func(w http.ResponseWriter, r *http.Request){
			vars := mux.Vars(r)
			b,_:=strconv.Atoi(vars["nb"])
			a(w, r, b)
		})
		for _,v := range fce.Config.Keywords{
			for _,v2 := range v{
				// r.PathPrefix("/static/"+v2).Handler(http.FileServer(http.Dir("./")))
				r.PathPrefix("/static/"+v2).Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))
			}
		}
		http.ListenAndServe(":"+strconv.Itoa(fce.Config.Port), r)
	}
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
	// NOTE(doc): go can't create files with a name>255 on windows -bug report
	return replace(shorter(p.Filename,200)+p.Ext, ' ', '_')
}
func (p *Post)GetImgURL()string{
	return "https://i.4cdn.org/b/" + p.GetFilenameWithExt()
}
func (p *Post)GetFilenameWithExt()string{
	return strconv.Itoa(p.Tim)+p.Ext
}
func replace(s string, from byte, to byte)string{
	var newS []byte
	for i:=0; i<len(s); i++{
		if(s[i]==from){
			newS=append(newS, to)
		}else{
			newS=append(newS, s[i])
		}
	}
	return string(newS)
}
func shorter(s string, i int) string {
	if(len(s)<i){
		return s
	}else{
		return s[:i-3]+"..."
	}
}