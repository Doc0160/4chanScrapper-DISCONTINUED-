/* ========================================================================
   $File: $
   $Date: $
   $Revision: $
   $Creator: Tristan Magniez $
   ======================================================================== */

package main

import(
	"fmt"
	"strconv"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"os"
	"io"
	"bytes"
	"errors"
	"sync"
	"regexp"
    "runtime/debug"
    "compress/gzip"
)

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

	switch {
	case erra == io.EOF && errb == io.EOF:
		a.Close()
		b.Close()
		return true, nil
	case erra != nil && errb != nil:
		return false, errors.New(erra.Error() + " " + errb.Error())
	case erra != nil:
		b.Close()
		return false, erra
	case errb != nil:
		a.Close()
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
		if f1.Name() != f2.Name() {
			r, err := isSameFile(f1, f2, Folder)
			if !r {
				if err != nil {
					println("ERROR : " + err.Error())
				}
			} else {
				if f1.ModTime().Before(f2.ModTime()) {
					println("Duplicate removed", f2.Name())
					os.Remove("./"+Folder+"/"+f1.Name())
				} else {
					println("Duplicate removed", f1.Name())
					os.Remove("./"+Folder+"/"+f2.Name())
				}
			}
		}
	}
}

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

func download_loop(dl chan dl_struct, Config * Config){
	var err error
	var file dl_struct
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
					println("ERROR : " + file.Filename + " " + err.Error())
				}
			}
		}
	}
}

func checkKeywords(config *Config, decoded_thread Thread)string{
	for v2, k := range config.ParsedKeywords {
		a := strings.ToUpper(v2)
		if strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Sub), a) || strings.Contains(strings.ToUpper(decoded_thread.Posts[0].Com), a){
			return k
		}
	}
	return ""
}

func doThread(wg *sync.WaitGroup, threadURL string, dl chan dl_struct, config Config, last_update int){
	defer wg.Done()
	decoded_thread := Thread{}
	err := decoded_thread.Load(threadURL)
	if err == nil || err == io.EOF {
		if decoded_thread.Posts != nil {
			folder := checkKeywords(&config, decoded_thread);
			if folder != "" {
				for _, post := range decoded_thread.Posts {
					if post.Time > last_update && post.Tim != 0 {
						dl<-dl_struct{
							Filename: CleanName(strconv.Itoa(post.Tim) + "_" + post.Filename) + post.Ext,
							Folder: folder,
							Url: "https://i.4cdn.org/b/" + strconv.Itoa(post.Tim) + post.Ext, 
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

func FullDupeCheck(c Config){
	for Folder, _ := range c.Keywords {
		files, _ := ioutil.ReadDir("./"+Folder)
		for _, f2 := range files{
			println(Folder, f2.Name())
			checkDuplicates("./"+Folder+"/"+f2.Name(), Folder)
		}
	}
}

func main(){
	debug.SetGCPercent(500)
	var config Config

	// download every picture sent in dl chan
	dl:=make(chan dl_struct,100)
	go download_loop(dl, &config)
	go download_loop(dl, &config)

    //
	var real_last_update int64 = 0
	var last_updates map[int]int = make(map[int]int)
	var decoded_pages Pages
	var wg sync.WaitGroup

    //
	for {
		// load config
		ok, err := config.CheckConfig("config.json")
		if ok {
			// go FullDupeCheck(config)
			println("New config")
			println("\tTimeout: " + strconv.FormatInt(config.Timeout, 10) + "s")
			println("\tMinimum time between updates: " + strconv.FormatInt(config.MinTimeBetweenUpdates, 10) + "s")
			println("\tDownload retries: " + strconv.Itoa(int(config.DownloadRetries)) + " times")
			print("\t")
			fmt.Println(config.Keywords)
		} else if err != nil {
			println("ERROR : " + err.Error())
		}
		real_last_update=time.Now().Unix()
		err = decoded_pages.Load()
		var neww = 0
		if err==io.EOF{
			println("ERROR : " + err.Error())
		}else if err!=nil{
			println("ERROR : " + err.Error())
		}else{

			for _,page := range decoded_pages{
				for _, thread := range page.Threads {
					if last_updates[thread.No]<thread.LastModified {
						wg.Add(1)
						go doThread(&wg, strconv.Itoa(thread.No), dl, config, last_updates[thread.No])
						last_updates[thread.No] = thread.LastModified
						neww++
					}
				}
			}
			wg.Wait()
        }

        println("Check", "Took: " + strconv.FormatInt((time.Now().Unix() - real_last_update), 10) + "s for " + strconv.Itoa(neww) + " updates, Pending download: " + strconv.Itoa(len(dl)))
		pause(config.MinTimeBetweenUpdates - (time.Now().Unix() - real_last_update))
	}
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

type dl_struct struct{
	Tries uint8
	Filename string
	Folder string
	Url string
	Size int
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
