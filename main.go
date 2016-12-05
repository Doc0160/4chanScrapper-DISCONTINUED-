/* ========================================================================
   $File: $
   $Date: $
   $Revision: $
   $Creator: Tristan Magniez $
   ======================================================================== */

package main

import(
	"strconv"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"os"
	"io"
	"bytes"
	//"sync"
	"regexp"
    "runtime/debug"
    "compress/gzip"
    "log"
)

var _ = log.Println
var _ = debug.SetGCPercent 

func download(filename string, url string, Folder string)error{
    client := http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err;
    }
    req.Header.Add("Accept-Encoding", "gzip")
    r, err := client.Do(req)
    if err != nil{
        return err
    }
    defer r.Body.Close()
    out, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer out.Close()
    if len(r.Header["Content-Encoding"]) > 0 {
        if r.Header["Content-Encoding"][0] == "gzip" {
            g, err := gzip.NewReader(r.Body)
            if err != nil {
                return err
            }
            defer g.Close()
            _, err = io.Copy(out, g)
            out.Close()
            r.Body.Close()
            if err==nil{
                checkDuplicates(filename, Folder)
                return nil
            }else{
                os.Remove(filename)
                return err
            }
        } else {
            panic(r.Header["Content-Encoding"])
        }
    }else {
        _, err := io.Copy(out, r.Body)
		if err==nil{
			checkDuplicates(filename, Folder)
			return nil
		}else{
			os.Remove(filename)
			return err
		}
    }
    return nil
}
/*
func download_loop(dl chan Picture, Config * Config){
	var err error
	var file Picture
	for {
		file = <-dl
		t := "./"+file.Folder+"/"+file.Filename
		if strings.Contains(file.Filename,"."){
			if _, err = os.Stat(t); os.IsNotExist(err) {
				println("Downloading", file.Filename)
				err := download(t, file.Url, file.Folder)
				if err==nil{
					println("Downloaded", file.Filename + " " + format_o(file.Size))
				}else{
					file.Tries++
					if file.Tries < Config.DownloadRetries {
						dl<-file
						println("Retry", file.Filename)
					}
					log.Println("ERROR : " + file.Filename + " " + err.Error())
				}
			}
		}
	}
}
*/
func checkKeywords(config *Config, decoded_thread Thread)string{
	for v2, k := range config.ParsedKeywords {
		a := strings.ToUpper(v2)
		if strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Sub), a) || strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Com), a){
			return k
		}
	}
	return ""
}
/*
func doThread(wg *sync.WaitGroup, threadURL string, dl chan Picture, config Config, last_update int){
	defer wg.Done()
	decoded_thread := Thread{}
	err := decoded_thread.Load(threadURL)
	if err == nil || err == io.EOF {
		if decoded_thread.Posts != nil {
			folder := checkKeywords(&config, decoded_thread);
			if folder != "" {
				for _, post := range decoded_thread.Posts {
					if post.Time > last_update && post.Tim != 0 {
						dl<-Picture{
							Filename: CleanName(strconv.Itoa(post.Tim) + "_" + post.Filename) + post.Ext,
							Folder: folder,
							Url: post.GetFileUrl(), 
							Size: post.Fsize,
						}
					}
				}
			}
		}
	} else {
		println("ERROR : " + err.Error())
	}
}
*/
func FullDupeCheck(c Config){
	for Folder, _ := range c.Keywords {
		files, _ := ioutil.ReadDir("./"+Folder)
		for _, f2 := range files{
			//println(Folder, f2.Name())
			checkDuplicates("./"+Folder+"/"+f2.Name(), Folder)
		}
	}
}

var config Config
var last_updates map[int]int = make(map[int]int)
//var dl = make(chan Picture, 100)
func main(){
	//debug.SetGCPercent(500)
    go bot.t()

	// download every picture sent in dl chan
	//dl:=make(chan dl_struct,100)
	//go download_loop(dl, &config)
	//go download_loop(dl, &config)

    //
	var real_last_update int64 = 0
	var decoded_pages Pages

	for {
		// load config
		ok, err := config.CheckConfig("config.json")
		if ok {
			// go FullDupeCheck(config)
			println("New config")
			println("\tTimeout: " + strconv.FormatInt(config.Timeout, 10) + "s")
			println("\tMinimum time between updates: " + strconv.FormatInt(config.MinTimeBetweenUpdates, 10) + "s")
			println("\tDownload retries: " + strconv.Itoa(int(config.DownloadRetries)) + " times")
			log.Println(config.Keywords)
		} else if err != nil {
			println("ERROR : " + err.Error())
		}
		real_last_update=time.Now().Unix()
		err = decoded_pages.Load()
		var neww = 0
		if err!=nil{
			println("ERROR : " + err.Error())
		}else{
			for _,page := range decoded_pages{
				for _, thread := range page.Threads {
					if last_updates[thread.No]<thread.LastModified {
                        //bot.tickets.Wait(10)
                        bot.tasks <- thread
						neww++
					}
				}
			}
        }

        println("Check", "Took: " + strconv.FormatInt((time.Now().Unix() - real_last_update), 10) + "s for " + strconv.Itoa(neww) + " updates, Pending download: " + strconv.Itoa(len(bot.tasks)))
		pause(config.MinTimeBetweenUpdates - (time.Now().Unix() - real_last_update))
	}
}

var baseNameSeparators = regexp.MustCompile(`[./]`)
var illegalName = regexp.MustCompile(`[^[:alnum:]-.]`)
var transliterations = map[rune]string{
	'À': "A",
	'Á': "A",
	'Â': "A",
	'Ã': "A",
	'Ä': "A",
	'Å': "AA",
	'Æ': "AE",
	'Ç': "C",
	'È': "E",
	'É': "E",
	'Ê': "E",
	'Ë': "E",
	'Ì': "I",
	'Í': "I",
	'Î': "I",
	'Ï': "I",
	'Ð': "D",
	'L': "L",
	'Ñ': "N",
	'Ò': "O",
	'Ó': "O",
	'Ô': "O",
	'Õ': "O",
	'Ö': "O",
	'Ø': "OE",
	'Ù': "U",
	'Ú': "U",
	'Ü': "U",
	'Û': "U",
	'Ý': "Y",
	'Þ': "Th",
	'ß': "ss",
	'à': "a",
	'á': "a",
	'â': "a",
	'ã': "a",
	'ä': "a",
	'å': "aa",
	'æ': "ae",
	'ç': "c",
	'è': "e",
	'é': "e",
	'ê': "e",
	'ë': "e",
	'ì': "i",
	'í': "i",
	'î': "i",
	'ï': "i",
	'ð': "d",
	'l': "l",
	'ñ': "n",
	'n': "n",
	'ò': "o",
	'ó': "o",
	'ô': "o",
	'õ': "o",
	'o': "o",
	'ö': "o",
	'ø': "oe",
	's': "s",
	'ù': "u",
	'ú': "u",
	'û': "u",
	'u': "u",
	'ü': "u",
	'ý': "y",
	'þ': "th",
	'ÿ': "y",
	'z': "z",
	'Œ': "OE",
	'œ': "oe",
}

var separators = regexp.MustCompile(`[ &_=+:]`)
var dashes = regexp.MustCompile(`[\-]+`)

func CleanName(s string) string {
	if len(s) > 200 {
		s = s[:200]
	}
	baseName := baseNameSeparators.ReplaceAllString(s, "-")
	baseName = cleanString(baseName, illegalName)
	return baseName
}

func cleanString(s string, r *regexp.Regexp) string {
	s = strings.Trim(s, " ")
	s = Accents(s)
	s = separators.ReplaceAllString(s, "-")
	s = r.ReplaceAllString(s, "")
	s = dashes.ReplaceAllString(s, "-")
	return s
}

func Accents(s string) string {
	b := bytes.NewBufferString("")
	for _, c := range s {
		if val, ok := transliterations[c]; ok {
			b.WriteString(val)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

