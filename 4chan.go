package main
import(
	"net/http"
	"encoding/json"
	// "strconv"
)
// https://github.com/4chan/4chan-API
type Pages []struct {
	Page int `json:"page"`
	Threads []struct {
		No int `json:"no"`
		LastModified int `json:"last_modified"`
	} `json:"threads"`
}
type Post struct {
	No int `json:"no"`
	Now string `json:"now"`
	Name string `json:"name"`
	Sub string `json:"sub"`
	Com string `json:"com"`
	Filename string `json:"filename,omitempty"`
	Ext string `json:"ext,omitempty"`
	W int `json:"w,omitempty"`
	H int `json:"h,omitempty"`
	TnW int `json:"tn_w,omitempty"`
	TnH int `json:"tn_h,omitempty"`
	Tim int `json:"tim,omitempty"`
	Time int `json:"time"`
	Md5 string `json:"md5,omitempty"`
	Fsize int `json:"fsize,omitempty"`
	Resto int `json:"resto"`
	Bumplimit int `json:"bumplimit,omitempty"`
	Imagelimit int `json:"imagelimit,omitempty"`
	SemanticURL string `json:"semantic_url,omitempty"`
	Replies int `json:"replies,omitempty"`
	Images int `json:"images,omitempty"`
	UniqueIps int `json:"unique_ips,omitempty"`
}
type Thread struct {
	Posts []Post `json:"posts"`
}
//
/*func (p*Post)GetFileUrl() string {
	return "https://i.4cdn.org/b/" + p.GetFilename()
}
func (p*Post)GetFilename() string {
	return strconv.Itoa(p.Tim) + p.Ext
}
func (p*Post)GetDownloadFilename() string {
	return strconv.Itoa(p.Tim) + "_" + CleanName(p.Filename) + p.Ext
}*/
func (p*Pages)Load() error{
	r, err := http.Get("https://a.4cdn.org/b/threads.json")
	if err != nil{
		return err
	}else{
		err = json.NewDecoder(r.Body).Decode(p)
		r.Body.Close()
		return err
	}
}
func (t*Thread)Load(threadNo string) error{
	r, err := http.Get("https://a.4cdn.org/b/thread/" + threadNo + ".json")
	if err != nil{
		return err
	}else{
		err = json.NewDecoder(r.Body).Decode(t)
		r.Body.Close()
		return err
	}
}