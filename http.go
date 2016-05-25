package main
import(
	"fmt"
	"strconv"
	"io/ioutil"
    "html/template"
	"github.com/gorilla/mux"
	"sort"
	"net/url"
	"net/http"
)
type Serv struct{
	Config Config
}
func http_serve(config *Config){
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
				l.URL="/"+url.QueryEscape(f.Name())+"/"
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
	for _,v := range config.Keywords{
		for _,v2 := range v{
			r.PathPrefix("/static/"+v2).Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))
		}
	}
	err = http.ListenAndServe(":"+strconv.Itoa(config.Port), r)
	fmt.Println(err)
}