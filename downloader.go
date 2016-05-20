package main
import(
	"strings"
	"net/http"
	"os"
	"fmt"
	"io"
	"path"
)
type DownloadTask struct{
	Filename string
	URL string
}
type Downloader struct{
	Queue chan DownloadTask
	// NOTE(doc): resusing same client take less time
	Client http.Client
	Loop bool
}
func (dl*Downloader)StartLoop(){
	dl.Loop=true
	for dl.Loop{
		file := <-dl.Queue
		// TODO(doc): delete that, check if ext!="" before putting it to dl
		if strings.Contains(file.Filename,"."){
			// TODO(doc): redo this in a better way:
			// create file path if it does'nt exist
			os.MkdirAll("./"+path.Dir(file.Filename), 0777)
			//
			t := "./"+file.Filename
			if !file_exist(t){
				out, err := os.Create(t)
				if err==nil{
					resp, err := dl.Client.Get(file.URL)
					if err==nil{
						fmt.Println("\""+path.Base(file.Filename)+"\" downloading")
						_, err := io.Copy(out, resp.Body)	
						if err==nil{
							out.Close()
							resp.Body.Close()
							fmt.Println("\""+path.Base(file.Filename)+"\" downloaded")
							go check_for_duplicates(t, path.Dir(file.Filename))
						}else{
							fmt.Println("\""+path.Base(file.Filename)+"\" download fail")
							println(err.Error())
							// NOTE(doc): the main loop will retry automaticly 
							//	if the filename is not found
							os.Remove(t)
						}
					}else{
						println("dl1 "+err.Error())
					}
				}else{
					fmt.Println(err)
					println(file.URL)
				}
			}
		}
	}
}