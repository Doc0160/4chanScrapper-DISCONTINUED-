package main
/*import(
	"time"
	"encoding/json"
	"os"
	"net/http"
)
type Config struct {
	LastModified time.Time
	ParsedKeywords map[string]string
	//
	Log *bool `json:"log"`
	DownloadRetries uint8 `json:"download_retries"`
	MinTimeBetweenUpdates int64 `json:"min_time_between_updates"`
	Timeout int64 `json:"timeout"`
	Keywords map[string][]string `json:"keywords"`
}
func (config *Config) CheckConfig(file string) (bool, error) {
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
			Timeout : time.Duration(config.Timeout) * time.Second,
		}
		// TODO(doc): check if this is needed
		config.ParsedKeywords = nil
		config.ParsedKeywords = make(map[string]string)
		//
		for k,v := range config.Keywords{
			os.MkdirAll("./"+k+"/", 0777)
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
		if config.Log == nil {
			*config.Log = true
		}
		doLog = *config.Log
		config.LastModified = t.ModTime()
		return true, nil
	}
	return false, nil
}*/