package main
import(
	"fmt"
	// "strconv"
	"io/ioutil"
	"encoding/json"
	"strings"
	// "time"
	"os"
	// "io"
	"bufio"
	"bytes"
	// "path"
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
			if is_it_the_right_thread/*yet?*/(v2, post){
				return k //YEEES!!!
			}
			// NO è_é
		}
	}
	return "" // 'K bye 
}
// NOTE(doc): return true if names(path included) are the same
//	false if size are different
//	else go for a deep check et stop when at least 1 byte is different
func compare_files(f1 os.FileInfo, f2 os.FileInfo, Folder string)bool{
	if os.SameFile(f1,f2){
		return true
	}
	// TODO(doc): reconsider that?
	/*if f1.Name()==f2.Name(){
		return true
	}*/
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
	// TDO(doc): check if dscan as more bytes
	return true
}
// NOTE(doc): if duplicate found, delete the old copy
func check_for_duplicates(t string, Folder string){
	f1,_ := os.Stat(t)
	files, _ := ioutil.ReadDir("./"+Folder)
	for _, f2 := range files{
		nn := misc.CorrectName(f2.Name())
		if nn!=f2.Name(){
			print("corrected")
			os.Rename("./"+Folder+"/"+f2.Name(), "./"+Folder+"/"+nn)
		}else if f1.Name()!=f2.Name() && compare_files(f1,f2,Folder){
			if f1.ModTime().Before(f2.ModTime()){
				fmt.Println("\""+f2.Name()+"\"", "duplicate removed")
				os.Remove("./"+Folder+"/"+f1.Name())
			}else{
				fmt.Println("\""+f1.Name()+"\"", "duplicate removed")
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
	Config, err := LoadConfig();
	if err==nil{
		Download_queue := make(chan DownloadTask,100)
		var fce FourChanExplorer
		var dler Downloader
		var srv Serv
		fce.Config = Config
		srv.Config = Config
		dler.Queue = Download_queue
		fce.Queue = Download_queue
		//
		fmt.Println(fce.Config)
		go dler.StartLoop()
		go fce.StartLoop()
		http_serve(&fce.Config)
	}else{
		fmt.Println(err)
	}
}
/*func format_o(i int)string{
	switch{
		case i>1073741824:
		return strconv.Itoa(i/1073741824)+"Go"
		case i>1048576:
		return strconv.Itoa(i/1048576)+"Mo"
		case i>1024:
		return strconv.Itoa(i/1024)+"ko"
	}
	return strconv.Itoa(i)+"o"
}*/
// NOTE(doc): check struct Config to know how to structure the json file
func LoadConfig()(Config,error){
	var c Config
	fc, err := os.Open("config.json")
	if err != nil {
		if os.IsNotExist(err){
			return c,err
		}else{
			return c,err
		}
	}
	err = json.NewDecoder(fc).Decode(&c)
	fc.Close()
	if err!=nil{
		return c,err
	}
	// return config
	return c,nil
}