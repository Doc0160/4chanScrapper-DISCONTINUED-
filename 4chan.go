/* ========================================================================
   $File: $
   $Date: $
   $Revision: $
   $Creator: Tristan Magniez $
   ======================================================================== */

package main

import (
    "net/http"
    "encoding/json"
    "strconv"
    "compress/gzip"
    "log"
)

var _ = log.Println

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

func (t*Thread)Load(threadNo string) error{
    client := &http.Client{}
    req, err := http.NewRequest("GET", "https://a.4cdn.org/b/thread/" + threadNo + ".json", nil)
    req.Header.Add("Accept-Encoding", "gzip")
    r, err := client.Do(req)
    if err != nil{
        return err
    }
    defer r.Body.Close()
    if len(r.Header["Content-Encoding"]) > 0 {
        if r.Header["Content-Encoding"][0] == "gzip" {
            g, err := gzip.NewReader(r.Body)
            if err != nil {
                return err
            }
            defer g.Close()
            err = json.NewDecoder(g).Decode(t)
            if err != nil {
                return  err
            }
        } else {
            panic(r.Header["Content-Encoding"])
        }
    }else {
        err = json.NewDecoder(r.Body).Decode(t)
        if err != nil {
            return err
        }
    }
    return nil
}

func (p*Pages)Load() error{
    client := &http.Client{}
    req, err := http.NewRequest("GET", "https://a.4cdn.org/b/threads.json", nil)
    req.Header.Add("Accept-Encoding", "gzip")
    r, err := client.Do(req)
    if err != nil{
        return err
    }
    defer r.Body.Close()
    if len(r.Header["Content-Encoding"]) > 0 {
        if r.Header["Content-Encoding"][0] == "gzip" {
            g, err := gzip.NewReader(r.Body)
            if err != nil {
                return err
            }
            defer g.Close()
            err = json.NewDecoder(g).Decode(p)
            if err != nil {
                return  err
            }
        } else {
            //println(r.Header["Content-Encoding"])
            panic(r.Header["Content-Encoding"])
        }
    }else {
        err = json.NewDecoder(r.Body).Decode(p)
        if err != nil {
            return err
        }
    }
    return nil
}

func (p*Post)GetFileUrl() string {
	return "https://i.4cdn.org/b/" + p.GetFilename()
}

func (p*Post)GetFilename() string {
	return strconv.Itoa(p.Tim) + p.Ext
}

func (p*Post)GetDownloadFilename() string {
	return strconv.Itoa(p.Tim) + "_" + CleanName(p.Filename) + p.Ext
}
