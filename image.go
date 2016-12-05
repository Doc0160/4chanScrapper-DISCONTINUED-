/* ========================================================================
   $File: $
   $Date: $
   $Revision: $
   $Creator: Tristan Magniez $
   ======================================================================== */

package main

import (
    "os"
    "errors"
    "io"
    "bytes"
    "log"
    "io/ioutil"
)

func isSameFile(f1 os.FileInfo, f2 os.FileInfo, Folder string) (bool, error) {
	if f1.Size()!=f2.Size(){
		return false, nil
	}

    a, err := os.Open("./"+Folder+"/"+f1.Name())
    if err != nil {
        return false, err
    }
    defer a.Close()
	b, err := os.Open("./"+Folder+"/"+f2.Name())
    if err != nil {
        return false, err
    }
    defer b.Close()
 
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
	return false, err
}

func checkDuplicates(t string, Folder string){
	f1, err := os.Stat(t)
    if err != nil {
        return
    }
	files, err := ioutil.ReadDir("./"+Folder)
    if err != nil {
        return
    }
	for _, f2 := range files{
		if f1.Name() != f2.Name() {
			r, err := isSameFile(f1, f2, Folder)
			if !r {
				if err != nil {
					log.Println("ERROR : " + err.Error())
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
