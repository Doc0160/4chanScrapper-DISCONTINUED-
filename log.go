package main
/*import(
	"time"
	"os"
	"strconv"
)
var debugFilename = "out"
var debugFile *os.File = nil
var doLog = true
func log(a string, t string){
	f := "\x1b[32m[" + strconv.FormatInt(time.Now().UnixNano(), 10) + "] \x1b[33m[" + a + "] \x1b[36m" + t + "\x1b[0m\n"
	if doLog {
		logFile()
		if debugFile != nil {
			f2 := "[" + strconv.FormatInt(time.Now().UnixNano(), 10) + "] [" + a + "] " + t + "\n"
			debugFile.WriteString(f2)
		}
	}else{
		debugFile.Close()
		debugFile = nil
	}
	print(f)
}
func logError(t string){
	log("error", t)
}
func logFile(){
	if debugFile == nil {
		os.MkdirAll("./log/", 0777)
		debugFilename = "./log/" + strconv.FormatInt(time.Now().UnixNano(), 10)
		var err error
		debugFile, err = os.Create(debugFilename)
		if err != nil {
			logError(err.Error())
			debugFile = nil
		}
	}
}*/